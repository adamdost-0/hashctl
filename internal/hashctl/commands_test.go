package hashctl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func testEnv(values map[string]string) envLookup {
	return func(key string) string {
		return values[key]
	}
}

func runTest(args []string, env map[string]string, handler http.HandlerFunc) (int, string, string) {
	server := httptest.NewServer(handler)
	defer server.Close()
	fullArgs := append([]string{"--api-url", server.URL + "/"}, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := New(&stdout, &stderr, testEnv(env), server.Client()).Run(fullArgs)
	return code, stdout.String(), stderr.String()
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func jobPayload(id string, status string) map[string]any {
	return map[string]any{
		"job_id":              id,
		"requestor_object_id": "requestor-1",
		"status":              status,
		"cloud_name":          "azure_government",
		"source_account":      "sthashsource",
		"source_container":    "input",
		"expected_blob_count": 3,
		"correlation_id":      "corr-1",
		"created_at":          "2026-05-18T00:00:00Z",
		"updated_at":          "2026-05-18T00:00:10Z",
		"metadata":            map[string]string{},
	}
}

func TestJobCreateSendsExpectedBodyAndHeaders(t *testing.T) {
	var body JobCreateRequest
	code, stdout, stderr := runTest(
		[]string{
			"--output", "json",
			"--correlation-id", "corr-1",
			"--local-principal-id", "user-1",
			"--local-groups", "group-a,group-b",
			"job", "create",
			"--source-account", "sthashsource",
			"--source-container", "input",
			"--blob-name", "raw/a.txt",
			"--blob-name", "raw/b.txt",
			"--batch-size", "1",
			"--cloud-name", "azure_government",
			"--metadata", "scenario=unit",
		},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/job/create" {
				t.Fatalf("path = %s", r.URL.Path)
			}
			if r.Header.Get("x-correlation-id") != "corr-1" {
				t.Fatalf("missing correlation header")
			}
			if r.Header.Get("x-hash-engine-local-groups") != "group-a,group-b" {
				t.Fatalf("groups header = %q", r.Header.Get("x-hash-engine-local-groups"))
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			writeJSONResponse(t, w, http.StatusAccepted, jobPayload("job-1", "hashing"))
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if body.SourceAccount != "sthashsource" || len(body.BlobNames) != 2 || body.Metadata["scenario"] != "unit" {
		t.Fatalf("unexpected body: %+v", body)
	}
	if !strings.Contains(stdout, `"job_id": "job-1"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestJobCreateAutoDiscoveryWithoutPrefix(t *testing.T) {
	var body JobCreateRequest
	code, _, stderr := runTest(
		[]string{
			"--output", "json",
			"job", "create",
			"--source-account", "sthashsource",
			"--source-container", "input",
		},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			writeJSONResponse(t, w, http.StatusAccepted, jobPayload("job-1", "hashing"))
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if body.SourceAccount != "sthashsource" || body.SourceContainer != "input" {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.Prefix != "" || len(body.BlobNames) != 0 {
		t.Fatalf("expected no prefix or blob_names, got: prefix=%q blobs=%v", body.Prefix, body.BlobNames)
	}
}

func TestHealthReadyUsesReadyRoute(t *testing.T) {
	code, stdout, stderr := runTest([]string{"health", "ready"}, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health/ready" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeJSONResponse(t, w, http.StatusOK, map[string]any{"status": "ready"})
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if strings.TrimSpace(stdout) != "ready" {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestConfigFileSuppliesAPIURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeJSONResponse(t, w, http.StatusOK, map[string]any{"status": "ok"})
	}))
	defer server.Close()
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"api_url": "`+server.URL+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := New(&stdout, &stderr, testEnv(nil), server.Client()).Run([]string{"--config", configPath, "health"})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestBearerTokenFileSetsAuthorizationHeader(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenPath, []byte("token-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runTest([]string{"--bearer-token-file", tokenPath, "health"}, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token-value" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		writeJSONResponse(t, w, http.StatusOK, map[string]any{"status": "ok"})
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
}

func TestJobCommandsUseDocumentedRoutes(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		path   string
		method string
	}{
		{name: "get", args: []string{"job", "get", "job-1"}, path: "/job/job-1", method: http.MethodGet},
		{name: "list", args: []string{"job", "list"}, path: "/job/list", method: http.MethodGet},
		{name: "cancel", args: []string{"job", "cancel", "job-1"}, path: "/job/job-1", method: http.MethodDelete},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, stderr := runTest(tc.args, nil, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Fatalf("%s %s", r.Method, r.URL.Path)
				}
				if tc.name == "list" {
					writeJSONResponse(t, w, http.StatusOK, []any{jobPayload("job-1", "queued")})
					return
				}
				writeJSONResponse(t, w, http.StatusOK, jobPayload("job-1", "cancelled"))
			})
			if code != ExitSuccess {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
		})
	}
}

func TestManifestGetRequiresOutputFileInHumanMode(t *testing.T) {
	code, _, stderr := runTest([]string{"manifest", "get", "job-1"}, nil, func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called")
	})
	if code != ExitUsage {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "--output-file is required in human mode") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestManifestGetUsageShowsFlagBeforeJobID(t *testing.T) {
	code, _, stderr := runTest(
		[]string{"manifest", "get", "--output-file", "file-a.xml", "job-1", "extra"},
		nil,
		func(_ http.ResponseWriter, _ *http.Request) {
			t.Fatal("handler should not be called")
		},
	)
	if code != ExitUsage {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "usage: hashctl manifest get [--output-file path] <job_id>") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestManifestGetWritesXMLWhenOutputFileSupplied(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "manifest.xml")
	code, stdout, stderr := runTest(
		[]string{"manifest", "get", "--output-file", outputPath, "job-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/job/job-1/manifest" {
				t.Fatalf("path = %s", r.URL.Path)
			}
			writeJSONResponse(t, w, http.StatusOK, ManifestResponse{JobID: "job-1", Status: "complete", ManifestXML: "<Root/>"})
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "wrote manifest job-1 to "+outputPath) {
		t.Fatalf("stdout = %q", stdout)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "<Root/>" {
		t.Fatalf("data = %q", string(data))
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func TestManifestGetJSONReturnsManifestPayloadWithoutOutputFile(t *testing.T) {
	code, stdout, stderr := runTest(
		[]string{"--output", "json", "manifest", "get", "job-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/job/job-1/manifest" {
				t.Fatalf("path = %s", r.URL.Path)
			}
			writeJSONResponse(t, w, http.StatusOK, ManifestResponse{JobID: "job-1", Status: "complete", ManifestXML: "<Root/>"})
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"manifest_xml": "\u003cRoot/\u003e"`) {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestManifestBuildSuppressesXMLInHumanSummary(t *testing.T) {
	code, stdout, stderr := runTest([]string{"manifest", "build", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/job/job-1/manifest" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		writeJSONResponse(t, w, http.StatusAccepted, ManifestResponse{JobID: "job-1", Status: "awaiting_first_signature", ManifestXML: "<Root>secret</Root>"})
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if strings.Contains(stdout, "<Root>") || !strings.Contains(stdout, "bytes=") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestSignFirstSendsJobIDInBody(t *testing.T) {
	var body SignRequest
	code, _, stderr := runTest([]string{"sign", "first", "--key-name", "dev-first", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sign/first-signature" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		writeJSONResponse(t, w, http.StatusOK, map[string]any{
			"job":       jobPayload("job-1", "awaiting_second_signature"),
			"signature": map[string]any{"job_id": "job-1", "signature_number": 1, "signer_object_id": "requestor-1", "key_id": "key-1"},
		})
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if body.JobID != "job-1" || body.KeyName != "dev-first" {
		t.Fatalf("body = %+v", body)
	}
}

func TestJobGetJSONIncludesTimestampFields(t *testing.T) {
	code, stdout, stderr := runTest([]string{"--output", "json", "job", "get", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusOK, jobPayload("job-1", "queued"))
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"created_at": "2026-05-18T00:00:00Z"`) ||
		!strings.Contains(stdout, `"updated_at": "2026-05-18T00:00:10Z"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestSignJSONIncludesSignerAndKeyNameFields(t *testing.T) {
	code, stdout, stderr := runTest([]string{"--output", "json", "sign", "first", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusOK, map[string]any{
			"job": jobPayload("job-1", "awaiting_second_signature"),
			"signature": map[string]any{
				"job_id":           "job-1",
				"signature_number": 1,
				"signer_object_id": "requestor-1",
				"signer_key_name":  "dev-first-signer",
				"key_id":           "key-1",
				"key_version":      "version-1",
				"key_name":         "dev-first-signer",
				"signed_at":        "2026-05-18T00:00:20Z",
				"policy_metadata": map[string]any{
					"provider_mode": "request_key_name",
				},
			},
		})
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"signer_key_name": "dev-first-signer"`) ||
		!strings.Contains(stdout, `"key_name": "dev-first-signer"`) ||
		!strings.Contains(stdout, `"signed_at": "2026-05-18T00:00:20Z"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestSignAcceptsFlagsAfterJobID(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{
			name: "first",
			args: []string{
				"sign", "first", "job-1",
				"--key-id", "kid-1",
				"--key-version", "version-1",
				"--key-name", "dev-first",
			},
			path: "/sign/first-signature",
		},
		{
			name: "second",
			args: []string{
				"sign", "second", "job-1",
				"--key-id", "kid-1",
				"--key-version", "version-1",
				"--key-name", "dev-second",
			},
			path: "/sign/second-signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body SignRequest
			code, _, stderr := runTest(tt.args, nil, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Fatalf("path = %s", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				writeJSONResponse(t, w, http.StatusOK, map[string]any{
					"job": jobPayload("job-1", "complete"),
					"signature": map[string]any{
						"job_id":           "job-1",
						"signature_number": 1,
						"signer_object_id": "requestor-1",
						"key_id":           "key-1",
					},
				})
			})
			if code != ExitSuccess {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			expectedKeyName := "dev-" + tt.name
			if body.JobID != "job-1" ||
				body.KeyID != "kid-1" ||
				body.KeyVersion != "version-1" ||
				body.KeyName != expectedKeyName {
				t.Fatalf("body = %+v", body)
			}
		})
	}
}

func TestSignRejectsExtraPositionalArgsAfterFlexibleParsing(t *testing.T) {
	code, _, stderr := runTest(
		[]string{"sign", "first", "job-1", "extra", "--key-name", "dev-first"},
		nil,
		func(_ http.ResponseWriter, _ *http.Request) {
			t.Fatal("handler should not be called")
		},
	)
	if code != ExitUsage {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "usage: hashctl sign first <job_id>") {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestAPIErrorUsesExitCodeAndJSONShape(t *testing.T) {
	code, _, stderr := runTest([]string{"--output", "json", "sign", "first", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusConflict, map[string]any{
			"detail": map[string]any{
				"error_code":     "signing_conflict",
				"message":        "Job is not ready for signing.",
				"job_id":         "job-1",
				"correlation_id": "corr-1",
			},
		})
	})
	if code != ExitAPIClient {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, `"error_code": "signing_conflict"`) || !strings.Contains(stderr, `"http_status": 409`) {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestJobGetForbiddenSurfacesAPIError(t *testing.T) {
	code, _, stderr := runTest([]string{"--output", "json", "job", "get", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusForbidden, map[string]any{
			"detail": map[string]any{
				"error_code":     "job_forbidden",
				"message":        "Only the job requestor can access this job.",
				"job_id":         "job-1",
				"correlation_id": "corr-1",
			},
		})
	})
	if code != ExitAPIClient {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, `"error_code": "job_forbidden"`) || !strings.Contains(stderr, `"http_status": 403`) {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestLiteralBearerTokenFlagIsRejected(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := New(&stdout, &stderr, testEnv(map[string]string{"HASH_ENGINE_API": "http://127.0.0.1:1"}), http.DefaultClient).Run(
		[]string{"--bearer-token", "secret", "health"},
	)
	if code != ExitUsage {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(stderr.String(), "HASH_ENGINE_BEARER_TOKEN") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestSmokeMultiJobSubmitsFiveJobsAndDoesNotSign(t *testing.T) {
	var posts int
	var signs int
	var mu sync.Mutex
	correlations := map[string]bool{}
	code, stdout, stderr := runTest(
		[]string{"--poll-interval", "1ms", "--poll-timeout", "2s", "--output", "json", "smoke", "multi-job", "--source-account", "acct", "--source-container", "input", "--run-id", "run-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			if strings.HasPrefix(r.URL.Path, "/sign/") {
				signs++
			}
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/job/create":
				posts++
				correlation := r.Header.Get("x-correlation-id")
				if !strings.HasPrefix(correlation, "run-1-job-0") {
					t.Fatalf("unexpected correlation = %s", correlation)
				}
				correlations[correlation] = true
				writeJSONResponse(t, w, http.StatusAccepted, jobPayload("job-"+strings.TrimPrefix(correlation, "run-1-job-0"), "hashing"))
			case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/job/job-"):
				id := strings.TrimPrefix(r.URL.Path, "/job/")
				writeJSONResponse(t, w, http.StatusOK, jobPayload(id, "awaiting_first_signature"))
			default:
				t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
			}
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if posts != 5 || signs != 0 {
		t.Fatalf("posts=%d signs=%d", posts, signs)
	}
	if len(correlations) != 5 {
		t.Fatalf("correlations=%v", correlations)
	}
	if !strings.Contains(stdout, `"correlation_id": "run-1-job-05"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func runLocal(args []string, env map[string]string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := New(&stdout, &stderr, testEnv(env), nil).Run(args)
	return code, stdout.String(), stderr.String()
}

func TestTopLevelHelpCommandsDoNotRequireAPIConfig(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "no args", args: nil},
		{name: "--help", args: []string{"--help"}},
		{name: "-h", args: []string{"-h"}},
		{name: "help", args: []string{"help"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := runLocal(tc.args, nil)
			if code != ExitSuccess {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q", stderr)
			}
			if !strings.Contains(stdout, "Hash Engine command-line client") ||
				!strings.Contains(stdout, "Usage:") ||
				!strings.Contains(stdout, "hashctl [global flags] <command> [command flags]") {
				t.Fatalf("stdout = %q", stdout)
			}
		})
	}
}

func TestCommandHelpDoesNotRequireAPIConfig(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		contains string
	}{
		{name: "help health", args: []string{"help", "health"}, contains: "hashctl health [ready]"},
		{name: "health --help", args: []string{"health", "--help"}, contains: "hashctl health [ready]"},
		{name: "job --help", args: []string{"job", "--help"}, contains: "hashctl job <subcommand> [flags]"},
		{name: "manifest --help", args: []string{"manifest", "--help"}, contains: "hashctl manifest <subcommand> [flags]"},
		{name: "sign --help", args: []string{"sign", "--help"}, contains: "hashctl sign <first|second> <job_id> [flags]"},
		{name: "smoke --help", args: []string{"smoke", "--help"}, contains: "hashctl smoke <single-job|multi-job> [flags]"},
		{name: "job create --help", args: []string{"job", "create", "--help"}, contains: "hashctl job create --source-account NAME --source-container NAME [flags]"},
		{name: "manifest get --help", args: []string{"manifest", "get", "--help"}, contains: "hashctl manifest get [--output-file PATH] <job_id>"},
		{name: "sign first --help", args: []string{"sign", "first", "--help"}, contains: "hashctl sign first <job_id>"},
		{name: "smoke single-job --help", args: []string{"smoke", "single-job", "--help"}, contains: "hashctl smoke single-job --source-account NAME --source-container NAME --prefix PREFIX [flags]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := runLocal(tc.args, nil)
			if code != ExitSuccess {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if !strings.Contains(stdout, tc.contains) {
				t.Fatalf("stdout = %q", stdout)
			}
		})
	}
}

func TestVersionCommandsDoNotRequireAPIConfig(t *testing.T) {
	originalVersion := Version
	Version = "dev"
	t.Cleanup(func() {
		Version = originalVersion
	})
	cases := [][]string{
		{"version"},
		{"--version"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			code, stdout, stderr := runLocal(args, nil)
			if code != ExitSuccess {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if strings.TrimSpace(stdout) != "hashctl dev" {
				t.Fatalf("stdout = %q", stdout)
			}
		})
	}
}

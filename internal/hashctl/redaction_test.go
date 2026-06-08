package hashctl

import (
	"bytes"
	"encoding/json"
	xmlpkg "encoding/xml"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Obviously-fake credential fixtures planted into API responses to prove that
// every success-output path scrubs secrets (issue #3, ADR-0003). None of these
// are real: the base64 fragments decode to "fakesig"/"fakekey" and the JWT
// payload is {"sub":"fake"}.
const (
	fakeSASURL     = "https://acct.blob.core.windows.net/c/b?sig=ZmFrZXNpZw%3D%3D&se=2030-01-01"
	fakeAccountKey = "AccountKey=ZmFrZWtleQ=="
	fakeJWT        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmYWtlIn0.c2lnbmF0dXJlMTIzNDU2"
)

// redactionLeakMarkers are distinctive secret substrings that must never survive
// redaction in any user-facing output.
var redactionLeakMarkers = []string{"ZmFrZXNpZw", "ZmFrZWtleQ", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"}

func assertNoLeak(t *testing.T, where string, output string) {
	t.Helper()
	for _, marker := range redactionLeakMarkers {
		if strings.Contains(output, marker) {
			t.Fatalf("%s leaked secret %q:\n%s", where, marker, output)
		}
	}
	if !strings.Contains(output, "[REDACTED") {
		t.Fatalf("%s did not contain any redaction marker:\n%s", where, output)
	}
}

func secretJobPayload() map[string]any {
	return map[string]any{
		"job_id":              "job-1",
		"requestor_object_id": "requestor-1",
		"status":              "failed",
		"cloud_name":          "azure_government",
		"expected_blob_count": 1,
		"correlation_id":      "corr-1",
		"created_at":          "2026-05-18T00:00:00Z",
		"updated_at":          "2026-05-18T00:00:10Z",
		"output_path":         fakeSASURL,
		"failure_reason":      fakeAccountKey + " " + fakeJWT,
		"metadata":            map[string]string{"sas": fakeSASURL, "jwt": fakeJWT, "key": fakeAccountKey},
	}
}

func TestHumanSuccessOutputRedactsSecrets(t *testing.T) {
	code, stdout, stderr := runTest([]string{"job", "get", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusOK, secretJobPayload())
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "job job-1: status=failed") {
		t.Fatalf("human summary missing expected prefix:\n%s", stdout)
	}
	assertNoLeak(t, "human success output", stdout)
}

func TestJSONSuccessOutputRedactsSecretsAndStaysValidJSON(t *testing.T) {
	code, stdout, stderr := runTest([]string{"--output", "json", "job", "get", "job-1"}, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(t, w, http.StatusOK, secretJobPayload())
	})
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	assertNoLeak(t, "json success output", stdout)
	// The structure-preserving redactor must never corrupt the document: a naive
	// whole-string redact() eats the closing quote after sig=/AccountKey= values.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("redacted JSON success output is not valid JSON: %v\n%s", err, stdout)
	}
	if parsed["job_id"] != "job-1" {
		t.Fatalf("redaction altered JSON structure: job_id=%v", parsed["job_id"])
	}
}

func TestManifestIncludeXMLJSONRedactsSecrets(t *testing.T) {
	xml := "<Manifest><Blob href=\"" + fakeSASURL + "\"/><Cred>" + fakeAccountKey + "</Cred><Token>" + fakeJWT + "</Token></Manifest>"
	code, stdout, stderr := runTest(
		[]string{"--output", "json", "manifest", "get", "--include-manifest-xml", "job-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, ManifestResponse{JobID: "job-1", Status: "complete", ManifestXML: xml})
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	assertNoLeak(t, "manifest --include-manifest-xml json output", stdout)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("redacted manifest JSON is not valid JSON: %v\n%s", err, stdout)
	}
}

func TestManifestOutputFileContentsAreRedacted(t *testing.T) {
	// Well-formed XML manifest (the SAS '&' is XML-escaped as &amp;). Redaction
	// must replace the secret values without breaking well-formedness.
	manifestXML := "<Manifest><Blob href=\"https://acct.blob.core.windows.net/c/b?sig=ZmFrZXNpZw%3D%3D&amp;se=2030-01-01\"/><Cred>" + fakeAccountKey + "</Cred><Token>" + fakeJWT + "</Token></Manifest>"
	outputPath := filepath.Join(t.TempDir(), "manifest.xml")
	code, _, stderr := runTest(
		[]string{"manifest", "get", "--output-file", outputPath, "job-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, ManifestResponse{JobID: "job-1", Status: "complete", ManifestXML: manifestXML})
		},
	)
	if code != ExitSuccess {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	assertNoLeak(t, "manifest --output-file contents", string(data))

	// The redacted manifest must remain well-formed XML. A naive whole-string
	// redact() consumes the closing quote/angle bracket of sig=/AccountKey=
	// values inside attributes and tags, corrupting the document (issue #3
	// cross-vendor review). Verify the saved file still parses.
	decoder := xmlpkg.NewDecoder(bytes.NewReader(data))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("redacted manifest --output-file is not well-formed XML: %v\n%s", err, data)
		}
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func TestJSONErrorFieldsAreRedacted(t *testing.T) {
	code, _, stderr := runTest(
		[]string{"--output", "json", "job", "get", "job-1"},
		nil,
		func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusForbidden, map[string]any{
				"detail": map[string]any{
					"error_code":     fakeAccountKey,
					"message":        "denied " + fakeJWT,
					"job_id":         "job-1",
					"correlation_id": fakeSASURL,
				},
			})
		},
	)
	if code != ExitAPIClient {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	assertNoLeak(t, "json error output", stderr)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stderr), &parsed); err != nil {
		t.Fatalf("redacted JSON error output is not valid JSON: %v\n%s", err, stderr)
	}
}

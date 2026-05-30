package hashctl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestApplyHeadersRejectsBearerTokenOverNonLoopbackHTTP(t *testing.T) {
	client, err := NewClient(
		Config{
			APIURL:      "http://example.com",
			BearerToken: "token-value",
			Timeout:     5 * time.Second,
		},
		http.DefaultClient,
	)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = client.applyHeaders(req)
	if err == nil || !strings.Contains(err.Error(), "refusing to send bearer token over non-loopback plaintext http") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseAPIErrorSuppressesServerDetailMessage(t *testing.T) {
	errOut := parseAPIError(
		http.StatusForbidden,
		"GET /job/job-1",
		[]byte(`{"detail":{"error_code":"job_forbidden","message":"token=secret-value","job_id":"job-1","correlation_id":"corr-1"}}`),
	)
	if errOut.ErrorCode != "job_forbidden" || errOut.HTTPStatus != http.StatusForbidden {
		t.Fatalf("unexpected error payload: %+v", errOut)
	}
	if strings.Contains(errOut.Message, "secret-value") {
		t.Fatalf("message leaked sensitive detail: %q", errOut.Message)
	}
	if !strings.Contains(errOut.Error(), "status 403") || !strings.Contains(errOut.Error(), "job_forbidden") {
		t.Fatalf("error string missing status/error_code: %q", errOut.Error())
	}
}

func TestLoopbackHTTPWithBearerTokenIsAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token-value" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(
		Config{
			APIURL:      server.URL,
			BearerToken: "token-value",
			Timeout:     5 * time.Second,
		},
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := client.do(context.Background(), http.MethodGet, "/health", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

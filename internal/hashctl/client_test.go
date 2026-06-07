package hashctl

import (
	"context"
	"errors"
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

// TestWaitForJobSurfaces5xxOnDeadline verifies that persistent 5xx responses surface as
// APIError (exit 5) rather than being masked as PollError (exit 6) when the deadline
// expires. This covers ADR-0002 exit-code semantics.
func TestWaitForJobSurfaces5xxOnDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail":{"error_code":"internal_error"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIURL:       server.URL,
		Timeout:      5 * time.Second,
		PollInterval: 1 * time.Millisecond,
		PollTimeout:  20 * time.Millisecond,
	}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.WaitForJob(context.Background(), "job-1", "complete")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", apiErr.HTTPStatus)
	}
}

// TestWaitForJobPollTimeoutWhenServerRespondsOK verifies that a job stuck in a
// non-terminal state still returns PollError (exit 6) when the deadline expires
// and the last poll succeeded.
func TestWaitForJobPollTimeoutWhenServerRespondsOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"job_id":"job-1","status":"hashing","requestor_object_id":"r","cloud_name":"az","expected_blob_count":0,"created_at":"","updated_at":""}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIURL:       server.URL,
		Timeout:      5 * time.Second,
		PollInterval: 1 * time.Millisecond,
		PollTimeout:  20 * time.Millisecond,
	}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.WaitForJob(context.Background(), "job-1", "complete")
	if err == nil {
		t.Fatal("expected error")
	}
	var pollErr PollError
	if !errors.As(err, &pollErr) {
		t.Fatalf("expected PollError (exit 6), got %T: %v", err, err)
	}
	if pollErr.Terminal {
		t.Fatal("expected non-terminal poll timeout")
	}
}

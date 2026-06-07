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

// TestHTTPSToHTTPRedirectRefusedBearerTokenNotLeaked verifies that a 3xx redirect
// from an HTTPS server to a plaintext HTTP endpoint is refused by secureCheckRedirect,
// and that the "attacker" HTTP endpoint never receives the Authorization header.
func TestHTTPSToHTTPRedirectRefusedBearerTokenNotLeaked(t *testing.T) {
	// Attacker server: plain HTTP.  Records any Authorization header it sees.
	var attackerGotAuth string
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerGotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	// TLS server that issues a 302 redirect to the plain-HTTP attacker.
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+"/captured", http.StatusFound)
	}))
	defer tlsServer.Close()

	// Build the Client with the TLS server's transport (so the self-signed cert
	// is trusted) and a bearer token that must never reach the attacker.
	client, err := NewClient(
		Config{
			APIURL:      tlsServer.URL,
			BearerToken: "secret-bearer-token",
			Timeout:     5 * time.Second,
		},
		tlsServer.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// The call must return an error — the redirect must be refused.
	err = client.do(context.Background(), http.MethodGet, "/health", nil, nil)
	if err == nil {
		t.Fatal("expected error: redirect from https to http should be refused, got nil")
	}

	// The attacker endpoint must never have received the Authorization header.
	if attackerGotAuth != "" {
		t.Fatalf("bearer token leaked to attacker: Authorization = %q", attackerGotAuth)
	}
}

// TestLoopbackHTTPRedirectAllowed verifies that a same-scheme redirect within a
// loopback HTTP server (no downgrade) is still followed normally.
func TestLoopbackHTTPRedirectAllowed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/original", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirected", http.StatusFound)
	})
	mux.HandleFunc("/redirected", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(
		Config{
			APIURL:      server.URL,
			BearerToken: "token",
			Timeout:     5 * time.Second,
		},
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := client.do(context.Background(), http.MethodGet, "/original", nil, &out); err != nil {
		t.Fatalf("same-scheme loopback redirect should be allowed, got: %v", err)
	}
}

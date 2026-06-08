package hashctl

import (
	"fmt"
	"testing"
)

// fakeResolver returns a resolver that always returns the given addresses.
func fakeResolver(addrs ...string) func(string) ([]string, error) {
	return func(_ string) ([]string, error) { return addrs, nil }
}

func TestResolveIsLoopbackIPLiteral127(t *testing.T) {
	if !resolveIsLoopback("127.0.0.1", nil) {
		t.Fatal("expected 127.0.0.1 to be loopback")
	}
}

func TestResolveIsLoopbackIPLiteralIPv6(t *testing.T) {
	if !resolveIsLoopback("::1", nil) {
		t.Fatal("expected ::1 to be loopback")
	}
}

func TestResolveIsLoopbackLocalhostAllLoopback(t *testing.T) {
	if !resolveIsLoopback("localhost", fakeResolver("127.0.0.1")) {
		t.Fatal("expected localhost resolving to 127.0.0.1 to be accepted")
	}
}

func TestResolveIsLoopbackLocalhostMultipleAllLoopback(t *testing.T) {
	if !resolveIsLoopback("localhost", fakeResolver("127.0.0.1", "::1")) {
		t.Fatal("expected localhost resolving to 127.0.0.1 and ::1 to be accepted")
	}
}

func TestResolveIsLoopbackLocalhostNonLoopbackRejected(t *testing.T) {
	if resolveIsLoopback("localhost", fakeResolver("8.8.8.8")) {
		t.Fatal("expected localhost resolving to 8.8.8.8 to be rejected")
	}
}

func TestResolveIsLoopbackLocalhostMixedAddressesRejected(t *testing.T) {
	// ANY non-loopback address is sufficient to reject.
	if resolveIsLoopback("localhost", fakeResolver("127.0.0.1", "1.2.3.4")) {
		t.Fatal("expected mixed addresses (loopback+non-loopback) to be rejected")
	}
}

func TestResolveIsLoopbackLocalhostResolutionFails(t *testing.T) {
	fail := func(_ string) ([]string, error) { return nil, fmt.Errorf("no route to host") }
	if resolveIsLoopback("localhost", fail) {
		t.Fatal("expected failed DNS resolution to be rejected (fail-closed)")
	}
}

func TestResolveIsLoopbackLocalhostEmptyAddressesRejected(t *testing.T) {
	if resolveIsLoopback("localhost", fakeResolver()) {
		t.Fatal("expected empty address list to be rejected")
	}
}

func TestResolveIsLoopbackNonLoopbackNameRejected(t *testing.T) {
	// Non-"localhost" hostnames always fail regardless of what the resolver returns.
	if resolveIsLoopback("example.com", fakeResolver("127.0.0.1")) {
		t.Fatal("expected example.com to be rejected even with loopback resolver result")
	}
}

func TestResolveIsLoopbackLocalhostCaseInsensitive(t *testing.T) {
	if !resolveIsLoopback("LOCALHOST", fakeResolver("127.0.0.1")) {
		t.Fatal("expected LOCALHOST (uppercase) to match")
	}
}

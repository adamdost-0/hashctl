package hashctl

import (
	"strings"
	"testing"
)

func TestRedactScrubsSensitiveValues(t *testing.T) {
	input := strings.Join([]string{
		"Authorization: Bearer abcdef1234567890",
		"url=https://acct.blob.core.usgovcloudapi.net/c/input.txt?sv=2024-01-01&sig=abcdef",
		"connection=DefaultEndpointsProtocol=https;AccountName=acct;AccountKey=abc123SECRET==;EndpointSuffix=core.usgovcloudapi.net",
		"token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyIn0.signature",
		"secret=AbCdEf0123456789AbCdEf0123456789AbCdEf01",
	}, " ")

	redacted := redact(input)
	if strings.Contains(redacted, "abcdef1234567890") ||
		strings.Contains(redacted, "sig=abcdef") ||
		strings.Contains(redacted, "AccountKey=abc123SECRET==") ||
		strings.Contains(redacted, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") ||
		strings.Contains(redacted, "AbCdEf0123456789AbCdEf0123456789AbCdEf01") {
		t.Fatalf("redacted output leaked secret: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] marker in %q", redacted)
	}
}

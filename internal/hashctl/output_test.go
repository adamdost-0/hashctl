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

func TestRedactHighEntropyHeuristics(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		needle      string
		shouldScrub bool
	}{
		{
			name:        "padded base64 token",
			input:       `value "RkFLRV9TRUNSRVRfVE9LRU5fSVNTVUVfRk9VUl9YWA==" , done`,
			needle:      "RkFLRV9TRUNSRVRfVE9LRU5fSVNTVUVfRk9VUl9YWA==",
			shouldScrub: true,
		},
		{
			name:        "32 char hex token",
			input:       "token 0123456789abcdef0123456789abcdef;",
			needle:      "0123456789abcdef0123456789abcdef",
			shouldScrub: true,
		},
		{
			name:        "36 char base62 token",
			input:       "token FakeBase62Token0123456789abcdefXYZ123",
			needle:      "FakeBase62Token0123456789abcdefXYZ123",
			shouldScrub: true,
		},
		{
			name:        "dot delimited opaque token",
			input:       "id fake.segment.token.0123456789abcdef.ABCDEF",
			needle:      "fake.segment.token.0123456789abcdef.ABCDEF",
			shouldScrub: true,
		},
		{
			name:        "secret followed by closing paren",
			input:       "failed: token AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00(invalid)",
			needle:      "AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00",
			shouldScrub: true,
		},
		{
			name:        "secret in pipe delimited log",
			input:       "SEVERITY|AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00|context",
			needle:      "AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00",
			shouldScrub: true,
		},
		{
			name:        "secret in xml tag value",
			input:       "<SignatureValue>AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00</SignatureValue>",
			needle:      "AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00",
			shouldScrub: true,
		},
		{
			name:        "secret in parentheses at start",
			input:       "error(AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00)",
			needle:      "AAAAAAAABBBBBBBBCCCCCCCCDDDDDDDD00",
			shouldScrub: true,
		},
		{
			name:        "normal sentence",
			input:       "this is a normal sentence about hashctl output handling",
			needle:      "[REDACTED_SECRET]",
			shouldScrub: false,
		},
		{
			name:        "uuid",
			input:       "id 123e4567-e89b-12d3-a456-426614174000",
			needle:      "[REDACTED_SECRET]",
			shouldScrub: false,
		},
		{
			name:        "short hex",
			input:       "checksum deadbeef",
			needle:      "[REDACTED_SECRET]",
			shouldScrub: false,
		},
		{
			name:        "job id and file path",
			input:       "job=job-20260607-abcdef12 path=/var/lib/hashctl/jobs/job-20260607-abcdef12/manifest.xml",
			needle:      "[REDACTED_SECRET]",
			shouldScrub: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := redact(tc.input)
			if tc.shouldScrub {
				if strings.Contains(got, tc.needle) {
					t.Fatalf("expected %q to be redacted, got: %q", tc.needle, got)
				}
				if !strings.Contains(got, "[REDACTED_SECRET]") {
					t.Fatalf("expected [REDACTED_SECRET] marker in %q", got)
				}
				return
			}
			if strings.Contains(got, tc.needle) != strings.Contains(tc.input, tc.needle) {
				t.Fatalf("unexpected transformation: input=%q output=%q", tc.input, got)
			}
			if got != tc.input {
				t.Fatalf("expected non-sensitive input unchanged, got: %q", got)
			}
		})
	}
}

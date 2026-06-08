package hashctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// writeJSON is the single choke point for all JSON output. It serializes value,
// scrubs every string literal through redact(), and writes the result in one
// call. Routing all JSON (success results and error payloads) through here makes
// credential redaction impossible to forget when a new output path is added
// (ADR-0003).
func writeJSON(w io.Writer, value any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return err
	}
	_, err := w.Write(redactJSONBytes(buf.Bytes()))
	return err
}

// redactJSONBytes rewrites a serialized JSON document, replacing the contents of
// every string literal with its redact()-scrubbed form. Each literal is decoded
// to plaintext, redacted, then re-encoded, so structure, key order, numbers, and
// escaping are preserved and the output is always valid JSON. This is required
// because applying redact() to the whole serialized document lets greedy patterns
// (Azure SAS query parameters, AccountKey=, ...) consume the closing quote of a
// value and corrupt the JSON.
func redactJSONBytes(data []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(data))
	for i := 0; i < len(data); {
		if data[i] != '"' {
			out.WriteByte(data[i])
			i++
			continue
		}
		j := i + 1
		for j < len(data) {
			if data[j] == '\\' {
				j += 2
				continue
			}
			if data[j] == '"' {
				break
			}
			j++
		}
		if j >= len(data) {
			out.Write(data[i:])
			break
		}
		literal := data[i : j+1]
		var decoded string
		if err := json.Unmarshal(literal, &decoded); err != nil {
			out.Write(literal)
		} else if scrubbed, err := json.Marshal(redact(decoded)); err != nil {
			out.Write(literal)
		} else {
			out.Write(scrubbed)
		}
		i = j + 1
	}
	return out.Bytes()
}

// writeRedactedText is the single choke point for free-form text output. It
// scrubs the complete rendered string through redact() and writes it in one
// call, so a secret can never be split across multiple writes (which could let a
// token straddle a chunk boundary and escape redaction).
func writeRedactedText(w io.Writer, text string) error {
	_, err := io.WriteString(w, redact(text))
	return err
}

func writeJobSummary(w io.Writer, job JobRecord) error {
	correlationID := ""
	if job.CorrelationID != nil {
		correlationID = *job.CorrelationID
	}
	_, err := fmt.Fprintf(
		w,
		"job %s: status=%s expected_blobs=%d correlation_id=%s\n",
		job.JobID,
		job.Status,
		job.ExpectedBlobCount,
		correlationID,
	)
	if err != nil {
		return err
	}
	if job.OutputPath != nil {
		_, err = fmt.Fprintf(w, "manifest_path=%s\n", *job.OutputPath)
	}
	if job.FailureReason != nil {
		_, err = fmt.Fprintf(w, "failure_reason=%s\n", *job.FailureReason)
	}
	return err
}

func writeJobListSummary(w io.Writer, jobs []JobRecord) error {
	for _, job := range jobs {
		if err := writeJobSummary(w, job); err != nil {
			return err
		}
	}
	if len(jobs) == 0 {
		_, err := fmt.Fprintln(w, "no jobs")
		return err
	}
	return nil
}

func writeManifestSummary(w io.Writer, manifest ManifestResponse) error {
	_, err := fmt.Fprintf(w, "manifest %s: status=%s bytes=%d\n", manifest.JobID, manifest.Status, len(manifest.ManifestXML))
	return err
}

func writeSignatureSummary(w io.Writer, response SignatureResponse) error {
	signer := ""
	if response.Signature.SignerObjectID != nil {
		signer = *response.Signature.SignerObjectID
	}
	_, err := fmt.Fprintf(
		w,
		"job %s: status=%s signature=%d signer=%s\n",
		response.Job.JobID,
		response.Job.Status,
		response.Signature.SignatureNumber,
		signer,
	)
	return err
}

func writeSmokeSummary(w io.Writer, result SmokeResult) error {
	if _, err := fmt.Fprintf(w, "smoke target_state=%s jobs=%d\n", result.TargetState, len(result.Jobs)); err != nil {
		return err
	}
	for _, job := range result.Jobs {
		if _, err := fmt.Fprintf(
			w,
			"%s prefix=%s correlation_id=%s final_state=%s\n",
			job.JobID,
			job.Prefix,
			job.CorrelationID,
			job.FinalState,
		); err != nil {
			return err
		}
	}
	return nil
}

func errorExitCode(err error) int {
	switch e := err.(type) {
	case APIError:
		if e.HTTPStatus >= 500 {
			return ExitAPIServer
		}
		return ExitAPIClient
	case TransportError:
		return ExitTransport
	case PollError:
		return ExitPollTimeout
	default:
		return ExitUsage
	}
}

func writeError(w io.Writer, output string, err error) {
	if output == "json" {
		payload := map[string]any{"error": map[string]any{"message": redact(err.Error())}}
		switch e := err.(type) {
		case APIError:
			payload["error"] = map[string]any{
				"http_status":    e.HTTPStatus,
				"error_code":     redact(e.ErrorCode),
				"message":        redact(e.Message),
				"route":          redact(e.Route),
				"job_id":         redact(e.JobID),
				"correlation_id": redact(e.CorrelationID),
			}
		case TransportError:
			payload["error"] = map[string]any{"message": redact(e.Err.Error()), "route": redact(e.Route)}
		case PollError:
			payload["error"] = map[string]any{
				"message":      redact(e.Error()),
				"job_id":       redact(e.Job.JobID),
				"target_state": redact(e.TargetState),
				"final_state":  redact(e.Job.Status),
			}
		}
		_ = writeJSON(w, payload)
		return
	}
	_, _ = fmt.Fprintf(w, "error: %s\n", redact(err.Error()))
}

func redact(value string) string {
	if value == "" {
		return value
	}
	redacted := value
	for _, pattern := range redactionPatterns {
		redacted = pattern.ReplaceAllString(redacted, "$1[REDACTED]")
	}
	redacted = querySecretPattern.ReplaceAllStringFunc(redacted, func(match string) string {
		parts := strings.SplitN(match, "=", 2)
		return parts[0] + "=[REDACTED]"
	})
	redacted = jwtPattern.ReplaceAllString(redacted, "[REDACTED_JWT]")
	redacted = highEntropyPattern.ReplaceAllStringFunc(redacted, func(token string) string {
		if looksSensitiveToken(token) {
			return "[REDACTED_SECRET]"
		}
		return token
	})
	return redacted
}

var (
	redactionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(authorization\s*[:=]\s*bearer\s+)[^\s",;]+`),
		regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9\-._~+/]+=*`),
		regexp.MustCompile(`(?i)(sharedaccesssignature=)[^;\s"'<>]+`),
		regexp.MustCompile(`(?i)(accountkey=)[^;\s"'<>]+`),
		regexp.MustCompile(`(?i)((?:clientsecret|password|pwd|secret)\s*[:=]\s*)[^;\s",'<>]+`),
	}
	querySecretPattern = regexp.MustCompile(`(?i)[?&](sig|signature|token|access_token|se|sp|sv|spr|sr|skoid|sktid|skt|ske|sks|skv)=[^&\s"'<>]+`)
	jwtPattern         = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`)
	highEntropyPattern = regexp.MustCompile(`\b[A-Za-z0-9+/_=-]{32,}\b`)
)

func looksSensitiveToken(value string) bool {
	var hasUpper bool
	var hasLower bool
	var hasDigit bool
	var hasSpecial bool
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case strings.ContainsRune("+/_-=", r):
			hasSpecial = true
		default:
			return false
		}
	}
	if len(value) >= 48 && hasLower && hasDigit {
		return true
	}
	return len(value) >= 40 && hasUpper && hasLower && hasDigit && hasSpecial
}

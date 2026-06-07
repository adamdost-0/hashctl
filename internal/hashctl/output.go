package hashctl

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

func writeJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
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
				"error_code":     e.ErrorCode,
				"message":        redact(e.Message),
				"route":          e.Route,
				"job_id":         e.JobID,
				"correlation_id": e.CorrelationID,
			}
		case TransportError:
			payload["error"] = map[string]any{"message": redact(e.Err.Error()), "route": e.Route}
		case PollError:
			payload["error"] = map[string]any{
				"message":      redact(e.Error()),
				"job_id":       e.Job.JobID,
				"target_state": e.TargetState,
				"final_state":  e.Job.Status,
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
	redacted = redactHighEntropyTokens(redacted)
	return redacted
}

var (
	redactionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(authorization\s*[:=]\s*bearer\s+)[^\s",;]+`),
		regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9\-._~+/]+=*`),
		regexp.MustCompile(`(?i)(sharedaccesssignature=)[^;\s]+`),
		regexp.MustCompile(`(?i)(accountkey=)[^;\s]+`),
		regexp.MustCompile(`(?i)((?:clientsecret|password|pwd|secret)\s*[:=]\s*)[^;\s",]+`),
	}
	querySecretPattern = regexp.MustCompile(`(?i)[?&](sig|signature|token|access_token|se|sp|sv|spr|sr|skoid|sktid|skt|ske|sks|skv)=[^&\s]+`)
	jwtPattern         = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`)
	highEntropyPattern = regexp.MustCompile("(^|[\\s\"'`,;])([A-Za-z0-9+/_=.:-]{32,})")
	uuidPattern        = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func looksSensitiveToken(value string) bool {
	if len(value) < 32 || uuidPattern.MatchString(value) || looksPathLike(value) {
		return false
	}
	if strings.Contains(value, "=") && strings.Contains(value, "/") && strings.Contains(value, ".") {
		return false
	}
	var hasUpper bool
	var hasLower bool
	var hasDigit bool
	var hasSpecial bool
	var hexOnly = true
	var coreLen int
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
			coreLen++
			hexOnly = hexOnly && ((r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f'))
		case r >= 'a' && r <= 'z':
			hasLower = true
			coreLen++
			hexOnly = hexOnly && ((r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f'))
		case r >= '0' && r <= '9':
			hasDigit = true
			coreLen++
		case strings.ContainsRune("+/_-=.:", r):
			hasSpecial = true
		default:
			return false
		}
	}
	if coreLen < 32 {
		return false
	}
	if hexOnly && coreLen >= 32 {
		return true
	}
	if hasDigit && (hasUpper || hasLower) {
		return true
	}
	return hasSpecial && hasUpper && hasLower
}

func redactHighEntropyTokens(value string) string {
	matches := highEntropyPattern.FindAllStringSubmatchIndex(value, -1)
	if len(matches) == 0 {
		return value
	}
	var b strings.Builder
	last := 0
	for _, m := range matches {
		matchStart, matchEnd := m[0], m[1]
		prefixStart, prefixEnd := m[2], m[3]
		tokenStart, tokenEnd := m[4], m[5]
		b.WriteString(value[last:matchStart])
		prefix := value[prefixStart:prefixEnd]
		token := value[tokenStart:tokenEnd]
		b.WriteString(prefix)
		if hasRightTokenBoundary(value, tokenEnd) && looksSensitiveToken(token) {
			b.WriteString("[REDACTED_SECRET]")
		} else {
			b.WriteString(token)
		}
		last = matchEnd
	}
	b.WriteString(value[last:])
	return b.String()
}

func hasRightTokenBoundary(value string, index int) bool {
	if index >= len(value) {
		return true
	}
	return isDelimiter(value[index])
}

func isDelimiter(ch byte) bool {
	switch ch {
	case '"', '\'', '`', ',', ';':
		return true
	default:
		return unicode.IsSpace(rune(ch))
	}
}

func looksPathLike(value string) bool {
	if strings.Count(value, "/") < 2 || !strings.Contains(value, ".") {
		return false
	}
	return !strings.ContainsAny(value, "+=:")
}

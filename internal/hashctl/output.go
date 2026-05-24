package hashctl

import (
	"encoding/json"
	"fmt"
	"io"
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
			payload["error"] = e
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
	return value
}

package hashctl

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	ExitSuccess     = 0
	ExitUsage       = 2
	ExitTransport   = 3
	ExitAPIClient   = 4
	ExitAPIServer   = 5
	ExitPollTimeout = 6
)

const (
	defaultPollInterval = 5 * time.Second
	defaultPollTimeout  = 10 * time.Minute
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type keyValueList map[string]string

func (k keyValueList) String() string {
	if len(k) == 0 {
		return ""
	}
	out := make([]string, 0, len(k))
	for key, value := range k {
		out = append(out, key+"="+value)
	}
	return strings.Join(out, ",")
}

func (k keyValueList) Set(value string) error {
	key, val, ok := strings.Cut(value, "=")
	if !ok {
		return fmt.Errorf("metadata must use key=value format")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("metadata key is required")
	}
	k[key] = val
	return nil
}

type JobCreateRequest struct {
	JobID           string            `json:"job_id,omitempty"`
	SourceAccount   string            `json:"source_account"`
	SourceContainer string            `json:"source_container"`
	OutputAccount   string            `json:"output_account,omitempty"`
	OutputContainer string            `json:"output_container,omitempty"`
	OutputPath      string            `json:"output_path,omitempty"`
	Prefix          string            `json:"prefix,omitempty"`
	BlobNames       []string          `json:"blob_names,omitempty"`
	BatchSize       int               `json:"batch_size,omitempty"`
	CloudName       string            `json:"cloud_name,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type JobRecord struct {
	JobID             string            `json:"job_id"`
	RequestorObjectID string            `json:"requestor_object_id"`
	Status            string            `json:"status"`
	CloudName         string            `json:"cloud_name"`
	SourceAccount     string            `json:"source_account"`
	SourceContainer   string            `json:"source_container"`
	SourcePrefix      *string           `json:"source_prefix"`
	OutputAccount     *string           `json:"output_account"`
	OutputContainer   *string           `json:"output_container"`
	OutputPath        *string           `json:"output_path"`
	BatchSize         *int              `json:"batch_size"`
	ExpectedBlobCount int               `json:"expected_blob_count"`
	CorrelationID     string            `json:"correlation_id"`
	FailureReason     *string           `json:"failure_reason"`
	Metadata          map[string]string `json:"metadata"`
	Raw               json.RawMessage   `json:"-"`
}

type ManifestResponse struct {
	JobID       string `json:"job_id"`
	Status      string `json:"status"`
	ManifestXML string `json:"manifest_xml"`
}

type SignRequest struct {
	JobID      string `json:"job_id"`
	KeyID      string `json:"key_id,omitempty"`
	KeyVersion string `json:"key_version,omitempty"`
	KeyName    string `json:"key_name,omitempty"`
}

type SignatureRecord struct {
	JobID           string         `json:"job_id"`
	SignatureNumber int            `json:"signature_number"`
	SignerObjectID  string         `json:"signer_object_id"`
	KeyID           string         `json:"key_id"`
	KeyVersion      string         `json:"key_version"`
	PolicyMetadata  map[string]any `json:"policy_metadata"`
}

type SignatureResponse struct {
	Job       JobRecord       `json:"job"`
	Signature SignatureRecord `json:"signature"`
}

type SmokeJobResult struct {
	JobID         string `json:"job_id"`
	CorrelationID string `json:"correlation_id"`
	Prefix        string `json:"prefix"`
	FinalState    string `json:"final_state"`
}

type SmokeResult struct {
	TargetState string           `json:"target_state"`
	Jobs        []SmokeJobResult `json:"jobs"`
}

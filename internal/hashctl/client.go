package hashctl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	config     Config
}

func NewClient(cfg Config, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(cfg.APIURL)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{baseURL: parsed, httpClient: httpClient, config: cfg}, nil
}

func (c *Client) Health(ctx context.Context, ready bool) (map[string]any, error) {
	path := "/health"
	if ready {
		path = "/health/ready"
	}
	var out map[string]any
	err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

func (c *Client) CreateJob(ctx context.Context, request JobCreateRequest) (JobRecord, error) {
	var out JobRecord
	err := c.do(ctx, http.MethodPost, "/job/create", request, &out)
	return out, err
}

func (c *Client) GetJob(ctx context.Context, jobID string) (JobRecord, error) {
	var out JobRecord
	err := c.do(ctx, http.MethodGet, "/job/"+url.PathEscape(jobID), nil, &out)
	return out, err
}

func (c *Client) ListJobs(ctx context.Context) ([]JobRecord, error) {
	var out []JobRecord
	err := c.do(ctx, http.MethodGet, "/job/list", nil, &out)
	return out, err
}

func (c *Client) CancelJob(ctx context.Context, jobID string) (JobRecord, error) {
	var out JobRecord
	err := c.do(ctx, http.MethodDelete, "/job/"+url.PathEscape(jobID), nil, &out)
	return out, err
}

func (c *Client) BuildManifest(ctx context.Context, jobID string) (ManifestResponse, error) {
	var out ManifestResponse
	err := c.do(ctx, http.MethodPost, "/job/"+url.PathEscape(jobID)+"/manifest", nil, &out)
	return out, err
}

func (c *Client) GetManifest(ctx context.Context, jobID string) (ManifestResponse, error) {
	var out ManifestResponse
	err := c.do(ctx, http.MethodGet, "/job/"+url.PathEscape(jobID)+"/manifest", nil, &out)
	return out, err
}

func (c *Client) Sign(ctx context.Context, first bool, request SignRequest) (SignatureResponse, error) {
	path := "/sign/second-signature"
	if first {
		path = "/sign/first-signature"
	}
	var out SignatureResponse
	err := c.do(ctx, http.MethodPost, path, request, &out)
	return out, err
}

func (c *Client) WaitForJob(ctx context.Context, jobID string, target string) (JobRecord, error) {
	deadline := time.Now().Add(c.config.PollTimeout)
	interval := c.config.PollInterval
	var last JobRecord
	for {
		job, err := c.GetJob(ctx, jobID)
		if err == nil {
			last = job
			if job.Status == target {
				return job, nil
			}
			if job.Status == "failed" || job.Status == "cancelled" {
				return job, PollError{Job: job, TargetState: target, Terminal: true}
			}
		} else if apiErr, ok := err.(APIError); ok && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			return last, err
		}
		if time.Now().Add(interval).After(deadline) {
			return last, PollError{Job: last, TargetState: target}
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-time.After(interval):
		}
		if interval < 30*time.Second {
			interval *= 2
		}
	}
}

func (c *Client) do(ctx context.Context, method string, path string, body any, out any) error {
	fullURL := c.join(path)
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TransportError{Err: err, Route: method + " " + path}
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return TransportError{Err: err, Route: method + " " + path}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, method+" "+path, data)
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w", method, path, err)
	}
	return nil
}

func (c *Client) join(path string) string {
	base := *c.baseURL
	base.Path = strings.TrimRight(base.Path, "/") + "/" + strings.TrimLeft(path, "/")
	base.RawQuery = ""
	base.Fragment = ""
	return base.String()
}

func (c *Client) applyHeaders(req *http.Request) {
	if c.config.CorrelationID != "" {
		req.Header.Set("x-correlation-id", c.config.CorrelationID)
	}
	if c.config.LocalPrincipalID != "" {
		req.Header.Set("x-hash-engine-local-principal-id", c.config.LocalPrincipalID)
	}
	if len(c.config.LocalGroups) > 0 {
		req.Header.Set("x-hash-engine-local-groups", strings.Join(c.config.LocalGroups, ","))
	}
	if len(c.config.LocalAppRoles) > 0 {
		req.Header.Set("x-hash-engine-local-app-roles", strings.Join(c.config.LocalAppRoles, ","))
	}
	if c.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	}
}

type TransportError struct {
	Err   error
	Route string
}

func (e TransportError) Error() string {
	return e.Err.Error()
}

type APIError struct {
	HTTPStatus    int    `json:"http_status,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	Message       string `json:"message"`
	Route         string `json:"route"`
	JobID         string `json:"job_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func (e APIError) Error() string {
	if e.ErrorCode != "" {
		return e.ErrorCode + ": " + e.Message
	}
	return e.Message
}

type PollError struct {
	Job         JobRecord
	TargetState string
	Terminal    bool
}

func (e PollError) Error() string {
	if e.Terminal {
		return fmt.Sprintf("job %s reached terminal state %s before %s", e.Job.JobID, e.Job.Status, e.TargetState)
	}
	return fmt.Sprintf("timed out waiting for job %s to reach %s; last state %s", e.Job.JobID, e.TargetState, e.Job.Status)
}

func parseAPIError(status int, route string, data []byte) APIError {
	errOut := APIError{HTTPStatus: status, Route: route, Message: strings.TrimSpace(string(data))}
	var envelope struct {
		Detail json.RawMessage `json:"detail"`
	}
	if json.Unmarshal(data, &envelope) != nil || len(envelope.Detail) == 0 {
		if errOut.Message == "" {
			errOut.Message = http.StatusText(status)
		}
		return errOut
	}
	var detailObj struct {
		ErrorCode     string `json:"error_code"`
		Message       string `json:"message"`
		JobID         string `json:"job_id"`
		CorrelationID string `json:"correlation_id"`
	}
	if json.Unmarshal(envelope.Detail, &detailObj) == nil && (detailObj.Message != "" || detailObj.ErrorCode != "") {
		errOut.ErrorCode = detailObj.ErrorCode
		errOut.Message = detailObj.Message
		errOut.JobID = detailObj.JobID
		errOut.CorrelationID = detailObj.CorrelationID
		return errOut
	}
	var detailString string
	if json.Unmarshal(envelope.Detail, &detailString) == nil && detailString != "" {
		errOut.Message = detailString
		return errOut
	}
	errOut.Message = strings.TrimSpace(string(envelope.Detail))
	return errOut
}

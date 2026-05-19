package hashctl

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sync"
)

func (a *app) runSmoke(ctx context.Context, client *Client, cfg Config, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("smoke subcommand is required")
	}
	switch args[0] {
	case "single-job":
		return a.runSingleSmoke(ctx, client, cfg, args[1:])
	case "multi-job":
		return a.runMultiSmoke(ctx, client, cfg, args[1:])
	default:
		return nil, fmt.Errorf("unknown smoke subcommand %q", args[0])
	}
}

func (a *app) runSingleSmoke(ctx context.Context, client *Client, cfg Config, args []string) (SmokeResult, error) {
	fs := flag.NewFlagSet("smoke single-job", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var sourceAccount, sourceContainer, prefix, runID, target, cloudName string
	var batchSize int
	fs.StringVar(&sourceAccount, "source-account", "", "source account")
	fs.StringVar(&sourceContainer, "source-container", "", "source container")
	fs.StringVar(&prefix, "prefix", "", "source prefix")
	fs.StringVar(&runID, "run-id", "", "run identifier")
	fs.StringVar(&target, "target-state", "awaiting_first_signature", "target job state")
	fs.StringVar(&cloudName, "cloud-name", "azure_government", "cloud profile")
	fs.IntVar(&batchSize, "batch-size", 1, "batch size")
	if err := fs.Parse(args); err != nil {
		return SmokeResult{}, err
	}
	if sourceAccount == "" || sourceContainer == "" || prefix == "" {
		return SmokeResult{}, fmt.Errorf("--source-account, --source-container, and --prefix are required")
	}
	correlationID := cfg.CorrelationID
	if correlationID == "" {
		correlationID = runID
	}
	if correlationID == "" {
		correlationID = "hashctl-smoke-single"
	}
	smokeClient, err := clientWithCorrelation(cfg, client, correlationID)
	if err != nil {
		return SmokeResult{}, err
	}
	job, err := smokeClient.CreateJob(ctx, JobCreateRequest{
		SourceAccount:   sourceAccount,
		SourceContainer: sourceContainer,
		Prefix:          prefix,
		BatchSize:       batchSize,
		CloudName:       cloudName,
		Metadata: map[string]string{
			"verification": "hashctl-smoke-single",
			"run_id":       runID,
		},
	})
	if err != nil {
		return SmokeResult{}, err
	}
	final, err := smokeClient.WaitForJob(ctx, job.JobID, target)
	if err != nil {
		return SmokeResult{}, err
	}
	return SmokeResult{
		TargetState: target,
		Jobs: []SmokeJobResult{{
			JobID:         final.JobID,
			CorrelationID: correlationID,
			Prefix:        prefix,
			FinalState:    final.Status,
		}},
	}, nil
}

func (a *app) runMultiSmoke(ctx context.Context, client *Client, cfg Config, args []string) (SmokeResult, error) {
	fs := flag.NewFlagSet("smoke multi-job", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var sourceAccount, sourceContainer, runID, target, cloudName string
	fs.StringVar(&sourceAccount, "source-account", "", "source account")
	fs.StringVar(&sourceContainer, "source-container", "", "source container")
	fs.StringVar(&runID, "run-id", "", "run identifier")
	fs.StringVar(&target, "target-state", "awaiting_first_signature", "target job state")
	fs.StringVar(&cloudName, "cloud-name", "azure_government", "cloud profile")
	if err := fs.Parse(args); err != nil {
		return SmokeResult{}, err
	}
	if sourceAccount == "" || sourceContainer == "" || runID == "" {
		return SmokeResult{}, fmt.Errorf("--source-account, --source-container, and --run-id are required")
	}

	type submittedJob struct {
		jobID         string
		correlationID string
		prefix        string
	}
	submitted := make([]submittedJob, 5)
	var submitWG sync.WaitGroup
	var submitMu sync.Mutex
	var submitErr error
	for index := 1; index <= 5; index++ {
		index := index
		submitWG.Add(1)
		go func() {
			defer submitWG.Done()
			prefix := fmt.Sprintf("%s/job-0%d/", runID, index)
			correlationID := fmt.Sprintf("%s-job-0%d", runID, index)
			smokeClient, err := clientWithCorrelation(cfg, client, correlationID)
			if err != nil {
				submitMu.Lock()
				if submitErr == nil {
					submitErr = err
				}
				submitMu.Unlock()
				return
			}
			job, err := smokeClient.CreateJob(ctx, JobCreateRequest{
				SourceAccount:   sourceAccount,
				SourceContainer: sourceContainer,
				Prefix:          prefix,
				BatchSize:       1,
				CloudName:       cloudName,
				Metadata: map[string]string{
					"verification": "hashctl-smoke-multi",
					"run_id":       runID,
					"job_index":    fmt.Sprintf("%d", index),
				},
			})
			if err != nil {
				submitMu.Lock()
				if submitErr == nil {
					submitErr = err
				}
				submitMu.Unlock()
				return
			}
			submitted[index-1] = submittedJob{
				jobID:         job.JobID,
				correlationID: correlationID,
				prefix:        prefix,
			}
		}()
	}
	submitWG.Wait()
	if submitErr != nil {
		return SmokeResult{}, submitErr
	}

	result := SmokeResult{TargetState: target}
	for _, submittedJob := range submitted {
		smokeClient, err := clientWithCorrelation(cfg, client, submittedJob.correlationID)
		if err != nil {
			return SmokeResult{}, err
		}
		final, err := smokeClient.WaitForJob(ctx, submittedJob.jobID, target)
		if err != nil {
			return SmokeResult{}, err
		}
		result.Jobs = append(result.Jobs, SmokeJobResult{
			JobID:         final.JobID,
			CorrelationID: submittedJob.correlationID,
			Prefix:        submittedJob.prefix,
			FinalState:    final.Status,
		})
	}
	return result, nil
}

func clientWithCorrelation(cfg Config, base *Client, correlationID string) (*Client, error) {
	cfg.CorrelationID = correlationID
	return NewClient(cfg, base.httpClient)
}

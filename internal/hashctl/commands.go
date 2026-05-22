package hashctl

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type app struct {
	stdout     io.Writer
	stderr     io.Writer
	getenv     envLookup
	httpClient *http.Client
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	return New(stdout, stderr, os.Getenv, nil).Run(args)
}

func New(stdout io.Writer, stderr io.Writer, getenv envLookup, httpClient *http.Client) *app {
	return &app{stdout: stdout, stderr: stderr, getenv: getenv, httpClient: httpClient}
}

func (a *app) Run(args []string) int {
	global, commandArgs, err := parseGlobal(args)
	if err != nil {
		writeError(a.stderr, global.Output, err)
		return ExitUsage
	}
	if len(commandArgs) == 0 {
		writeError(a.stderr, global.Output, fmt.Errorf("command is required"))
		return ExitUsage
	}
	cfg, err := resolveConfig(global, a.getenv)
	if err != nil {
		writeError(a.stderr, global.Output, err)
		return ExitUsage
	}
	client, err := NewClient(cfg, a.httpClient)
	if err != nil {
		writeError(a.stderr, cfg.Output, err)
		return ExitUsage
	}
	commandTimeout := cfg.Timeout
	if commandArgs[0] == "smoke" && cfg.PollTimeout+cfg.PollInterval > commandTimeout {
		commandTimeout = cfg.PollTimeout + cfg.PollInterval
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	var result any
	switch commandArgs[0] {
	case "health":
		result, err = a.runHealth(ctx, client, cfg, commandArgs[1:])
	case "job":
		result, err = a.runJob(ctx, client, cfg, commandArgs[1:])
	case "manifest":
		result, err = a.runManifest(ctx, client, cfg, commandArgs[1:])
	case "sign":
		result, err = a.runSign(ctx, client, cfg, commandArgs[1:])
	case "smoke":
		result, err = a.runSmoke(ctx, client, cfg, commandArgs[1:])
	default:
		err = fmt.Errorf("unknown command %q", commandArgs[0])
	}
	if err != nil {
		writeError(a.stderr, cfg.Output, err)
		return errorExitCode(err)
	}
	if result == nil {
		return ExitSuccess
	}
	if cfg.Output == "json" {
		if err := writeJSON(a.stdout, result); err != nil {
			writeError(a.stderr, cfg.Output, err)
			return ExitUsage
		}
		return ExitSuccess
	}
	if err := a.writeHuman(result); err != nil {
		writeError(a.stderr, cfg.Output, err)
		return ExitUsage
	}
	return ExitSuccess
}

func parseGlobal(args []string) (Config, []string, error) {
	cfg := Config{}
	var groups stringList
	var roles stringList
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			cfg.LocalGroups = groups
			cfg.LocalAppRoles = roles
			return cfg, args[idx:], nil
		}
		name, value, hasValue := strings.Cut(arg, "=")
		take := func() (string, error) {
			if hasValue {
				return value, nil
			}
			idx++
			if idx >= len(args) {
				return "", fmt.Errorf("%s requires a value", name)
			}
			return args[idx], nil
		}
		switch name {
		case "--api-url":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.APIURL = v
		case "--config":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.ConfigPath = v
		case "--output", "-o":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.Output = v
		case "--timeout":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.Timeout, err = time.ParseDuration(v)
			if err != nil {
				return cfg, nil, fmt.Errorf("invalid timeout: %w", err)
			}
		case "--correlation-id":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.CorrelationID = v
		case "--local-principal-id":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.LocalPrincipalID = v
		case "--local-group", "--local-groups":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			for _, item := range strings.Split(v, ",") {
				item = strings.TrimSpace(item)
				if item != "" {
					groups = append(groups, item)
				}
			}
		case "--local-app-role", "--local-app-roles":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			for _, item := range strings.Split(v, ",") {
				item = strings.TrimSpace(item)
				if item != "" {
					roles = append(roles, item)
				}
			}
		case "--bearer-token-file":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.BearerTokenFile = v
		case "--bearer-token":
			return cfg, nil, fmt.Errorf("literal bearer token arguments are not supported; use HASH_ENGINE_BEARER_TOKEN or --bearer-token-file")
		case "--poll-interval":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.PollInterval, err = time.ParseDuration(v)
			if err != nil {
				return cfg, nil, fmt.Errorf("invalid poll interval: %w", err)
			}
		case "--poll-timeout":
			v, err := take()
			if err != nil {
				return cfg, nil, err
			}
			cfg.PollTimeout, err = time.ParseDuration(v)
			if err != nil {
				return cfg, nil, fmt.Errorf("invalid poll timeout: %w", err)
			}
		case "--help", "-h":
			return cfg, []string{"help"}, nil
		default:
			return cfg, nil, fmt.Errorf("unknown global flag %s", name)
		}
	}
	cfg.LocalGroups = groups
	cfg.LocalAppRoles = roles
	return cfg, nil, nil
}

func (a *app) runHealth(ctx context.Context, client *Client, cfg Config, args []string) (any, error) {
	ready := false
	if len(args) > 0 {
		if len(args) != 1 || args[0] != "ready" {
			return nil, fmt.Errorf("usage: hashctl health [ready]")
		}
		ready = true
	}
	body, err := client.Health(ctx, ready)
	if err != nil {
		return nil, err
	}
	if cfg.Output == "json" {
		return body, nil
	}
	if ready {
		_, err = fmt.Fprintln(a.stdout, "ready")
	} else {
		_, err = fmt.Fprintln(a.stdout, "healthy")
	}
	return nil, err
}

func (a *app) runJob(ctx context.Context, client *Client, _ Config, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("job subcommand is required")
	}
	switch args[0] {
	case "create":
		return a.runJobCreate(ctx, client, args[1:])
	case "get":
		if len(args) != 2 {
			return nil, fmt.Errorf("usage: hashctl job get <job_id>")
		}
		return client.GetJob(ctx, args[1])
	case "list":
		if len(args) != 1 {
			return nil, fmt.Errorf("usage: hashctl job list")
		}
		return client.ListJobs(ctx)
	case "cancel":
		if len(args) != 2 {
			return nil, fmt.Errorf("usage: hashctl job cancel <job_id>")
		}
		return client.CancelJob(ctx, args[1])
	default:
		return nil, fmt.Errorf("unknown job subcommand %q", args[0])
	}
}

func (a *app) runJobCreate(ctx context.Context, client *Client, args []string) (any, error) {
	fs := flag.NewFlagSet("job create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var blobs stringList
	metadata := keyValueList{}
	request := JobCreateRequest{Metadata: metadata}
	var blobListFile string
	fs.StringVar(&request.JobID, "job-id", "", "optional job id")
	fs.StringVar(&request.SourceAccount, "source-account", "", "source storage account")
	fs.StringVar(&request.SourceContainer, "source-container", "", "source container")
	fs.StringVar(&request.OutputAccount, "output-account", "", "output storage account")
	fs.StringVar(&request.OutputContainer, "output-container", "", "output container")
	fs.StringVar(&request.OutputPath, "output-path", "", "output path")
	fs.StringVar(&request.Prefix, "prefix", "", "source prefix")
	fs.Var(&blobs, "blob-name", "explicit blob name")
	fs.StringVar(&blobListFile, "blob-list-file", "", "newline-delimited blob list")
	fs.IntVar(&request.BatchSize, "batch-size", 0, "batch size")
	fs.StringVar(&request.CloudName, "cloud-name", "", "cloud profile")
	fs.Var(metadata, "metadata", "metadata key=value")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if request.SourceAccount == "" || request.SourceContainer == "" {
		return nil, fmt.Errorf("--source-account and --source-container are required")
	}
	request.BlobNames = append(request.BlobNames, blobs...)
	if blobListFile != "" {
		fileBlobs, err := readBlobList(blobListFile)
		if err != nil {
			return nil, err
		}
		request.BlobNames = append(request.BlobNames, fileBlobs...)
	}
	if request.BatchSize < 0 {
		return nil, fmt.Errorf("--batch-size must be at least 1 when supplied")
	}
	if len(request.Metadata) == 0 {
		request.Metadata = nil
	}
	return client.CreateJob(ctx, request)
}

func readBlobList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read blob list file: %w", err)
	}
	var blobs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			blobs = append(blobs, line)
		}
	}
	return blobs, nil
}

func (a *app) runManifest(ctx context.Context, client *Client, cfg Config, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("manifest subcommand is required")
	}
	switch args[0] {
	case "build":
		if len(args) != 2 {
			return nil, fmt.Errorf("usage: hashctl manifest build <job_id>")
		}
		return client.BuildManifest(ctx, args[1])
	case "get":
		fs := flag.NewFlagSet("manifest get", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		var outputFile string
		fs.StringVar(&outputFile, "output-file", "", "write manifest XML to file")
		if err := fs.Parse(args[1:]); err != nil {
			return nil, err
		}
		if fs.NArg() != 1 {
			return nil, fmt.Errorf("usage: hashctl manifest get <job_id> [--output-file path]")
		}
		manifest, err := client.GetManifest(ctx, fs.Arg(0))
		if err != nil {
			return nil, err
		}
		if cfg.Output == "json" {
			return manifest, nil
		}
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(manifest.ManifestXML), 0o600); err != nil {
				return nil, fmt.Errorf("write manifest file: %w", err)
			}
			_, err = fmt.Fprintf(a.stdout, "wrote manifest %s to %s\n", manifest.JobID, outputFile)
			return nil, err
		}
		_, err = fmt.Fprint(a.stdout, manifest.ManifestXML)
		return nil, err
	default:
		return nil, fmt.Errorf("unknown manifest subcommand %q", args[0])
	}
}

func (a *app) runSign(ctx context.Context, client *Client, _ Config, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("sign subcommand is required")
	}
	first := args[0] == "first"
	if !first && args[0] != "second" {
		return nil, fmt.Errorf("unknown sign subcommand %q", args[0])
	}
	fs := flag.NewFlagSet("sign "+args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	req := SignRequest{}
	fs.StringVar(&req.KeyID, "key-id", "", "development key id")
	fs.StringVar(&req.KeyVersion, "key-version", "", "development key version")
	fs.StringVar(&req.KeyName, "key-name", "", "existing Key Vault key name")
	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}
	if fs.NArg() != 1 {
		return nil, fmt.Errorf("usage: hashctl sign %s <job_id>", args[0])
	}
	req.JobID = fs.Arg(0)
	return client.Sign(ctx, first, req)
}

func (a *app) writeHuman(result any) error {
	switch value := result.(type) {
	case JobRecord:
		return writeJobSummary(a.stdout, value)
	case []JobRecord:
		return writeJobListSummary(a.stdout, value)
	case ManifestResponse:
		return writeManifestSummary(a.stdout, value)
	case SignatureResponse:
		return writeSignatureSummary(a.stdout, value)
	case SmokeResult:
		return writeSmokeSummary(a.stdout, value)
	default:
		return writeJSON(a.stdout, value)
	}
}

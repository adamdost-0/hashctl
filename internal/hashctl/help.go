package hashctl

import (
	"fmt"
	"io"
	"strings"
)

func writeHelp(w io.Writer, path []string) error {
	switch strings.Join(path, " ") {
	case "":
		_, err := fmt.Fprint(w, topLevelHelpText)
		return err
	case "health":
		_, err := fmt.Fprint(w, healthHelpText)
		return err
	case "job":
		_, err := fmt.Fprint(w, jobHelpText)
		return err
	case "job create":
		_, err := fmt.Fprint(w, jobCreateHelpText)
		return err
	case "job get":
		_, err := fmt.Fprint(w, jobGetHelpText)
		return err
	case "job list":
		_, err := fmt.Fprint(w, jobListHelpText)
		return err
	case "job cancel":
		_, err := fmt.Fprint(w, jobCancelHelpText)
		return err
	case "manifest":
		_, err := fmt.Fprint(w, manifestHelpText)
		return err
	case "manifest build":
		_, err := fmt.Fprint(w, manifestBuildHelpText)
		return err
	case "manifest get":
		_, err := fmt.Fprint(w, manifestGetHelpText)
		return err
	case "sign":
		_, err := fmt.Fprint(w, signHelpText)
		return err
	case "sign first":
		_, err := fmt.Fprint(w, signFirstHelpText)
		return err
	case "sign second":
		_, err := fmt.Fprint(w, signSecondHelpText)
		return err
	case "smoke":
		_, err := fmt.Fprint(w, smokeHelpText)
		return err
	case "smoke single-job":
		_, err := fmt.Fprint(w, smokeSingleHelpText)
		return err
	case "smoke multi-job":
		_, err := fmt.Fprint(w, smokeMultiHelpText)
		return err
	case "version":
		_, err := fmt.Fprint(w, versionHelpText)
		return err
	default:
		return fmt.Errorf("unknown help topic %q", strings.Join(path, " "))
	}
}

const topLevelHelpText = `Hash Engine command-line client for CTS manifest jobs and signing.

Usage:
  hashctl [global flags] <command> [command flags]
  hashctl help [command]

Commands:
  health                Liveness and readiness checks
  job                   Create, list, inspect, and cancel jobs
  manifest              Build and retrieve manifest XML
  sign                  Apply first or second signature
  smoke                 Run smoke test flows
  version               Print hashctl version

Common commands:
  hashctl health
  hashctl job create --source-account NAME --source-container NAME [flags]
  hashctl sign first <job_id> --key-name NAME
  hashctl sign second <job_id> --key-name NAME

Global flags (must appear before <command>):
  --api-url URL
  --config PATH
  --output human|json
  --timeout DURATION
  --correlation-id ID
  --local-principal-id ID
  --local-groups LIST
  --local-app-roles LIST
  --bearer-token-file PATH
  --poll-interval DURATION
  --poll-timeout DURATION
  --help, -h
  --version
`

const healthHelpText = `Usage:
  hashctl health [ready]

Examples:
  hashctl health
  hashctl health ready
`

const jobHelpText = `Usage:
  hashctl job <subcommand> [flags]

Subcommands:
  create                Create a new hashing job
  get <job_id>          Retrieve one job
  list                  List jobs
  cancel <job_id>       Cancel a job
`

const jobCreateHelpText = `Usage:
  hashctl job create --source-account NAME --source-container NAME [flags]

Flags:
  --job-id ID
  --source-account NAME
  --source-container NAME
  --output-account NAME
  --output-container NAME
  --output-path PATH
  --prefix PREFIX
  --blob-name NAME                (repeatable)
  --blob-list-file PATH
  --batch-size N
  --cloud-name NAME
  --metadata key=value            (repeatable)
`

const jobGetHelpText = `Usage:
  hashctl job get <job_id>
`

const jobListHelpText = `Usage:
  hashctl job list
`

const jobCancelHelpText = `Usage:
  hashctl job cancel <job_id>
`

const manifestHelpText = `Usage:
  hashctl manifest <subcommand> [flags]

Subcommands:
  build <job_id>                 Trigger manifest build
  get <job_id> [--output-file]   Retrieve manifest XML
`

const manifestBuildHelpText = `Usage:
  hashctl manifest build <job_id>
`

const manifestGetHelpText = `Usage:
  hashctl manifest get <job_id> [--output-file PATH]
`

const signHelpText = `Usage:
  hashctl sign <first|second> <job_id> [flags]

Flags:
  --key-id ID
  --key-version VERSION
  --key-name NAME
`

const signFirstHelpText = `Usage:
  hashctl sign first <job_id> [--key-id ID] [--key-version VERSION] [--key-name NAME]
`

const signSecondHelpText = `Usage:
  hashctl sign second <job_id> [--key-id ID] [--key-version VERSION] [--key-name NAME]
`

const smokeHelpText = `Usage:
  hashctl smoke <single-job|multi-job> [flags]
`

const smokeSingleHelpText = `Usage:
  hashctl smoke single-job --source-account NAME --source-container NAME --prefix PREFIX [flags]

Flags:
  --run-id ID
  --target-state STATE              (default awaiting_first_signature)
  --cloud-name NAME                 (default azure_government)
  --batch-size N                    (default 1)
`

const smokeMultiHelpText = `Usage:
  hashctl smoke multi-job --source-account NAME --source-container NAME --run-id ID [flags]

Flags:
  --target-state STATE              (default awaiting_first_signature)
  --cloud-name NAME                 (default azure_government)
`

const versionHelpText = `Usage:
  hashctl version
  hashctl --version
`

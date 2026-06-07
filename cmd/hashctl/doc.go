/*
Hashctl is a command-line client for the Hash Engine REST API.

Usage:

	hashctl [global flags] <command> [command flags]
	hashctl help [command]

Commands:

	health      Liveness and readiness checks
	job         Create, list, inspect, and cancel jobs
	manifest    Build and retrieve manifest XML
	sign        Apply first or second signature
	smoke       Run smoke test flows
	version     Print hashctl version

Global flags must appear before the command. The API URL must use HTTPS for all
non-loopback hosts. The literal --bearer-token flag is intentionally unsupported;
provide tokens via HASH_ENGINE_BEARER_TOKEN or a chmod 600 --bearer-token-file.

All application logic lives in the internal/hashctl package; this package is a thin
entry point that calls [github.com/adamdost-0/hashctl/internal/hashctl.Run] and exits
with the returned status code. See the README and docs/ for full documentation.
*/
package main

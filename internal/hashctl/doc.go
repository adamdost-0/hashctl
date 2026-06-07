// Package hashctl implements the hashctl command-line client for the Hash Engine
// REST API. It manages CTS manifest hashing jobs and signing workflows over HTTP(S)
// and supports both human-readable and JSON output.
//
// # Entry point
//
// [Run] is the production entry point: it constructs an app with os.Getenv and the
// default HTTP client, then dispatches the requested command. [New] injects stdout,
// stderr, an environment lookup, and an *http.Client, and is the seam used by tests.
//
// # Configuration
//
// Configuration is merged in precedence order: built-in defaults, then a config file,
// then environment variables (HASH_ENGINE_API, HASH_ENGINE_BEARER_TOKEN), then CLI
// flags. Non-loopback API URLs must use HTTPS, and bearer tokens are refused over
// non-loopback plaintext HTTP.
//
// # Exit codes
//
// [ExitSuccess] (0), [ExitUsage] (2), [ExitTransport] (3), [ExitAPIClient] (4),
// [ExitAPIServer] (5), and [ExitPollTimeout] (6).
//
// # Security
//
// All user-facing output is routed through redaction to strip JWTs, Azure SAS query
// parameters, AccountKey values, and other high-entropy secrets. [Version] is set at
// build time via -ldflags.
package hashctl

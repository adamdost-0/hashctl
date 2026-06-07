# Security Policy

The maintainers take the security of `hashctl` seriously. `hashctl` is a CLI that
handles bearer tokens, signing operations, and Azure storage credentials, so we treat
credential handling and output redaction as first-class concerns.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Instead, report them privately through GitHub's coordinated disclosure feature:

1. Go to the [**Security** tab](https://github.com/adamdost-0/hashctl/security) of this
   repository.
2. Click **Report a vulnerability** to open a private security advisory.

Please include as much of the following as you can to help us triage quickly:

- The type of issue (e.g. credential leakage, TLS downgrade, command injection)
- Full paths of the source file(s) related to the issue
- The location of the affected source code (tag, branch, commit, or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce
- Proof-of-concept or exploit code, if possible
- The impact of the issue and how an attacker might exploit it

You should receive an acknowledgement within **5 business days**. We aim to provide a
remediation plan or status update within **30 days** of acknowledgement.

## Coordinated Disclosure

This project follows the principle of
[Coordinated Vulnerability Disclosure](https://aka.ms/security.md/cvd). Please give us a
reasonable opportunity to release a fix before any public disclosure.

## Supported Versions

Until `hashctl` reaches `1.0.0`, only the latest released version (and `main`) receives
security fixes. See [`CHANGELOG.md`](CHANGELOG.md) for release history.

## Preferred Languages

We prefer all communications to be in English.

<div align="center">

# Mavetis

**Enterprise-grade security analysis for Git-based development workflows**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8.svg)](https://go.dev)
[![Status](https://img.shields.io/badge/Status-Stable-brightgreen.svg)](#)

Complete static analysis with zero external dependencies, zero network calls, and zero telemetry.

</div>

---

## Contents

[Overview](#overview) · [Installation](#installation) · [Usage](#usage) · [Configuration](#configuration) · [Rules](#rules) · [Baseline](#baseline) · [Detection](#detection) · [Output](#output) · [Hooks](#hooks) · [Updates](#updates) · [Development](#development)

---

<a id="overview"></a>
## Overview

Mavetis delivers change-focused security analysis with complete network isolation. Pure Go standard library implementation. No third-party dependencies.

**Core Capabilities**

| Capability | Description |
|---|---|
| Air-Gapped Operation | Complete offline analysis with zero external network dependencies |
| Change-Focused Analysis | Security evaluation of staged changes, branch diffs, and merge candidates |
| File Review Mode | Direct security review of arbitrary local files without Git diff context |
| Regression Prevention | Detection of removed security controls and weakened policies |
| Policy-Aware Review | Built-in review profiles and trust zones for risk-weighted analysis |
| Boundary Enforcement | Architectural boundary checks for privileged modules and trust edges |
| Flexible Rule Engine | Customizable YAML-based rules with contextual scoping |
| Enterprise Integration | Native JSON and SARIF output for CI/CD pipelines |
| Supply-Chain Trust | Dependency lifecycle correlation and registry trust policies |
| Security Intent Analysis | Detects security-named code that no longer performs protective logic |

---

<a id="installation"></a>
## Installation

**macOS and Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/pimatis/mavetis/main/install.sh | sh
```

**Windows PowerShell**

```powershell
iwr https://raw.githubusercontent.com/pimatis/mavetis/main/install.ps1 -UseBasicParsing | iex
```

**Go Install**

```bash
go install github.com/pimatis/mavetis@latest
```

**Removal**

```bash
# macOS / Linux
sudo rm -f /usr/local/bin/mavetis
rm -f "$HOME/.local/bin/mavetis"

# Windows
Remove-Item "$HOME\AppData\Local\mavetis\bin\mavetis.exe" -Force
```

---

<a id="usage"></a>
## Usage

```bash
# Review staged auth changes with bounded imported dependency context
mavetis review --staged --path 'src/**' --profile auth --with-context --explain

# Compare backend changes against base branch
mavetis review --base main --path 'src/**' --profile backend --with-context

# Review local files directly (no Git diff required)
mavetis review src/auth/login.go src/api/handler.ts --explain

# CI/CD integration with JSON output
mavetis ci --base main --format json --profile fintech

# Initialize project configuration interactively
mavetis init

# Create a baseline from current findings to suppress known issues
mavetis baseline --create --base main

# Explain why a rule fires and how to reduce false positives
mavetis rules explain --id inject.sql.raw

# Install Git hooks for automated scanning
mavetis hooks install
```

### Command Reference

| Command | Description |
|---|---|
| `mavetis review` | Analyze code changes or file targets with configurable scope and rule profile |
| `mavetis ci` | Optimized analysis for CI/CD with profile-aware policy evaluation |
| `mavetis init` | Initialize project configuration with interactive or default `.mavetis.yaml` |
| `mavetis baseline --create` | Capture current findings as a baseline to suppress known issues |
| `mavetis hooks install` | Configure pre-commit and pre-push scanning |
| `mavetis hooks uninstall` | Remove configured Git hooks |
| `mavetis rules validate` | Validate custom rule definitions |
| `mavetis rules list` | Display available security rules |
| `mavetis rules show` | Display detailed rule information |
| `mavetis rules explain` | Explain rule triggers, guards, examples, and remediation |
| `mavetis rules test` | Test rules against sample diffs |
| `mavetis rules matrix` | Generate compliance coverage matrix |
| `mavetis rules snapshot` | Generate repository security snapshots |
| `mavetis update` | Self-update to latest version |
| `mavetis version` | Display version information |

### File Review Mode

Scan local files directly using the same engine, rule DSL, and output formats without requiring staged or branch diff data.

```bash
mavetis review src/auth/login.go --explain
mavetis review src/auth/*.go --severity high
mavetis review src/scan/load.go --with-suggested
mavetis review @config/nginx.conf --profile backend --format json
```

- Accepts plain relative paths and `@path` targets
- Rejects binary targets, directories, and oversized files
- Emits bounded local dependency suggestions for nearby imports
- `--with-suggested` reviews those suggested files in the same run
- Diff and CI review support `--with-context` / `--changed-with-context` to include bounded local dependencies imported by changed files

---

<a id="configuration"></a>
## Configuration

Mavetis loads configuration from `.mavetis.yaml` or `.mavetis.yml` in the working directory.

```yaml
severity: low
fail-on: high
output: text
profile: fintech
ignore:
  - vendor/**
allow:
  paths:
    - fixtures/**
  values:
    - example-secret
  regexes:
    - '^demo_[A-Za-z0-9]+$'
company:
  prefixes:
    - corp_
supply:
  allow-packages:
    - '@company/*'
  deny-packages:
    - left-pad
    - event-stream
  trusted-registries:
    - registry.company.local
snapshot:
  path: .mavetis-snapshots.yaml
baseline:
  path: .mavetis-baseline.yaml
zones:
  critical:
    - src/auth/**
    - src/lib/security/**
    - src/api/admin/**
  restricted:
    - src/payments/**
    - src/backoffice/**
```

### Profiles

| Profile | Focus |
|---|---|
| `auth` | Authentication, authorization, session, token, crypto, and related telemetry |
| `fintech` | Full default policy surface for high-assurance review workflows |
| `backend` | Server-side security, supply-chain, config, network, and abuse-prevention |
| `frontend` | Browser-facing auth, session, XSS, CORS, privacy, and client config |

### Trust Zones

| Zone | Behavior |
|---|---|
| `zones.critical` | Raises findings by two severity levels; blocks at `fail-on=low` |
| `zones.restricted` | Raises findings by one severity level; blocks at `fail-on=medium` |

---

<a id="rules"></a>
## Rules

### Custom Security Rules

Define organization-specific policies through YAML rules:

```yaml
rules:
  - id: company.fetch.untrusted
    title: Untrusted Fetch Operation
    message: Request-controlled URL reached a sensitive fetch sink.
    remediation: Validate and allowlist outbound destinations before processing.
    vulnerable-example: fetch(request.query.url)
    safe-example: fetch(allowlistedOutboundURL(request.query.url))
    category: inject
    severity: high
    confidence: medium
    target: added
    paths:
      - src/**
    require:
      - '(?i)fetch'
    near:
      - 'query|params|body'
    absent:
      - 'allowlist|whitelist|trustedHost'
    standards:
      - OWASP-ASVS-V5.3
```

### Rule Explanations

Explain a builtin, custom, or synthetic semantic rule without running a scan:

```bash
mavetis rules explain --id inject.sql.raw
mavetis rules explain --id semantic.go.ssrf
mavetis explain rule semantic.go.ssrf
```

The explanation output includes rule title, severity, confidence, category, ASVS mappings, trigger patterns, positive context, negative context / absent guards, representative vulnerable and safe snippets, and remediation guidance. Custom rules can provide `vulnerable-example` and `safe-example` fields so security teams can document organization-specific triage guidance next to the matcher definition.

### Typed Policies

```yaml
rules:
  - id: company.ui.auth-boundary
    type: forbiddenImport
    title: UI cannot import server auth helpers
    message: UI code imported a privileged auth helper.
    remediation: Move the logic behind a reviewed server boundary.
    category: boundary
    severity: high
    target: added
    paths:
      - src/ui/**
    imports:
      - '(?i)(^|/)(auth|security|internal)(/|$)'

  - id: company.prod-mode
    type: configKeyConstraint
    title: Runtime mode must stay production
    message: Runtime mode drifted outside the approved production value.
    remediation: Keep deployable runtime configuration pinned to production.
    category: config
    severity: high
    target: added
    key: NODE_ENV
    allowed-values:
      - production
```

### Rule Matchers

| Matcher | Description |
|---|---|
| `require` | Mandatory pattern presence |
| `any` | Alternative pattern matching |
| `near` | Contextual proximity matching |
| `absent` | Negative pattern matching |
| `forbiddenImport` | Block imports from forbidden modules |
| `deletedLineGuard` | Treat deleted security guard lines as violations |
| `forbiddenEnv` | Forbid risky environment keys |
| `requiredMiddleware` | Enforce required middleware on routes |
| `requiredCall` | Enforce critical side-effect calls |
| `configKeyConstraint` | Constrain config keys by allowed values or ranges |
| `pathBoundary` | Express source-to-target trust boundaries |

### Repository Snapshots

Generate and enforce repository-specific security baselines:

```bash
mavetis rules snapshot --output .mavetis-snapshots.yaml --path 'src/auth/**'
```

Enable in configuration:

```yaml
snapshot:
  path: .mavetis-snapshots.yaml
```

---

<a id="baseline"></a>
## Baseline / Suppressions

Legacy codebases often contain a large number of historical findings that cannot be addressed immediately. Without a baseline, every scan produces the same noise and the tool becomes unusable in practice.

Mavetis supports baselines so teams can record known findings and focus only on newly introduced issues.

### Creating a Baseline

```bash
mavetis baseline --create --base main
mavetis baseline --create --output .mavetis-baseline.yaml --base main
```

This runs a full review against the specified base, captures all findings, and writes them to `.mavetis-baseline.yaml`. The baseline file is automatically added to `.gitignore`.

### Using a Baseline

Pass the baseline file during review to suppress known findings:

```bash
mavetis review --base main --baseline .mavetis-baseline.yaml
mavetis ci --base main --baseline .mavetis-baseline.yaml
```

You can also set the baseline path in `.mavetis.yaml`:

```yaml
baseline:
  path: .mavetis-baseline.yaml
```

When a baseline is configured, only findings not present in the baseline are reported. This makes CI integration practical for teams working with existing code.

### Baseline File Format

```yaml
# Mavetis baseline
# Known findings suppressed in subsequent reviews

baseline:
  - rule: inject.sql.raw
    path: src/api/handler.go
    line: 45
  - rule: secret.generic
    path: config/.env
    line: 3
```

---

<a id="detection"></a>
## Detection

### Secrets and Cryptography

- Cloud provider credentials (AWS, Stripe, Supabase)
- Configuration file secrets (dotenv, JWT)
- Private key exposure and high-entropy secret patterns
- Weak randomness, hashing, and ciphers
- IV/nonce misuse and key confusion attacks

### Access Control and Sessions

- Authentication bypass and middleware removal
- Insecure token storage and session fixation
- Token rotation failures and scope filter removal
- IDOR patterns and operation-level permission regressions
- JWT security flaws (decode-without-verify, missing binding)
- OAuth weaknesses (state, PKCE, nonce, replay attacks)
- Password reset token logging and password-change reauthentication gaps
- Multi-tenant lookup and cross-tenant query scope gaps
- Webhook signature, raw-body, and replay-window weaknesses

### Injection and Input Validation

- SSRF, SQL injection, command injection, XSS
- Unsafe deserialization and path traversal
- File upload validation gaps and CORS misconfiguration
- TLS validation disablement and stack trace disclosure
- Dynamic code evaluation (eval) and SSTI

### Supply Chain

- Remote and git-based dependencies
- Version pinning violations and typosquatting
- Lockfile integrity and integrity hash removal
- Install-time script execution and shell downloads
- Mutable GitHub Action references
- Registry trust enforcement
- Cloud object storage public access and long-lived presigned URLs
- Wildcard IAM policies and public SSH ingress

### AI and LLM Security

- Secret material sent to prompts or model messages
- User-controlled content assigned to privileged system prompts
- Tool/function execution from untrusted model output without validation

### Regression Detection

- Deleted authentication and authorization middleware
- Removed access control checks and validation routines
- Timeout and rate limiting removal
- SameSite weakening, cookie lifetime expansion
- bcrypt cost downgrades and MFA weakening
- Architectural boundary violations
- Snapshot regressions against captured baselines

---

<a id="output"></a>
## Output

| Format | Use Case |
|---|---|
| `text` | Human-readable with ANSI colors for terminal workflows |
| `json` | Structured output for programmatic processing and integrations |
| `sarif` | Industry-standard format for security platforms and CI/CD |

**Environment Controls**

```bash
NO_COLOR=1 mavetis review --staged    # Disable color output
FORCE_COLOR=1 mavetis review --staged # Force color output
```

---

<a id="hooks"></a>
## Hooks

```bash
mavetis hooks install
```

Configures:

- **Pre-commit**: `mavetis review --staged --fail-on high`
- **Pre-push**: `mavetis review --base <default-branch> --fail-on high`

Existing hooks are automatically backed up (`.bak`) prior to modification.

---

<a id="updates"></a>
## Updates

```bash
mavetis update          # Download and install latest release
mavetis update --check  # Check for available updates
```

The update process queries GitHub releases, verifies cryptographic checksums, downloads the platform-appropriate archive, and performs atomic binary replacement.

---

<a id="development"></a>
## Development

```bash
go build -o mavetis .     # Build from source
go test ./...             # Run tests
```

---

## License

Apache License 2.0 · Copyright 2026 Pimatis

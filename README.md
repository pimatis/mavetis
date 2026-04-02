# Mavetis

Mavetis is an enterprise-grade security analysis tool for Git-based development workflows. Designed for organizations requiring rigorous code security review, it performs comprehensive static analysis on code changes with complete network isolation.

## Overview

Mavetis delivers security analysis capabilities through the following core features:

- **Air-Gapped Operation**: Complete offline analysis with zero external network dependencies
- **Change-Focused Analysis**: Precise security evaluation of staged changes, branch diffs, and merge candidates
- **Comprehensive Detection**: Coverage across secrets management, authentication, authorization, cryptography, injection vulnerabilities, and supply chain security
- **Regression Prevention**: Detection of removed security controls, validation mechanisms, and policy enforcement points
- **Policy-Aware Review**: Built-in review profiles and trust zones for risk-weighted enterprise diff analysis
- **Boundary Enforcement**: Diff-level architectural boundary checks for privileged modules, admin surfaces, and UI/server trust edges
- **Flexible Rule Engine**: Customizable YAML-based rules with contextual scoping, typed policies, and repository snapshots
- **Enterprise Integration**: Native JSON and SARIF output formats for seamless CI/CD pipeline integration
- **Deterministic Execution**: Pure Go standard library implementation ensuring consistent, reproducible results

## Delivered Capability Layers

### Regression Core

Mavetis treats security weakening as a first-class signal in Git diffs, not only newly introduced dangerous code.

Delivered capabilities include:

- **Security Downgrade Detection**: SameSite weakening, cookie and token lifetime growth, bcrypt cost reduction, rate-limit threshold increases, timeout expansion, and MFA weakening
- **Config Drift Detection**: Debug mode activation, non-production environment fallbacks, wildcard CORS, weakened CSP, legacy TLS configuration, and privileged container settings
- **Observability Leak Detection**: Request body logging, authorization material leakage, PII in telemetry, raw error serialization, and sensitive tracing attributes

### Policy Layer

Mavetis can now operate as a policy-aware diff review layer instead of a single global rule set.

Delivered capabilities include:

- **Rule Profiles**: `auth`, `fintech`, `backend`, and `frontend` review modes for different engineering surfaces
- **Trust Zones**: `zones.critical` and `zones.restricted` path groups that automatically raise severity and tighten blocking thresholds
- **Policy Metadata in Output**: Text, JSON, and SARIF outputs now include active policy context for downstream review and CI enforcement

### Boundary Enforcement and Typed Rules

Mavetis can enforce architectural trust boundaries directly in changed lines and can express policy beyond regex-only matching.

Delivered capabilities include:

- **Permission Boundary Rules**: Built-in detection for public routes importing internal admin code, UI layers importing auth or security helpers, and public surfaces reaching privileged services
- **Typed Custom Rule DSL**: `forbiddenImport`, `deletedLineGuard`, `forbiddenEnv`, `requiredMiddleware`, `requiredCall`, `configKeyConstraint`, and `pathBoundary`
- **Diff-Bounded Evaluation**: Typed policies stay local to changed files and changed hunks, avoiding full repository graph traversal in the review path

### Supply-Chain Trust

Mavetis now treats dependency trust changes as policy events rather than simple package changes.

Delivered capabilities include:

- **Lifecycle and Dependency Correlation**: Alerts when dependency additions and install-time scripts land in the same branch
- **Registry Trust Drift Detection**: Detects private-to-public registry moves and registries outside the configured trust allowlist
- **Package Trust Policies**: Supports `supply.allow-packages`, `supply.deny-packages`, and `supply.trusted-registries`
- **Lockfile Consistency Checks**: Flags manifest changes that land without lockfile review in the same branch

### Security Intent and Repository Snapshots

Mavetis can now detect when security-named code stops behaving like security code and can preserve repository-specific secure baselines.

Delivered capabilities include:

- **Security Intent Mismatch Detection**: Flags security-named functions whose changed logic no longer reflects their declared protective purpose
- **Repository Security Snapshots**: `rules snapshot` can generate opt-in snapshot baselines and `snapshot.path` can enforce them during review
- **Diff-Local Baseline Enforcement**: Snapshot checks stay bounded to changed hunks and trigger only when the repository-specific baseline is weakened

## Security Architecture

Mavetis operates under a zero-trust security model with the following guarantees:

- No third-party dependencies or external modules
- Zero cloud service dependencies
- No authentication or account requirements
- No telemetry or usage data collection
- No artificial intelligence or machine learning components
- Complete network isolation during analysis
- Automatic redaction of sensitive data in findings
- Cryptographically verified installation packages

## Installation

### macOS and Linux

```bash
curl -fsSL https://raw.githubusercontent.com/pimatis/mavetis/main/install.sh | sh
```

### Windows PowerShell

```powershell
iwr https://raw.githubusercontent.com/pimatis/mavetis/main/install.ps1 -UseBasicParsing | iex
```

### Go Install

```bash
go install github.com/pimatis/mavetis@latest
```

### Maintenance

```bash
mavetis update          # Update to latest version
mavetis update --check  # Check for available updates
```

### Removal

macOS and Linux:

```bash
sudo rm -f /usr/local/bin/mavetis
rm -f "$HOME/.local/bin/mavetis"
```

Windows PowerShell:

```powershell
Remove-Item "$HOME\AppData\Local\mavetis\bin\mavetis.exe" -Force
```

Remove the binary from the location that was used during installation.

## Usage

### Basic Operations

```bash
# Review staged authentication changes with profile-aware explanations
mavetis review --staged --path 'src/**' --profile auth --explain

# Compare backend changes against a base branch
mavetis review --base main --path 'src/**' --profile backend

# CI/CD integration with fintech-focused policy output
mavetis ci --base main --format json --profile fintech

# Install Git hooks for automated scanning
mavetis hooks install
```

## Command Reference

### Analysis Commands

- `mavetis review` — Analyze code changes with configurable scope, output, and rule profile selection
- `mavetis ci` — Optimized analysis for continuous integration environments with profile-aware policy evaluation

### Git Hook Management

- `mavetis hooks install` — Configure automated pre-commit and pre-push scanning
- `mavetis hooks uninstall` — Remove configured Git hooks

### Rule Management

- `mavetis rules validate` — Validate custom rule definitions
- `mavetis rules list` — Display available security rules
- `mavetis rules show` — Display detailed rule information
- `mavetis rules test` — Test rules against sample diffs
- `mavetis rules matrix` — Generate compliance coverage matrix
- `mavetis rules snapshot` — Generate repository security snapshots from existing security-sensitive code anchors

### System Commands

- `mavetis update` — Self-update functionality
- `mavetis version` — Display version information

## Detection Capabilities

### Secrets Management and Cryptography

Mavetis identifies exposure of sensitive credentials and cryptographic weaknesses:

- Cloud provider credentials (AWS, Stripe, Supabase)
- Configuration file secrets (dotenv, JWT)
- Private key exposure and high-entropy secret patterns
- Cryptographic implementation flaws (weak randomness, weak hashing, weak ciphers)
- IV/nonce misuse and reuse patterns
- Verification bypass mechanisms and key confusion attacks
- Insecure algorithm selection and remote key retrieval

### Access Control and Session Management

Comprehensive analysis of authentication and authorization mechanisms:

- Authentication bypass vulnerabilities
- Deleted or disabled authentication middleware
- Insecure token storage implementations
- Session fixation and invalidation issues
- Timeout control removal
- Refresh token rotation failures
- Ownership verification deletion
- Authorization scope filter removal
- Operation-level permission regressions
- Insecure Direct Object Reference (IDOR) patterns
- JWT security flaws (decode-without-verify, missing binding, incomplete validation)
- OAuth implementation weaknesses (state, PKCE, nonce, replay attacks)

### Injection and Input Validation

Detection of injection vulnerabilities and unsafe data handling:

- Server-Side Request Forgery (SSRF)
- SQL injection
- Command injection
- Cross-Site Scripting (XSS)
- Unsafe deserialization
- Path traversal and Zip Slip vulnerabilities
- File upload validation gaps
- Cross-Origin Resource Sharing (CORS) misconfigurations
- TLS validation disablement
- Sensitive data logging
- Stack trace information disclosure
- Dynamic code evaluation (eval)
- Server-Side Template Injection (SSTI)
- Data flow analysis from request sources to sensitive sinks

### Supply Chain Security

Analysis of dependency and build pipeline security:

- Remote and git-based dependencies
- Version pinning violations and floating versions
- Public registry drift detection
- Remote replacement injections
- Typosquatting attack patterns
- Lockfile integrity violations
- Integrity hash removal
- Install-time script execution
- Direct shell download execution
- Mutable GitHub Action references
- Overly permissive workflow permissions
- `pull_request_target` misconfigurations
- Dependency and lifecycle-script correlation
- Registry trust allowlist enforcement
- Package allowlist and denylist enforcement
- Manifest-without-lockfile drift detection

## Regression Detection

Mavetis implements comprehensive regression detection by analyzing removed or weakened security controls with the same priority as newly introduced vulnerabilities. The system identifies:

- Deleted authentication and authorization middleware
- Removed role-based access control checks
- Eliminated ownership verification mechanisms
- Deleted input validation and sanitization routines
- Removed timeout and rate limiting controls
- Disabled token single-use validations
- Deleted file upload validation
- Removed scope and permission filters
- SameSite policy weakening and cookie lifetime expansion
- bcrypt cost downgrades and MFA requirement weakening
- Config drift that disables production-grade browser, transport, or deployment protections
- Architectural boundary violations across public, admin, UI, and privileged layers
- Security-intent regressions in validation, sanitization, ownership, MFA, and token functions
- Repository-specific snapshot regressions for security baselines captured from current code

## Output Formats

Mavetis supports multiple output formats for integration with various toolchains:

### Interactive Text (`text`)
Human-readable output with ANSI color coding for terminal environments. Suitable for developer workflows and manual review processes.

### Machine-Readable JSON (`json`)
Structured output format for programmatic processing and custom integrations.

### SARIF (`sarif`)
Industry-standard Static Analysis Results Interchange Format for integration with security platforms and CI/CD systems.

### Environment Controls

```bash
NO_COLOR=1 mavetis review --staged    # Disable color output
FORCE_COLOR=1 mavetis review --staged # Force color output
```

## Configuration

Mavetis loads configuration from `.mavetis.yaml` or `.mavetis.yml` in the current working directory.

### Base Configuration

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
zones:
  critical:
    - src/auth/**
    - src/lib/security/**
    - src/api/admin/**
  restricted:
    - src/payments/**
    - src/backoffice/**
```

### Profiles and Trust Zones

Profiles change which built-in rule families are active during review:

- `auth` — authentication, authorization, session, token, crypto, and related auth telemetry coverage
- `fintech` — full default policy surface for high-assurance review workflows
- `backend` — server-side security, supply-chain, config, network, and abuse-prevention coverage
- `frontend` — browser-facing auth, session, XSS, CORS, privacy, telemetry, and client config coverage

Trust zones raise enforcement in sensitive directories:

- `zones.critical` — raises matched findings by two severity levels up to `critical` and tightens blocking to `fail-on=low`
- `zones.restricted` — raises matched findings by one severity level and tightens blocking to `fail-on=medium`

Example:

```bash
mavetis review --staged --profile auth --config .mavetis.yaml
```

Output formats include the active policy, matched zone, base severity, and effective fail threshold so CI systems can distinguish policy escalation from the original detector severity.

### Custom Security Rules

Organizations can define organization-specific security policies through custom YAML rules.

Regex-based example:

```yaml
rules:
  - id: company.fetch.untrusted
    title: Untrusted Fetch Operation
    message: Request-controlled URL reached a sensitive fetch sink.
    remediation: Validate and allowlist outbound destinations before processing the request.
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

Typed policy examples:

```yaml
rules:
  - id: company.ui.auth-boundary
    type: forbiddenImport
    title: UI cannot import server auth helpers
    message: UI code imported a privileged auth helper.
    remediation: Move the logic behind a reviewed server boundary.
    category: boundary
    severity: high
    confidence: high
    target: added
    paths:
      - src/ui/**
      - src/components/**
    imports:
      - '(?i)(^|/)(auth|security|internal)(/|$)'

  - id: company.routes.require-auth
    type: requiredMiddleware
    title: Protected routes must include auth middleware
    message: A route was added without the required auth middleware.
    remediation: Attach the approved middleware before exposing the route.
    category: boundary
    severity: critical
    confidence: high
    target: added
    require:
      - 'router\.(get|post|put|delete)'
    middleware:
      - 'requireAuth'

  - id: company.prod-mode
    type: configKeyConstraint
    title: Runtime mode must stay production
    message: Runtime mode drifted outside the approved production value.
    remediation: Keep deployable runtime configuration pinned to production.
    category: config
    severity: high
    confidence: high
    target: added
    key: NODE_ENV
    allowed-values:
      - production
```

### Rule Matchers and Typed Policy Surface

Custom rules support both regex matchers and typed policy primitives:

- `require` — Mandatory pattern presence
- `any` — Alternative pattern matching
- `near` — Contextual proximity matching
- `absent` — Negative pattern matching
- `forbiddenImport` — Block imports from forbidden modules or trust zones
- `deletedLineGuard` — Treat deleted security guard lines as policy violations
- `forbiddenEnv` — Forbid risky environment keys and values
- `requiredMiddleware` — Enforce required middleware around route additions
- `requiredCall` — Enforce critical side-effect calls such as audit logging or authorization helpers
- `configKeyConstraint` — Constrain deployable config keys by allowed values, denied values, pattern, or numeric ranges
- `pathBoundary` — Express source-path to target-path trust boundaries directly
- Aliases: `forbidden`, `protected`, `required`, `context`, `mitigate`
- Path scoping with `paths`, `from-paths`, and `ignore`
- Compliance matrix generation via `mavetis rules matrix`

All regular expressions are compiled during initialization with immediate validation feedback.

### Repository Security Snapshots

Generate repository-specific security baselines from the current codebase:

```bash
mavetis rules snapshot --output .mavetis-snapshots.yaml --path 'src/auth/**'
```

Then enable snapshot enforcement in config:

```yaml
snapshot:
  path: .mavetis-snapshots.yaml
```

Snapshots are opt-in, local-only, and enforced only when a changed hunk weakens the captured baseline behavior.

## Git Hook Integration

Automated security scanning can be integrated into Git workflows:

```bash
mavetis hooks install
```

This configures:

- **Pre-commit**: `mavetis review --staged --fail-on high`
- **Pre-push**: `mavetis review --base <default-branch> --fail-on high`

Existing hook configurations are automatically backed up (`.bak`) prior to modification.

## Automated Updates

Mavetis includes secure self-update capabilities:

```bash
mavetis update          # Download and install latest release
mavetis update --check  # Check for available updates
```

The update process:

1. Queries GitHub releases for the latest version
2. Verifies cryptographic checksums
3. Downloads platform-appropriate archive
4. Performs atomic binary replacement

**Platform-Specific Behavior:**

- **macOS/Linux**: Attempts direct replacement; escalates via `sudo` when directory permissions require elevation
- **Windows**: Schedules replacement post-process exit; may require elevated terminal for protected directories

## Development

### Building from Source

Mavetis builds from the repository root entrypoint.

```bash
go build -o mavetis .
```

### Running Tests

```bash
go test ./...
```

## License

This project is licensed under the Apache License 2.0.

Copyright 2026 Pimatis.

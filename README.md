# Mavetis

Mavetis is an enterprise-grade security analysis tool for Git-based development workflows. Designed for organizations requiring rigorous code security review, it performs comprehensive static analysis on code changes with complete network isolation.

## Overview

Mavetis delivers security analysis capabilities through the following core features:

- **Air-Gapped Operation**: Complete offline analysis with zero external network dependencies
- **Change-Focused Analysis**: Precise security evaluation of staged changes, branch diffs, and merge candidates
- **Comprehensive Detection**: Coverage across secrets management, authentication, authorization, cryptography, injection vulnerabilities, and supply chain security
- **Regression Prevention**: Detection of removed security controls, validation mechanisms, and policy enforcement points
- **Flexible Rule Engine**: Customizable YAML-based rules with contextual scoping and intelligent suppression
- **Enterprise Integration**: Native JSON and SARIF output formats for seamless CI/CD pipeline integration
- **Deterministic Execution**: Pure Go standard library implementation ensuring consistent, reproducible results

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
# Review staged changes with detailed explanations
mavetis review --staged --path 'src/**' --explain

# Compare against base branch
mavetis review --base main --path 'src/**'

# CI/CD integration with JSON output
mavetis ci --base main --format json

# Install Git hooks for automated scanning
mavetis hooks install
```

## Command Reference

### Analysis Commands

- `mavetis review` — Analyze code changes with configurable scope and output
- `mavetis ci` — Optimized analysis for continuous integration environments

### Git Hook Management

- `mavetis hooks install` — Configure automated pre-commit and pre-push scanning
- `mavetis hooks uninstall` — Remove configured Git hooks

### Rule Management

- `mavetis rules validate` — Validate custom rule definitions
- `mavetis rules list` — Display available security rules
- `mavetis rules show` — Display detailed rule information
- `mavetis rules test` — Test rules against sample diffs
- `mavetis rules matrix` — Generate compliance coverage matrix

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

## Regression Detection

Mavetis implements comprehensive regression detection by analyzing removed security controls with the same priority as newly introduced vulnerabilities. The system identifies:

- Deleted authentication and authorization middleware
- Removed role-based access control checks
- Eliminated ownership verification mechanisms
- Deleted input validation and sanitization routines
- Removed timeout and rate limiting controls
- Disabled token single-use validations
- Deleted file upload validation
- Removed scope and permission filters

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
```

### Custom Security Rules

Organizations can define organization-specific security policies through custom YAML rules:

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

### Rule Matchers

Custom rules support comprehensive matching capabilities:

- `require` — Mandatory pattern presence
- `any` — Alternative pattern matching
- `near` — Contextual proximity matching
- `absent` — Negative pattern matching
- Aliases: `forbidden`, `protected`, `required`, `context`, `mitigate`
- Path scoping with `paths` and `ignore`
- Compliance matrix generation via `mavetis rules matrix`

All regular expressions are compiled during initialization with immediate validation feedback.

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

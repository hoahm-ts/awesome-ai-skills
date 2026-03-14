# awesome-ai-skills

[![codecov](https://codecov.io/gh/hoahm-ts/awesome-ai-skills/graph/badge.svg)](https://codecov.io/gh/hoahm-ts/awesome-ai-skills)

A curated collection of configuration files, instructions, and best practices for working with AI coding agents. This repository is a personal journey into the world of AI agents, where I share my experiences, insights, and lessons learned while exploring various AI skills and technologies.

## Table of Contents

- [Directory Structure](#directory-structure)
- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Repository Initialisation](docs/repository-initialisation.md)
- [Contributing](#contributing)
- [License](#license)

## Directory Structure

```
repo root/
├── AGENTS.md                        # Unified instructions for AI coding agents & contributors
├── CLAUDE.md                        # Claude-specific coding instructions
├── CODEX.md                         # Codex/ChatGPT-specific coding instructions
├── JUNIE.md                         # Junie (JetBrains AI) specific instructions
├── .cursorrules                     # Cursor AI editor rules
├── docs/                            # Architecture and design documentation
│   ├── architecture.md              # High-level system architecture overview
│   └── repository-initialisation.md # Step-by-step guide to initialise a new repository
├── openspec/                        # Spec-driven workflow artifacts
│   └── config.yaml
├── .claude/
│   ├── settings.json                # Claude project settings
│   ├── hooks/                       # Claude lifecycle hooks
│   ├── commands/                    # Claude custom slash commands
│   └── skills/                      # Claude reusable skills
├── .github/
│   ├── copilot-instructions.md      # GitHub Copilot instructions
│   ├── init-labels.sh               # Script to create standard GitHub issue labels
│   ├── init-repo-settings.sh        # Script to apply standard repository settings
│   ├── labeler.yml                  # Label-to-file-pattern mappings for the labeler workflow
│   ├── rulesets/
│   │   ├── protect-main.json        # Ruleset: protect the default branch
│   │   └── general-rule.json        # Ruleset: general rules for all branches
│   ├── release.yaml                 # Release drafter configuration
│   ├── PULL_REQUEST_TEMPLATE.md     # Default pull request template
│   └── workflows/
│       ├── ci.yml                   # CI workflow: lint, test, and coverage reporting
│       └── labeler.yml              # GitHub Actions workflow to auto-label pull requests
└── src/                             # Go module root
    ├── go.mod
    ├── go.sum
    ├── Makefile
    ├── .golangci.yml                # golangci-lint configuration
    ├── api/                         # OpenAPI specs + codegen config/inputs
    │   └── openapi.yaml
    ├── cmd/                         # Application binaries
    │   ├── api/                     # HTTP API server
    │   ├── worker/                  # Temporal background worker
    │   └── migration/               # Database migration runner
    ├── internal/                    # Application core (not importable outside this module)
    │   ├── handler/                 # HTTP route handlers (delivery layer)
    │   ├── integration/             # External service adapters (infra layer)
    │   ├── lifecycle/               # Startup and graceful shutdown
    │   ├── timeline/                # Temporal workflow and activity definitions
    │   ├── event/                   # Kafka event publishers and consumers
    │   ├── shared/                  # Shared kernel: types, sentinel errors, pagination
    │   └── wire/                    # Single composition root — all DI wiring lives here
    ├── pkg/                         # Reusable helper libraries (no business logic)
    │   ├── config/                  # Application configuration loading
    │   ├── logger/                  # zerolog setup and context helpers
    │   ├── middleware/              # HTTP middleware
    │   ├── response/                # Standard JSON HTTP response helpers
    │   ├── utils/                   # Small general-purpose helpers
    │   └── security/                # Password hashing and verification
    ├── migrations/                  # SQL migration files (golang-migrate)
    ├── docker/                      # Dockerfile(s)
    ├── etc/config/                  # Environment-specific config files
    └── scripts/                     # Build and operational scripts
```

## Overview

This repository provides a unified set of instructions and configuration files for the most popular AI coding agents and tools. Each file targets a specific agent or editor, while `AGENTS.md` serves as the canonical source of truth shared across all of them.

| File / Directory | AI Agent / Tool |
|---|---|
| `AGENTS.md` | All agents (canonical reference) |
| `CLAUDE.md` | [Claude](https://claude.ai) by Anthropic |
| `CODEX.md` | [Codex](https://openai.com/blog/openai-codex) / ChatGPT by OpenAI |
| `JUNIE.md` | [Junie](https://www.jetbrains.com/junie/) by JetBrains |
| `.cursorrules` | [Cursor](https://www.cursor.com/) AI editor |
| `.claude/` | [Claude Code](https://docs.anthropic.com/en/docs/claude-code) settings & extensions |
| `.github/copilot-instructions.md` | [GitHub Copilot](https://github.com/features/copilot) |

## Tech Stack

| Category | Technology |
|---|---|
| Language | [Go](https://go.dev/) 1.24+ |
| HTTP framework | [chi](https://github.com/go-chi/chi) |
| Database | [PostgreSQL](https://www.postgresql.org/) |
| Cache | [Redis](https://redis.io/) |
| Messaging | [Kafka](https://kafka.apache.org/) |
| Workflow engine | [Temporal](https://temporal.io/) |
| Tracing & Observability | [Datadog](https://www.datadoghq.com/) ([OpenTelemetry](https://opentelemetry.io/)) |
| Dependency injection | [Google Wire](https://github.com/google/wire) |
| Logging | [zerolog](https://github.com/rs/zerolog) |

## Getting Started

1. Clone this repository as a reference or template.
2. Copy the relevant configuration files into your own project.
3. Customise the instructions to match your project's conventions, tech stack, and coding standards.

## Repository Initialisation

For a step-by-step guide on setting up a new repository with the team's standard configuration, see [docs/repository-initialisation.md](docs/repository-initialisation.md).

---

## Contributing

Contributions, improvements, and new agent configurations are welcome. Please read `AGENTS.md` for contribution guidelines before submitting a pull request.

## License

MIT © Hoa Hoang — see [LICENSE](LICENSE) for details.

# awesome-ai-skills

[![CI](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/hoahm-ts/awesome-ai-skills/badges/.github/badges/coverage.json)](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml)

A curated collection of configuration files, instructions, and best practices for working with AI coding agents. This repository is a personal journey into the world of AI agents, where I share my experiences, insights, and lessons learned while exploring various AI skills and technologies.

## Table of Contents

- [Directory Structure](#directory-structure)
- [Overview](#overview)
- [Tech Stack](docs/tech-stacks.md)
- [Architecture Decision Records](docs/adr/README.md)
- [High-Level Design Documents](docs/hld/README.md)
- [Getting Started](#getting-started)
- [Repository Initialisation](#repository-initialisation)
- [Using Claude Code Skills](docs/claude-code-skills.md)
- [AI-Augmented SDLC Guide](docs/ai-sdlc-guide.md)
- [Claude Desktop MCP Integration & Custom Skills Guide](docs/claude-desktop-mcp-guide.md)
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
│   ├── ai-sdlc-guide.md             # AI-augmented SDLC: environment setup and development process
│   ├── repository-initialisation.md # Step-by-step guide to initialise a new repository
│   ├── claude-desktop-mcp-guide.md  # Guide: connect Claude Desktop to MCP tools & create custom skills
│   ├── tech-stacks.md               # Tech stack reference
│   ├── claude-code-skills.md        # Using Claude Code Skills: available skills and usage guide
│   ├── adr/                         # Architecture Decision Records (ADR index + template)
│   │   ├── README.md                # ADR index with summary table and changelog
│   │   └── ADR-00-template.md       # Reusable template for writing new ADRs
│   ├── hld/                         # High-Level Design Documents (HLD index + template)
│   │   ├── README.md                # HLD index with summary table and changelog
│   │   └── HLD-00-template.md       # Reusable template for writing new HLDs
│   └── conventions/                 # Stack-specific and cross-cutting coding conventions
│       ├── diagram-conventions.md   # Diagram format, type selection, and usage guidelines
│       ├── env-conventions.md       # Environment variable and Docker conventions
│       ├── golang-conventions.md    # Go style, testing, and architecture conventions
│       ├── kafka-conventions.md     # Kafka producer/consumer conventions
│       ├── logging-conventions.md   # Logging, errors, and observability conventions
│       ├── postgresql-conventions.md # PostgreSQL and GORM conventions
│       ├── python-conventions.md    # Python style, testing, and architecture conventions
│       ├── redis-conventions.md     # Redis usage conventions
│       ├── restful-conventions.md   # RESTful API conventions
│       └── temporal-conventions.md  # Temporal workflow conventions
├── openspec/                        # Spec-driven workflow artifacts
│   └── config.yaml
├── .claude/
│   ├── settings.json                # Claude project settings
│   ├── hooks/                       # Claude lifecycle hooks
│   ├── commands/                    # Claude custom slash commands
│   │   ├── eng-team-analysis.md     # Engineering team analysis command
│   │   ├── estimate-release.md      # Per-ticket release estimation command
│   │   ├── project-status-summary.md # Project status summary command
│   │   └── weekly-bottleneck-report.md # Weekly sprint bottleneck report command
│   └── skills/                      # Claude reusable skills
│       ├── eng-team-analysis/       # Engineering team analysis skill
│       ├── project-status-summary/  # Project status summary skill
│       └── weekly-bottleneck-report/ # Weekly bottleneck report skill
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
│       ├── docker.yml               # Docker workflow: multi-arch image build and push via buildx
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

For the full technology reference, see [docs/tech-stacks.md](docs/tech-stacks.md).

## Getting Started

1. Clone this repository as a reference or template.
2. Copy the relevant configuration files into your own project.
3. Customise the instructions to match your project's conventions, tech stack, and coding standards.
4. Follow the [AI-Augmented SDLC Guide](docs/ai-sdlc-guide.md) to set up your development environment and start using AI tools across the full development lifecycle.

## Repository Initialisation

> If you are cloning this repository as a scaffold for a new project, follow the step-by-step guide in [docs/repository-initialisation.md](docs/repository-initialisation.md) to apply the team's standard configuration to your new repository.

## Using Claude Code Skills

For the full guide on available skills and how to use them, see [docs/claude-code-skills.md](docs/claude-code-skills.md).

---

## Contributing

Contributions, improvements, and new agent configurations are welcome. Please read `AGENTS.md` for contribution guidelines before submitting a pull request.

## License

MIT © Hoa Hoang — see [LICENSE](LICENSE) for details.

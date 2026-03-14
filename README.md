# awesome-ai-skills

[![CI](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/hoahm-ts/awesome-ai-skills/badges/.github/badges/coverage.json)](https://github.com/hoahm-ts/awesome-ai-skills/actions/workflows/ci.yml)

A curated collection of configuration files, instructions, and best practices for working with AI coding agents. This repository is a personal journey into the world of AI agents, where I share my experiences, insights, and lessons learned while exploring various AI skills and technologies.

## Table of Contents

- [Directory Structure](#directory-structure)
- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Using Claude Code Skills](#using-claude-code-skills)
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

## Using Claude Code Skills

Claude Code skills live in `.claude/skills/`. Each skill is a self-contained Markdown file that gives Claude a reusable, well-defined behaviour — like a saved workflow you can invoke by name.

### Available skills

| Skill | Slash command | Description |
|-------|---------------|-------------|
| `project-status-summary` | `/project-status-summary` | Aggregate project health signals from Slack, Jira, Confluence, and email into a structured executive summary, then post it to Slack |
| `openspec-explore` | `/opsx:explore` | Enter explore mode — a thinking partner for ideas, problems, and requirements |
| `openspec-propose` | `/opsx:propose` | Propose a new change and generate all artifacts (proposal, design, tasks) in one step |
| `openspec-apply-change` | `/opsx:apply` | Implement tasks from an existing OpenSpec change |
| `openspec-archive-change` | `/opsx:archive` | Archive a completed change |

### How to use the `project-status-summary` skill

**Option A — natural language**

Ask Claude in plain English:

```
Please summarize the current project status for the project: "tpbank", on the CO board.
```

To look back further than the default 7 days:

```
Please summarize the project "tpbank" on the CO board for the last 14 days.
```

**Option B — slash command**

```
/project-status-summary tpbank CO                     # last 7 days, send to your DM
/project-status-summary tpbank CO 14                  # last 14 days
/project-status-summary tpbank CO 7 #tpbank-updates   # post to a channel
/project-status-summary tpbank CO 7 @john             # DM to @john
/project-status-summary "my project" CO               # multi-word project name
```

**Parameters**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `project` | ✅ | — | Project name (use quotes for multi-word names) |
| `board` | ✅ | — | Jira board key (e.g., `CO`) |
| `days` | ❌ | `7` | Number of days to look back |
| `channel` | ❌ | user's own DM | Slack destination — a channel (`#name`) or DM handle (`@name`) |

**What it does**

1. Searches Slack (channels and messages), Jira (board tickets and sprint data), Confluence (updated pages), and email for activity matching the project name within the specified time window.
2. Produces a structured executive summary with numbered source references on every bullet.
3. Lists all sources in a table with direct links (Slack message permalinks, Jira ticket URLs, Confluence page URLs, email subject lines).
4. Sends the formatted summary to your Slack DM (or a channel you specify).

**Prerequisites**

The skill requires MCP integrations for the sources you want to search. Configure the relevant integrations in your Claude Code session before invoking the skill:
- **Slack** — [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack)
- **Jira** — [Atlassian Jira MCP server](https://github.com/sooperset/mcp-atlassian)
- **Confluence** — [Atlassian Confluence MCP server](https://github.com/sooperset/mcp-atlassian) (same server as Jira)
- **Email** — [Gmail MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/gmail) or [Outlook MCP server](https://github.com/modelcontextprotocol/servers)

If a source is not connected, the skill notes it clearly and continues with the remaining sources.

## Repository Initialisation

For a step-by-step guide on setting up a new repository with the team's standard configuration, see [docs/repository-initialisation.md](docs/repository-initialisation.md).

---

## Contributing

Contributions, improvements, and new agent configurations are welcome. Please read `AGENTS.md` for contribution guidelines before submitting a pull request.

## License

MIT © Hoa Hoang — see [LICENSE](LICENSE) for details.

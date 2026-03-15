# AGENTS.md — Unified Instructions for AI Coding Agents & Contributors

This file is the **canonical reference** for all AI coding agents and human contributors working in this repository. All agent-specific instruction files (`CLAUDE.md`, `CODEX.md`, `JUNIE.md`, `.cursorrules`, `.github/copilot-instructions.md`) should be kept consistent with the guidelines here.

## Stack-Specific Conventions

> **Before writing or reviewing code in any of these stacks, read the corresponding convention file first.**

| Stack | Convention File |
|---|---|
| Go (style, testing, architecture) | [`docs/conventions/golang-conventions.md`](docs/conventions/golang-conventions.md) |
| Logging, Errors & Observability | [`docs/conventions/logging-conventions.md`](docs/conventions/logging-conventions.md) |
| PostgreSQL & GORM | [`docs/conventions/postgresql-conventions.md`](docs/conventions/postgresql-conventions.md) |
| Kafka | [`docs/conventions/kafka-conventions.md`](docs/conventions/kafka-conventions.md) |
| Temporal workflows | [`docs/conventions/temporal-conventions.md`](docs/conventions/temporal-conventions.md) |
| Redis | [`docs/conventions/redis-conventions.md`](docs/conventions/redis-conventions.md) |
| RESTful APIs | [`docs/conventions/restful-conventions.md`](docs/conventions/restful-conventions.md) |
| Environment variables & Docker | [`docs/conventions/env-conventions.md`](docs/conventions/env-conventions.md) |

---

## Table of Contents

- [Project Overview](#project-overview)
- [Core Principles](#core-principles)
- [Security & Data Handling](#security--data-handling)
- [Documentation Requirements](#documentation-requirements)
- [File & Directory Conventions](#file--directory-conventions)
- [Contribution Guidelines](#contribution-guidelines)
- [Git Workflow & Pull Request Guidelines](#git-workflow--pull-request-guidelines)
- [AI Agent Behaviour](#ai-agent-behaviour)

---

## Project Overview

`awesome-ai-skills` is a curated collection of configuration files, instructions, and best practices for working with AI coding agents. The goal is to provide a ready-to-use set of files that can be dropped into any project to give AI tools the context they need to be maximally helpful.

---

## Core Principles

These principles are non-negotiable and apply to every change made in this repository.

- **Modular design**: business logic lives in well-defined, bounded modules. Entry points (CLI, HTTP, workers) are thin wrappers that wire dependencies and delegate to the core.
- **One composition root**: dependency wiring happens in a single, explicit place. Avoid hidden global singletons.
- **Edges vs core**:
  - *Edges*: delivery/transport layers (HTTP handlers, CLI commands) and external integrations (third-party clients, adapters).
  - *Core*: domain modules containing business rules and services.
- **Prefer interfaces/ports** when crossing module boundaries. Do not import concrete implementations from one domain into another.
- **Explicit over implicit**: no magic numbers, no unclear abbreviations, no unexplained side effects.

---

## Security & Data Handling

- Treat user and customer data as sensitive by default.
- Never commit secrets; use a secrets/config management solution.
- Validate and authorise in the handler layer; enforce invariants in domain services where appropriate.
- Keep auth and permission logic centralised rather than scattered across handlers.

---

## Documentation Requirements

When adding or altering a capability:

- Update `docs/` for changes that affect system design or architecture.
- Update the API spec if the API contract changes.
- Add inline package/module documentation for non-obvious modules or decisions.
- Keep the `README.md` directory structure table up to date if new files are added.

---

## File & Directory Conventions

- Markdown files use [GitHub Flavored Markdown](https://github.github.com/gfm/).
- JSON files must be valid and formatted with 2-space indentation.
- Keep all AI configuration files at the repository root or inside their designated directories (`.claude/`, `.github/`).

---

## Contribution Guidelines

1. Fork the repository and create a feature branch from `main`.
2. Make the smallest possible changes that fully address the issue.
3. Verify your changes before opening a pull request.
4. Ensure the `README.md` directory structure table stays up to date if new files are added.
5. Open a pull request with a clear title and description (see [Git Workflow & Pull Request Guidelines](#git-workflow--pull-request-guidelines) below).

---

## Git Workflow & Pull Request Guidelines

### Branch Naming

Branch format: `<type>/<ticket>` — ticket format is `JIRA-<number>`. No description suffix. Examples: `spec/JIRA-1`, `feat/JIRA-1`, `fix/JIRA-1`.

### Creating a Pull Request

- **Title format:** `<TICKET_NUMBER>: <description>` — e.g. `JIRA-29: init the project structure`
- Keep PRs focused; avoid drive-by refactors unless they are directly necessary.
- If behaviour changes, state explicitly what the old and new behaviour are.

**Description:** The PR body **must use the exact section structure** from [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md). Do not invent custom headings — copy the template and fill in every section:

```markdown
## References
- **Jira:** <link>
- **Related:** #<issue>
- **Materials:** <link>

## Type of Change
- [x] `<type>` — <label>

## What changed?
<concise summary of what was added, modified, or removed>

## Why?
<motivation / problem statement>

## How did you test it?
<steps to validate locally or in CI>

## Potential risks
<what could break in production>

## Is hotfix candidate?
<Yes / No>
```

Leave no section blank or with placeholder text.

---

## AI Agent Behaviour

### Always do first

1. **Read this file** before making any changes.
2. **Read the relevant stack convention files** from [`docs/conventions/`](docs/conventions/) for every stack you will touch.
3. Identify the target entry point(s): HTTP handler, worker, CLI command, or migration.
4. Identify the domain module(s) impacted.
5. Identify boundary changes: HTTP contract, external integration, workflow/state, or schema.

### During implementation

- Prefer small, incremental diffs.
- Keep changes localised to one domain when possible.
- Introduce interfaces (ports) when crossing domains or calling integrations.
- Register new dependencies in the composition root / DI wiring.
- Follow existing patterns in adjacent code; do not invent new frameworks.
- Do not modify files unrelated to the task.
- If unsure about a convention, look for existing examples before inventing one.

### Before finishing

- Run existing linters and tests; ensure compile-level correctness at minimum.
- Confirm DI/wiring compiles cleanly if applicable.
- Do not commit secrets, credentials, or sensitive data.
- Follow the directory structure documented in `README.md`.
- Write the PR description using the **exact section structure** from [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) — fill in every section, leave nothing blank.
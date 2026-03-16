# AI-Augmented SDLC Guide

This guide explains how to set up your development environment and use AI tools to accelerate every phase of the software development lifecycle — from gathering requirements through to code review and archiving.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
  - [1. Claude Code CLI](#1-claude-code-cli)
  - [2. OpenSpec](#2-openspec)
  - [3. oapi-codegen](#3-oapi-codegen)
  - [4. MCP Integrations](#4-mcp-integrations)
  - [5. IDE Setup](#5-ide-setup)
- [Development Process](#development-process)
  - [Step 1 — Gather Specs](#step-1--gather-specs)
  - [Step 2 — Analyze & Design](#step-2--analyze--design)
  - [Step 3 — Implement](#step-3--implement)
  - [Step 4 — Code Review](#step-4--code-review)
  - [Step 5 — Address Feedback](#step-5--address-feedback)
  - [Step 6 — Archive](#step-6--archive)
- [Quick Reference](#quick-reference)

---

## Overview

This guide documents the team's AI-augmented SDLC. The workflow integrates Claude Code, OpenSpec, GitHub Copilot, Junie, and other tools to automate code generation, testing, review, and delivery — with humans remaining in control of requirements, acceptance, and architectural decisions.

```mermaid
flowchart TD
    Start([**New Feature / Bug / Task**]) --> S1

    subgraph S1["**Step 1 - Gather Specs**"]
        direction TB
        R1[Collect BRD, PRD, ADRs, Drive, Confluence]
        R2[Create Jira ticket & Link docs]
        R3["Create PR spec/{TICKET}"]
        R1 --> R2 --> R3
    end

    S1 --> ambig{Requirements<br>clear?}
    
    ambig -- No --> explore
    ambig -- Yes --> propose

    subgraph S2["**Step 2 — Analyze & Design**"]
        direction TB
        explore["/opsx:explore {TICKET}"]
        propose["/opsx:propose {TICKET}"]
        create_pr["gp spec/{TICKET}"]
        review_art["Review artifacts:<br>proposal.md, design.md, tasks.md"]
        
        explore --> propose
        propose --> create_pr
        create_pr --> review_art
    end

    review_art --> art_ok{Artifacts<br>**approved?**}
    art_ok -- Revise --> propose
    art_ok -- Approved --> S3

    subgraph S3["**Step 3 — Implement**"]
        direction TB
        apply["/opsx:apply {task name}"]
        codegen["oapi-codegen stubs"]
        create_feat_pr["gp feat/{TICKET}"]
        apply --> codegen --> create_feat_pr
    end

    S3 --> S4

    subgraph S4["**Step 4 - Code Review**"]
        direction TB
        copilot["AI Review (Copilot/TS)"]
        skills["/pr-review-toolkit:review-pr"]
        ci["CI checks (All Green)"]
        human_review["Human code review"]
        
        copilot --> skills --> ci --> human_review
    end

    human_review --> review_ok{"Approved &<br>CI Green?"}

    subgraph S5["**Step 5 - Address Feedback**"]
        direction TB
        fix_cmd["Claude Code: fix review PR#123"]
        copilot_agent["GitHub Copilot Agent"]
        codex_cursor["Codex · Cursor"]
        junie_fix["GoLand + Junie"]
        manual["Manual IDE Fix"]
        fix_cmd --- copilot_agent --- codex_cursor --- junie_fix --- manual
    end

    review_ok -- "Needs Changes" --> S5
    S5 --> S4
    
    review_ok -- Approved --> merge["Squash merge → main<br>Branch auto-deleted"]

    merge --> S6

    subgraph S6["**Step 6 — Archive**"]
        direction TB
        archive["/opsx:archive<br>Move to openspec/changes/"]
        jira_done["Close Jira ticket<br>Add PR link"]
        archive --> jira_done
    end

    S6 --> Done([Done])

    %% Styling
    style S1 fill:#e8f4fd,stroke:#2196F3
    style S2 fill:#f3e8fd,stroke:#9C27B0
    style S3 fill:#e8fdf0,stroke:#4CAF50
    style S4 fill:#fdf8e8,stroke:#FF9800
    style S5 fill:#fde8e8,stroke:#F44336
    style S6 fill:#f0fde8,stroke:#8BC34A
```

---

## Prerequisites

### 1. Claude Code CLI

Claude Code is the primary AI coding agent. Install it and connect it to your project.

**Install**

```bash
npm install -g @anthropic-ai/claude-code
```

**Verify**

```bash
claude --version
```

**Authenticate**

```bash
claude auth login
```

Follow the browser prompt to complete OAuth authentication with your Anthropic account.

**Configure the project**

Claude Code reads its project-level settings from `.claude/settings.json`. The file in this repository is pre-configured with broad permissions for local development:

```json
{
  "version": "1.0.0",
  "permissions": {
    "allow": ["Bash(*)", "Read(*)", "Write(*)", "Edit(*)"],
    "deny": []
  }
}
```

> [!WARNING]
> The default configuration grants Claude broad access (`Bash(*)`, `Read(*)`, `Write(*)`). This is intentional for local development but is a significant blast radius if a prompt is misinterpreted. Tighten the `deny` list — for example, deny `Bash(rm *)`, `Bash(git push *)`, or restrict `Write` to specific directories — before using this configuration in shared, staging, or production environments.

---

### 2. OpenSpec

OpenSpec is the spec-driven workflow tool that creates and manages change artifacts (proposal, design, tasks).

**Install**

```bash
npm install -g openspec
```

**Configure the project**

The repository ships with a pre-configured `openspec/config.yaml`. Review and update the `context` field to match your project name and tech stack before use.

---

### 3. oapi-codegen

`oapi-codegen` generates Go server stubs and client code from OpenAPI 3.x specifications.

**Install**

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

**Verify**

```bash
oapi-codegen --version
```

**Usage**

The OpenAPI specification lives at `src/api/openapi.yaml`. Run code generation via the Makefile target:

```bash
cd src && make generate
```

> If a `generate` target does not yet exist, add it to `src/Makefile` pointing to your `oapi-codegen` configuration file.

---

### 4. MCP Integrations

Claude Code connects to external tools via the Model Context Protocol (MCP). Configure all required integrations in your Claude Code session before starting any workflow that queries them.

#### Required integrations

| Integration | Purpose | Reference |
|---|---|---|
| **Jira** | Read and update tickets, sprint data, blockers | [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) |
| **Confluence** | Read and write design docs, ADRs, meeting notes | [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) (same server) |
| **GitHub** | Create commits, open PRs, trigger CI, read PR reviews | [github-mcp-server](https://github.com/github/github-mcp-server) |
| **Slack** | Send summaries and notifications | [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack) |
| **Google Drive** | Read BRDs, PRDs, and other shared documents | [Google Drive MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/gdrive) |
| **Google Calendar** | Read sprint ceremonies, release dates, and deadlines | [Google Calendar MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/gcal) |

#### Connecting an integration

Add each server to your Claude Code MCP configuration:

```bash
claude mcp add <server-name> <command> [args...]
```

Example for the Atlassian server:

```bash
claude mcp add atlassian npx mcp-atlassian \
  --jira-url https://your-org.atlassian.net \
  --confluence-url https://your-org.atlassian.net/wiki
```

Consult each server's README for authentication details (API tokens, OAuth apps, etc.).

#### Verifying integrations

List connected integrations:

```bash
claude mcp list
```

---

### 5. IDE Setup

#### GoLand + Junie

1. Install [GoLand](https://www.jetbrains.com/go/).
2. Install the [Junie plugin](https://www.jetbrains.com/junie/) from the JetBrains Marketplace.
3. Open the `src/` directory as your project root.
4. Configure the GOPATH and Go SDK inside **Settings → Go → GOROOT**.

Junie reads the project instructions from `JUNIE.md` at the repository root. Keep this file up to date.

#### Visual Studio Code + GitHub Copilot

1. Install [Visual Studio Code](https://code.visualstudio.com/).
2. Install the [GitHub Copilot extension](https://marketplace.visualstudio.com/items?itemName=GitHub.copilot).
3. Sign in with your GitHub account that has an active Copilot license.
4. Open the repository root as your workspace.

GitHub Copilot reads its project instructions from `.github/copilot-instructions.md`.

#### Optional tools

| Tool | Purpose |
|---|---|
| [Codex CLI](https://github.com/openai/codex) | OpenAI code generation agent (alternative to Claude Code) |
| [Cursor](https://www.cursor.com/) | AI-native editor using `.cursorrules` from this repository |

---

## Development Process

The full AI-augmented workflow maps to these steps:

```
1. Gather specs                      ← create Jira ticket, link docs, open spec PR
2. Analyze & design                  ← Claude Code /opsx:explore + /opsx:propose
3. Implement                         ← Claude Code /opsx:apply + oapi-codegen
4. Code review                       ← AI review (Copilot) + CI + human review
5. Address feedback                  ← any tool: Claude Code / Copilot Agent / Codex / Cursor / GoLand+Junie / Manual IDE
6. Archive                           ← Claude Code /opsx:archive
```

---

### Step 1 — Gather Specs

**Goal**: collect all available context for the feature or fix and open a spec branch before any AI session begins.

#### 1.1 — Collect source materials

Gather the following materials and save links or copies where your team stores shared documents:

| Material | Where to look |
|---|---|
| Business Requirements Document (BRD) | Google Drive, Confluence |
| Product Requirements Document (PRD) | Google Drive, Confluence |
| High-level design / system design | Confluence, architecture docs |
| Tech stack & ADRs | `docs/architecture.md`, Confluence |
| Internal API specifications | `src/api/openapi.yaml`, Confluence |
| External API specifications | vendor docs, shared Drive folder |

#### 1.2 — Create a Jira ticket

1. Open Jira and create a new ticket in the relevant project board.
2. Set the ticket type (Story, Task, Bug, etc.) and priority.
3. Paste the raw requirement in the **Description** field.
4. Link all relevant materials using the **Link** field or description:
   - Confluence pages (BRD, PRD, design docs)
   - Google Drive documents
   - Related Jira epics or parent stories
5. Assign the ticket to yourself and move it to **Backlog** or **In Progress** as appropriate.

**Ticket naming convention**: use a clear, imperative title that describes the outcome.
Example: `Add user authentication via OAuth2`.

> **Tip — create the ticket with AI**: Once your Atlassian MCP is configured (see [Section 4](#4-mcp-integrations)), you can have Claude create and link the ticket for you instead of doing it manually:
> ```
> Create a Jira Story in board DOP titled "Add OAuth2 login via Google identity provider".
> Set priority to High and link the PRD at <confluence-url>.
> ```

#### 1.3 — Create the spec branch and PR

Create a dedicated branch for the spec artifacts and open a draft PR:

```bash
git checkout -b spec/TICKET-123
git push -u origin spec/TICKET-123
```

Opening this PR early gives your team visibility that work has started. In the PR description, include the Jira ticket link using the following format in the **References** section of `.github/PULL_REQUEST_TEMPLATE.md`:

```
## References
- **Jira:** https://your-org.atlassian.net/browse/TICKET-123
```

The OpenSpec artifacts generated in Step 2 will be committed to this branch.

---

### Step 2 — Analyze & Design

**Goal**: use AI to produce a structured proposal, detailed design, and implementation task list from the gathered requirements.

#### 2.1 — (Optional) Explore first with `/opsx:explore`

If the requirements are ambiguous or you want to think through options before committing to a design, start an exploration session:

```
/opsx:explore
```

Claude will act as a thinking partner: asking clarifying questions, exploring trade-offs, and helping you refine the problem statement before design begins.

#### 2.2 — Propose the change with `/opsx:propose`

Open a Claude Code session and run:

```
/opsx:propose
```

Claude will:
1. Ask what you want to build (or infer it from context).
2. Derive a kebab-case change name (e.g., `add-oauth2-auth`).
3. Run `openspec new change "<name>"` to scaffold the change directory.
4. Generate artifacts in dependency order:
   - **`proposal.md`** — what and why (scope, goals, non-goals)
   - **`design.md`** — how (architecture, data model, API contract, component interactions)
   - **`tasks.md`** — numbered, time-boxed implementation tasks with acceptance criteria
5. Show a final status summary when all artifacts are ready.

**Tip**: Attach the Jira ticket URL and any relevant Confluence pages in your prompt so Claude has the full context:

```
/opsx:propose

Context:
- Jira ticket: https://your-org.atlassian.net/browse/AGI-123
- PRD: https://your-org.atlassian.net/wiki/...
- We want to add OAuth2 login using Google as the identity provider.
```

#### 2.3 — Review and approve the artifacts

Before implementation begins, review each artifact:

| Artifact | What to check |
|---|---|
| `proposal.md` | Scope is correct; non-goals are explicit; stakeholder sign-off if required |
| `design.md` | Architecture follows the layering rules in `AGENTS.md`; no forbidden dependencies |
| `tasks.md` | Tasks are small enough (≤ 2 hours each); acceptance criteria are clear and testable |

#### 2.4 — Push the spec PR with artifacts

Once you are satisfied with the generated artifacts, commit them to the spec branch and push:

```bash
git add openspec/
git commit -m "spec(TICKET-123): add proposal, design, and tasks"
git push origin spec/TICKET-123
```

The spec PR is now ready for team review before implementation begins.

---

### Step 3 — Implement

**Goal**: use AI agents to write the code, following the task list generated in Step 2.

#### 3.1 — Run `/opsx:apply`

```
/opsx:apply
```

Claude will:
1. Read the context files (`proposal.md`, `design.md`, `tasks.md`).
2. Display current progress (`N/M tasks complete`).
3. Implement each pending task in sequence, marking `- [ ]` → `- [x]` as each is completed.
4. Pause and ask for guidance if a task is unclear or an implementation issue arises.

**Tip**: keep the Claude Code session open throughout implementation. You can interrupt at any time by simply typing your question or instruction.

#### 3.2 — Generate API code with `oapi-codegen`

If the change adds or modifies API endpoints, regenerate the Go server stubs after updating `src/api/openapi.yaml`:

```bash
cd src && make generate
```

Commit the regenerated files alongside the hand-written implementation code.

> **Tip**: Use your IDE AI for focused, in-file assistance during implementation. GoLand + Junie and VS Code + GitHub Copilot both read the project instruction files (`JUNIE.md`, `.github/copilot-instructions.md`) to stay aligned with team conventions.

#### 3.3 — Push the feature PR

Once implementation is complete, commit and push to open the feature PR:

```bash
git add src/ openspec/
git commit -m "feat(auth): add OAuth2 login via Google identity provider"
git push origin feat/TICKET-123
```

Open the pull request on GitHub using the PR title format:

```
TICKET-123: Add OAuth2 login via Google identity provider
```

Fill in all sections of `.github/PULL_REQUEST_TEMPLATE.md`. The `.github/workflows/labeler.yml` workflow automatically applies labels based on the branch name and changed files.

---

### Step 4 — Code Review

**Goal**: validate the change is correct, safe, and follows team conventions before merging.

#### 4.1 — AI review with GitHub Copilot (automatic)

The `.github/rulesets/protect-main.json` ruleset configures GitHub to **automatically request a Copilot code review** on every new push and draft PR. Copilot will post inline comments and a summary review within a few minutes of the PR being opened.

Review the Copilot feedback and address any comments you agree with before requesting human review.

#### 4.2 — Claude Code review skills

Use the built-in review skills inside Claude Code to catch additional issues early:

```
/pr-review-toolkit:review-pr
```

Specialist sub-skills that run automatically as part of the toolkit, or can be invoked individually:

| Skill | What it catches |
|---|---|
| `pr-review-toolkit:silent-failure-hunter` | Silent failures, swallowed errors, inappropriate fallback behaviour |
| `pr-review-toolkit:code-simplifier` | Over-engineered code, unnecessary complexity |
| `pr-review-toolkit:code-reviewer` | Style violations, AGENTS.md guideline breaches |

#### 4.3 — CI checks

All CI checks must pass before the PR can be merged. The CI pipeline runs tests and linting automatically on every push:

```bash
# Run locally to pre-validate before pushing
cd src && go test ./...
cd src && golangci-lint run ./...
```

Fix any failures. If a failure reveals a design issue, return to Step 2 and update the artifacts accordingly.

#### 4.4 — Human review

Assign the PR to a team reviewer. The branch ruleset requires **at least one approval** before merging.

Reviewers should verify:
- Business logic matches the Jira ticket and `proposal.md`.
- Architecture follows the layering rules in `AGENTS.md`.
- Tests cover the new or changed behaviour.
- No secrets, credentials, or PII are committed.

---

### Step 5 — Address Feedback

**Goal**: resolve all review comments before merging.

Use the tool you are most comfortable with. There is no single required tool — pick based on the complexity of the change and your own preference.

| Tool | Best for |
|---|---|
| **Claude Code** | Complex multi-file changes, architectural refactors, or fixes requiring broad context |
| **GitHub Copilot Agent** | Simple, focused fixes directly in VS Code; quick inline suggestions |
| **OpenAI Codex** | CLI-driven code generation; useful in terminal-centric workflows or when you prefer OpenAI models |
| **Cursor** | Iterative, in-context edits in an AI-native editor |
| **GoLand + Junie** | JetBrains-native AI assistance when GoLand is your primary IDE |
| **Manual IDE fix** | Fastest option for trivial, well-understood, or purely stylistic changes |

> **Tips**
> - Manual fixes are often the fastest for small, clearly-defined changes — don't over-engineer the workflow.
> - Ask Copilot for simple, single-location tasks (rename, add nil-check, reword a comment).
> - Reserve Claude Code for changes that require reading many files or understanding broader context.

#### 5.1 — Claude Code

For code-level feedback that touches multiple files or requires context across the codebase, use Claude Code:

```
fix review PR#123
```

Or describe the change explicitly:

```
The reviewer flagged that the handler is calling the repository directly.
Fix it by routing the call through the domain service instead.
```

#### 5.2 — GitHub Copilot Agent

For simple, focused fixes inside VS Code, use Copilot Agent mode (requires GitHub Copilot Chat extension):

1. Open the Copilot Chat panel (`Ctrl+Alt+I` / `Cmd+Alt+I`).
2. Switch to **Agent** mode using the mode selector in the chat input.
3. Describe the fix in plain language, for example:

```
Add a nil-check before dereferencing the user pointer in handlers/auth.go line 42.
```

Copilot will propose a diff that you can accept or discard inline.

#### 5.3 — OpenAI Codex

[Codex CLI](https://github.com/openai/codex) is a terminal-based AI coding agent from OpenAI. Install it with `npm install -g @openai/codex`, then run it in your repository root:

```bash
codex "Fix the review comment: route the handler call through the domain service"
```

Codex reads the repository context and proposes changes you can review before applying.

#### 5.4 — Cursor

In Cursor, use the inline edit shortcut (`Ctrl+K` / `Cmd+K`) to apply a targeted fix:

1. Select the flagged code block.
2. Press `Ctrl+K` (`Cmd+K` on macOS).
3. Type the instruction, for example:

```
Extract this logic into a separate validateInput helper function
```

Cursor applies the edit in-place. Review the diff in the editor before saving.

#### 5.5 — GoLand + Junie

If GoLand is your primary IDE:

1. Open the **Junie** panel (View → Tool Windows → Junie).
2. Describe the fix, for example:

```
The reviewer asked to add error wrapping with fmt.Errorf("...: %w", err) in service/user.go
```

Junie reads `JUNIE.md` (present at the repository root) for team conventions and applies changes following the project's coding guidelines.

#### 5.6 — Manual IDE fix

For trivial or purely stylistic changes, apply fixes directly in your editor without invoking an AI agent. This is often the fastest path for:

- Renaming a variable
- Rewording a comment or error message
- Adding a missing blank line or import

Use Copilot or Junie inline completions (`Tab` to accept) to speed up edits if helpful.

#### 5.7 — Re-trigger review

After making changes, push the updated branch — CI will re-run automatically and Copilot will review the new push. The cycle returns to [Step 4](#step-4--code-review) until all feedback is addressed and the PR is approved.

#### 5.8 — Merge

Once approved and all CI checks pass, merge the PR. The branch ruleset restricts merging to **squash only** — GitHub will squash all commits into a single commit on `main`.

The head branch is automatically deleted after merge.

---

### Step 6 — Archive

**Goal**: close out the OpenSpec change once the PR is merged.

#### 6.1 — Run `/opsx:archive`

```
/opsx:archive
```

Claude will:
1. Check artifact completion status and warn if any are incomplete.
2. Check task completion (`- [x]` vs `- [ ]`) and warn if tasks remain.
3. Assess whether delta specs need to be synced back to the main spec files.
4. Move the change directory to `openspec/changes/archive/YYYY-MM-DD-<name>/`.
5. Display a summary of the archive operation.

#### 6.2 — Update the Jira ticket

Move the Jira ticket to **Done** (or the equivalent closed status on your board). Add a comment with the PR link for traceability.

---

## Quick Reference

| Step | Command / Action |
|---|---|
| Gather specs | Create Jira ticket, link docs, `git checkout -b spec/TICKET-123` |
| Explore requirements | `claude` → `/opsx:explore TICKET-123` |
| Propose change | `claude` → `/opsx:propose TICKET-123` |
| Push spec PR | `git push origin spec/TICKET-123` then open on GitHub |
| Implement change | `claude` → `/opsx:apply <task name>` |
| Regenerate API stubs | `cd src && make generate` |
| Push feature PR | `git push origin feat/TICKET-123` then open on GitHub |
| Run tests | `cd src && go test ./...` |
| Run linter | `cd src && golangci-lint run ./...` |
| AI review PR | `claude` → `/pr-review-toolkit:review-pr` |
| Fix review feedback | Any of: `claude` → `fix review PR#123` · Copilot Agent · Codex · Cursor · GoLand+Junie · Manual IDE |
| Archive change | `claude` → `/opsx:archive` |

### Useful skills and commands

| Skill / Command | Purpose |
|---|---|
| `/opsx:explore` | Think through ideas before proposing a change |
| `/opsx:propose` | Generate proposal, design, and task artifacts |
| `/opsx:apply` | Implement tasks from an active change |
| `/opsx:archive` | Archive a completed change |
| `/pr-review-toolkit:review-pr` | Full PR review using specialist sub-agents |
| `/pr-review-toolkit:silent-failure-hunter` | Hunt for swallowed errors and silent failures |
| `/pr-review-toolkit:code-simplifier` | Simplify over-engineered code |
| `/project-status-summary` | Generate a project health report from Jira, Confluence, Slack |
| `/weekly-bottleneck-report` | Generate a sprint bottleneck and delay report |
| `/estimate-release` | Estimate the release date for a single Jira ticket |

See [README.md](../README.md#using-claude-code-skills) for full documentation of each skill.

# awesome-ai-skills

A curated collection of configuration files, instructions, and best practices for working with AI coding agents. This repository is a personal journey into the world of AI agents, where I share my experiences, insights, and lessons learned while exploring various AI skills and technologies.

## Directory Structure

```
repo root/
├── AGENTS.md                        # Unified instructions for AI coding agents & contributors
├── CLAUDE.md                        # Claude-specific coding instructions
├── CODEX.md                         # Codex/ChatGPT-specific coding instructions
├── JUNIE.md                         # Junie (JetBrains AI) specific instructions
├── .cursorrules                     # Cursor AI editor rules
├── .claude/
│   ├── settings.json                # Claude project settings
│   ├── hooks/                       # Claude lifecycle hooks
│   ├── commands/                    # Claude custom slash commands
│   └── skills/                      # Claude reusable skills
└── .github/
    ├── copilot-instructions.md      # GitHub Copilot instructions
    ├── init-labels.sh               # Script to create standard GitHub issue labels
    ├── init-repo-settings.sh        # Script to apply standard repository settings
    ├── labeler.yml                  # Label-to-file-pattern mappings for the labeler workflow
    ├── rulesets/
    │   ├── protect-main.json        # Ruleset: protect the default branch
    │   └── general-rule.json        # Ruleset: general rules for all branches
    ├── release.yaml                 # Release drafter configuration
    ├── PULL_REQUEST_TEMPLATE.md     # Default pull request template
    └── workflows/
        └── labeler.yml              # GitHub Actions workflow to auto-label pull requests
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

## Getting Started

1. Clone this repository as a reference or template.
2. Copy the relevant configuration files into your own project.
3. Customise the instructions to match your project's conventions, tech stack, and coding standards.

## Repository Initialisation

When setting up a new repository for the first time, run the steps below to apply the team's standard configuration.

### Prerequisites

- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`).
- You are inside the root of the repository you want to initialise.

### 1. Create standard issue labels

The repository ships with a script that creates all required GitHub labels in one step:

```bash
bash .github/init-labels.sh
```

This creates the following labels:

| Label | Colour |
|---|---|
| `feature` | `#0075ca` |
| `spec` | `#e4e669` |
| `chore` | `#ededed` |
| `fix` | `#d73a4a` |
| `docs` | `#0075ca` |
| `refactor` | `#c5def5` |
| `test` | `#bfd4f2` |
| `spec-archive` | `#f9d0c4` |
| `enhancement` | `#a2eeef` |
| `security` | `#ff0000` |
| `migration/database` | `#8a2be2` |

> **Note:** The script uses `--force` so it is safe to re-run at any time. However, any colour or description customisations you have made to a same-named label will be overwritten — review the script before re-running on a repository with hand-edited labels.

### 2. Apply standard repository settings

Run the script below to configure the repository's pull request, commit, and issue settings in one step:

```bash
bash .github/init-repo-settings.sh
```

This applies the following settings:

**Pull Requests**

| Setting | Value |
|---|---|
| Allow merge commits | ✅ Enabled |
| Default merge commit message | Pull request title |
| Allow squash merging | ✅ Enabled |
| Default squash commit message | Pull request title |
| Allow rebase merging | ✅ Enabled |
| Always suggest updating pull request branches | ✅ Enabled |
| Automatically delete head branches | ✅ Enabled |

**Commits**

| Setting | Value |
|---|---|
| Require contributors to sign off on web-based commits | ✅ Enabled |

**Issues**

| Setting | Value |
|---|---|
| Auto-close issues with merged linked pull requests | ✅ GitHub default — no script configuration needed. Use `Closes #<issue>` in PR description or commit message to link and auto-close an issue when the PR is merged. |

> **Note:** The script is idempotent and safe to re-run at any time.

### 3. Enable the labeler workflow

The repository includes a GitHub Actions workflow that automatically applies labels to pull requests based on the files changed.

The workflow is defined in `.github/workflows/labeler.yml` and uses the mapping in `.github/labeler.yml`. Labels are applied according to the following example rules (customise the paths to match your project):

| Label | Branch pattern | Changed files |
|---|---|---|
| `chore` | `chore/*` | — |
| `feature` | `feat/*` | — |
| `spec` | `spec/*` | `openspec/**/*` (excluding archive) |
| `spec-archive` | — | `openspec/changes/archive/**/*` |
| `fix` | `fix/*`, `hotfix/*` | — |
| `docs` | `docs/*` | `docs/**/*` |
| `refactor` | `refactor/*` | — |
| `test` | `test/*` | — |
| `dependencies` | — | `src/go.mod`, `src/go.sum` |
| `migration/database` | — | `src/migrations/**/*` |
| `api` | — | `api/**/*`, `openapi.yaml` |

> The file patterns above are templates; update them to match the directories and files in your repository.

The workflow triggers on `pull_request_target` events (opened, synchronized, or re-opened) and requires no additional configuration beyond the labels being present in the repository — run `init-labels.sh` first if you haven't already.

### 4. Copy AI agent configuration files

Copy the relevant files from this repository into your project (see the [Directory Structure](#directory-structure) table above) and customise them for your tech stack.

### 4. Configure branch rulesets

The repository ships with two ruleset definition files under `.github/rulesets/` that encode the team's branch-protection policy. Import them through the GitHub UI or apply them via the GitHub CLI:

```bash
# Replace OWNER and REPO with your actual repository owner and name (e.g. octocat/my-repo)
gh api repos/OWNER/REPO/rulesets --method POST --input .github/rulesets/protect-main.json
gh api repos/OWNER/REPO/rulesets --method POST --input .github/rulesets/general-rule.json
```

#### Ruleset: `protect-main.json` — Protect the default branch

Targets the **default branch** (`~DEFAULT_BRANCH`) with the following rules:

| Rule | Setting |
|---|---|
| Require signed commits | ✅ enabled |
| Restrict deletions | ✅ enabled |
| Block force pushes | ✅ enabled |
| Require a pull request before merging | ✅ enabled |
| — Required approvals | 1 |
| — Allowed merge methods | Squash only |
| Require code quality results | Errors |
| Automatically request Copilot code review | Review new pushes + Review draft pull requests |

#### Ruleset: `general-rule.json` — General rules for all branches

Targets **all branches** (`~ALL`) with the following rules:

| Rule | Setting |
|---|---|
| Require a pull request before merging | ✅ enabled |
| — Required approvals | 1 |
| — Allowed merge methods | Squash only |
| Require code quality results | Warnings and higher |
| Automatically request Copilot code review | Review new pushes + Review draft pull requests |

> **Note:** Both rulesets use `"enforcement": "active"`. Change this to `"evaluate"` if you want to trial the rules without enforcing them.

---

## Contributing

Contributions, improvements, and new agent configurations are welcome. Please read `AGENTS.md` for contribution guidelines before submitting a pull request.

## License

MIT © Hoa Hoang — see [LICENSE](LICENSE) for details.

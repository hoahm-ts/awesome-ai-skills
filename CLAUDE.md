# CLAUDE.md — Instructions for Claude

This file provides project-specific instructions for [Claude](https://claude.ai) by Anthropic when working in this repository. These instructions complement the unified guidelines in `AGENTS.md`.

---

## Getting Started

Before making any changes, read `AGENTS.md` for the canonical coding guidelines and project overview.

---

## Claude-Specific Behaviour

- Use the `TodoWrite` / `TodoRead` tools to track multi-step tasks.
- Prefer calling multiple independent tools in parallel to maximise efficiency.
- Make the smallest possible change that fully addresses the task.
- Run the repository's existing linters and tests after making changes.
- Use `report_progress` to commit and push verified changes incrementally.

---

## Code Style

- Match the style and conventions already present in the file being edited.
- Prefer explicit names over short abbreviations.
- Keep functions focused on a single responsibility.

---

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat: add new skill file
fix: correct typo in AGENTS.md
docs: update README directory structure
chore: add .gitkeep to empty directories
```

---

## References

- [Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code)
- [`AGENTS.md`](./AGENTS.md) — unified instructions for all agents

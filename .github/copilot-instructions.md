# GitHub Copilot Instructions

This file provides repository-specific instructions for [GitHub Copilot](https://github.com/features/copilot). Copilot will use these instructions as context when generating suggestions in this repository.

---

## Project Context

`awesome-ai-skills` is a curated collection of configuration files, instructions, and best practices for AI coding agents. The repository provides ready-to-use instruction files for Claude, Codex, Junie, Cursor, and GitHub Copilot itself.

---

## Coding Guidelines

- Read `AGENTS.md` for the canonical coding guidelines before making suggestions.
- Match the style and conventions already present in the file being edited.
- Prefer explicit, descriptive names over short abbreviations.
- Keep functions small and focused on a single responsibility.
- Do not introduce new dependencies unless strictly necessary.
- Do not suggest secrets, credentials, or sensitive data.

---

## File Conventions

- Markdown files use [GitHub Flavored Markdown](https://github.github.com/gfm/).
- JSON files must be valid and use 2-space indentation.
- Follow the directory structure documented in `README.md`.

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

- [`AGENTS.md`](../../AGENTS.md) — unified instructions for all agents
- [GitHub Copilot documentation](https://docs.github.com/en/copilot)

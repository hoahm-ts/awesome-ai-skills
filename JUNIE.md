# JUNIE.md — Instructions for Junie (JetBrains AI)

This file provides project-specific instructions for [Junie](https://www.jetbrains.com/junie/) by JetBrains when working in this repository. These instructions complement the unified guidelines in `AGENTS.md`.

---

## Getting Started

Before making any changes, read `AGENTS.md` for the canonical coding guidelines and project overview.

---

## Junie-Specific Behaviour

- Leverage JetBrains IDE inspections and code analysis where available.
- Follow the project's established code style; do not reformat unrelated code.
- Use refactoring tools (rename, extract method, etc.) when restructuring code rather than manual edits.
- Run existing tests after making changes to verify nothing is broken.

---

## Code Style

- Match the style and conventions already present in the file being edited.
- Use meaningful names that reflect intent.
- Keep changes minimal and focused on the task at hand.

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

- [Junie documentation](https://www.jetbrains.com/junie/)
- [`AGENTS.md`](./AGENTS.md) — unified instructions for all agents

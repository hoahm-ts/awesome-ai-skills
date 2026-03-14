# CODEX.md — Instructions for Codex / ChatGPT

This file provides project-specific instructions for [OpenAI Codex](https://openai.com/blog/openai-codex) and ChatGPT-based coding assistants when working in this repository. These instructions complement the unified guidelines in `AGENTS.md`.

---

## Getting Started

Read `AGENTS.md` first and follow it as the source of truth for all coding guidelines and project conventions. It takes precedence over any instructions in this file.

---

## Codex-Specific Behaviour

- Generate code that matches the existing style and conventions in the repository.
- Avoid introducing new dependencies unless strictly necessary.
- Provide clear explanations alongside any generated code when context is helpful.
- Prefer complete, working code snippets over partial examples.

---

## Code Style

- Match the style and conventions already present in the file being edited.
- Use descriptive variable and function names.
- Keep generated blocks concise and focused.

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

- [OpenAI Codex](https://openai.com/blog/openai-codex)
- [`AGENTS.md`](./AGENTS.md) — unified instructions for all agents

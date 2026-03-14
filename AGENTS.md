# AGENTS.md — Unified Instructions for AI Coding Agents & Contributors

This file is the **canonical reference** for all AI coding agents and human contributors working in this repository. All agent-specific instruction files (`CLAUDE.md`, `CODEX.md`, `JUNIE.md`, `.cursorrules`, `.github/copilot-instructions.md`) should be kept consistent with the guidelines here.

---

## Project Overview

`awesome-ai-skills` is a curated collection of configuration files, instructions, and best practices for working with AI coding agents. The goal is to provide a ready-to-use set of files that can be dropped into any project to give AI tools the context they need to be maximally helpful.

---

## General Coding Guidelines

- Write clean, readable, and well-structured code.
- Prefer explicit over implicit — avoid magic numbers, unclear abbreviations, and unexplained side effects.
- Follow the conventions already established in the codebase.
- Keep functions and modules small and focused on a single responsibility.
- Write meaningful commit messages using the [Conventional Commits](https://www.conventionalcommits.org/) format (e.g. `feat:`, `fix:`, `docs:`, `chore:`).

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
5. Open a pull request with a clear title and description.

---

## AI Agent Behaviour

When working as an AI agent in this repository:

- **Read this file first** before making any changes.
- Prefer targeted, surgical edits over large rewrites.
- Do not modify files unrelated to the task.
- If you are unsure about a convention, look for existing examples in the repository before inventing one.
- Do not commit secrets, credentials, or sensitive data.
- Follow the directory structure documented in `README.md`.

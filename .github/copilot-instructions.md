# GitHub Copilot Instructions

Read [`AGENTS.md`](../AGENTS.md) first and follow it as the source of truth for all coding guidelines and project conventions.

## Pull Request Requirements

When creating a pull request, you **must** follow these rules exactly:

### Title format

```
<TICKET_NUMBER>: <description>
```

Example: `JIRA-29: init the project structure`

### Description format

The PR body **must** use this exact section structure — copy it and fill in every section. Leave no section blank or with placeholder text:

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

Valid types: `feat`, `fix`, `docs`, `chore`, `refactor`, `migration/database`.

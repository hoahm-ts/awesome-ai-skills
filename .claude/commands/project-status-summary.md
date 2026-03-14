---
name: "Project Status Summary"
description: "Summarize the current project status from Slack, Jira, Confluence, and email"
category: Productivity
tags: [project-management, status, jira, slack, confluence, summary]
---

Summarize the current project status by searching recent activity across Slack, Jira, Confluence, and email.

**Input**: The argument after `/project-status-summary` should be the project name followed by the Jira board key. For single-word project names, separate with a space. For multi-word project names, wrap them in quotes.

Examples:
- `/project-status-summary tpbank CO`
- `/project-status-summary payments PYMT`
- `/project-status-summary "new payment system" PYMT`

If no input is provided, ask:
> "Which project and Jira board would you like to summarize? (e.g., project: tpbank, board: CO)"

---

## Steps

1. **Parse input**

   Extract:
   - `project` — the project name (first token, e.g., `tpbank`)
   - `board` — the Jira board key (second token, e.g., `CO`)

   If either is missing or ambiguous, ask the user to clarify before proceeding.

2. **Search all sources (last 7 days)**

   Search simultaneously across:

   **Slack**
   - Channels whose name contains `{project}`
   - Messages in any channel whose text contains `{project}`
   - Prioritize messages from project leads, engineering leads, PMs, and SRE members

   **Jira**
   - Tickets on the `{board}` board updated in the last 7 days
   - Status changes, new comments, open blockers, current sprint goal

   **Confluence**
   - Pages created or updated in the last 7 days, filtered by project name in title or space

   **Email**
   - Threads whose subject or body contains `{project}`, from the last 7 days

3. **Analyze and produce the summary**

   Follow the full analysis and output format defined in `.claude/skills/project-status-summary/SKILL.md`.

   Output structure:
   - Overall status badge (🟢 On-track / 🟡 At Risk / 🔴 Blocked)
   - Executive summary (3–5 sentences)
   - Key updates (bullet list)
   - Blockers & risks (bullet list)
   - Action items with owners (bullet list)
   - Decisions made (bullet list)
   - Deadlines & timelines (bullet list)
   - Updated docs & materials (bullet list)
   - Sources searched (provenance list)

---

## Guardrails

- Only report information found in actual sources — never fabricate
- Skip casual chat, off-topic messages, and automated bot noise
- If a source is inaccessible, note it clearly and continue with remaining sources
- Keep each bullet to one clear, concise sentence
- If no data is found for any section, state that explicitly rather than omitting the section

---
name: "Project Status Summary"
description: "Summarize the current project status from Slack, Jira, Confluence, and email"
category: Productivity
tags: [project-management, status, jira, slack, confluence, summary]
---

Summarize the current project status by searching recent activity across Slack, Jira, Confluence, and email.

**Input**: The argument after `/project-status-summary` accepts up to four positional arguments. For multi-word project names, wrap them in quotes.

```
/project-status-summary <project> <board> [days] [channel]
```

Examples:
- `/project-status-summary tpbank CO` — last 7 days, send to your own DM
- `/project-status-summary tpbank CO 14` — last 14 days
- `/project-status-summary tpbank CO 7 #tpbank-updates` — post to a channel
- `/project-status-summary tpbank CO 7 @john` — send as a DM to @john
- `/project-status-summary "new payment system" PYMT 30`

If no input is provided, ask:
> "Which project and Jira board would you like to summarize? (e.g., project: tpbank, board: CO, days: 7, channel: #tpbank-updates)"

---

## Steps

1. **Parse input**

   Extract:
   - `project` — the project name (quoted for multi-word names, e.g., `"new payment system"`)
   - `board` — the Jira board key (e.g., `CO`)
   - `days` — number of days to look back (default: `7` if not provided; accept any positive integer)
   - `channel` — Slack destination for the summary (e.g., `#tpbank-updates` or `@john`); defaults to the user's own DM if not provided

   If `project` or `board` is missing or ambiguous, ask the user to clarify before proceeding.

2. **Search all sources (last `{days}` days)**

   Search simultaneously across all connected sources. As you collect each data point, assign it a sequential reference number starting at **(1)** and record its source link or title:

   **Slack**
   - Channels whose name contains `{project}`
   - Messages in any channel whose text contains `{project}`
   - Record: channel name + message permalink
   - Prioritize messages from project leads, engineering leads, PMs, and SRE members

   **Jira**
   - Tickets on the `{board}` board updated in the last `{days}` days
   - Status changes, new comments, open blockers, current sprint goal
   - Record: ticket ID + direct URL

   **Confluence**
   - Pages created or updated in the last `{days}` days, filtered by project name in title or space
   - Record: page title + direct URL

   **Email**
   - Threads whose subject or body contains `{project}`, from the last `{days}` days
   - Record: thread subject line + sender name

3. **Analyze and produce the summary**

   Follow the full analysis and output format defined in `.claude/skills/project-status-summary/SKILL.md`.

   Output structure:
   - Overall status badge (🟢 On-track / 🟡 At Risk / 🔴 Blocked) with section icons
   - Executive summary (3–5 sentences)
   - Key updates — each bullet ends with its reference number(s) e.g. *(1)*
   - Blockers & risks — each bullet ends with its reference number(s)
   - Action items with owners — each bullet ends with its reference number(s)
   - Decisions made — each bullet ends with its reference number(s)
   - Deadlines & timelines — each bullet ends with its reference number(s)
   - Updated docs & materials — each bullet ends with its reference number(s)
   - Sources table with full links (Slack permalink / Jira URL / Confluence URL / email subject)

4. **Send the summary to Slack**

   After generating the summary, send a formatted version to the `channel` destination (or the user's own DM if `channel` was not provided) using Slack's text formatting (`*bold*`, `_italic_`, `•` bullets, `<url|label>` links).

   If Slack is not available, display the summary in-chat and note it could not be delivered.

---

## Guardrails

- Only report information found in actual sources — never fabricate
- Skip casual chat, off-topic messages, and automated bot noise
- Cite every bullet with at least one reference number in parentheses
- Populate the Sources table with full links or titles for every reference used
- If a source is inaccessible, note it in the Sources table and continue with remaining sources
- Keep each bullet to one clear, concise sentence
- If no data is found for any section, state that explicitly rather than omitting the section

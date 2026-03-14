---
name: project-status-summary
description: Summarize the current project status by searching Slack, email, Jira, and Confluence for recent updates, blockers, decisions, and action items. Produces a structured executive summary for any named project and Jira board.
license: MIT
compatibility: Requires MCP integrations or tool access for Slack, Jira, Confluence, and email.
metadata:
  author: awesome-ai-skills
  version: "1.1.0"
  generatedBy: "1.0.0"
---

Summarize the current project status. Search across all connected sources and produce a concise, structured executive summary.

---

## Overview

This skill collects recent project activity from multiple sources and distills it into a structured report. It is designed for project leads, engineering managers, and PMs who need a fast, reliable pulse on project health without manually trawling through channels and tickets.

---

## Inputs

The user must provide:

| Parameter | Description | Default | Example |
|-----------|-------------|---------|---------|
| `project` | Human-readable project name used to filter channels and messages | *(required)* | `tpbank` |
| `board` | Jira board key used to query tickets and sprints | *(required)* | `CO` |
| `days` | Number of days to look back | `7` | `14` |
| `channel` | Slack destination for the summary (DM handle or channel name) | user's own DM | `@john` or `#tpbank-updates` |

If `project` or `board` is missing, ask the user before proceeding:

> "Which project and Jira board would you like to summarize? (e.g., project: tpbank, board: CO, days: 7, channel: #tpbank-updates)"

If `days` is not provided, use **7** as the default. Accept any positive integer.

---

## Search Strategy

### Time window

Search the last **`{days}`** days (default: 7). Prioritize the most recent updates. Ignore anything older unless it is still referenced as an active blocker or open action item.

### Sources and filters

Search each source using the project name as the primary keyword. Apply the filters below.

**Important — numbering and source links:**
As you collect each individual data point (a Slack message, a Jira ticket, a Confluence page, an email thread), assign it a unique sequential reference number starting at **(1)**. Every distinct data point gets its own number — even multiple messages from the same channel or multiple tickets from the same board. Record the following for each reference:

| # | Source type | What to record |
|---|-------------|----------------|
| (n) | Slack | Channel name + message permalink (e.g., `https://workspace.slack.com/archives/C.../p...`) |
| (n) | Jira | Ticket ID + direct URL (e.g., `https://company.atlassian.net/browse/CO-123`) |
| (n) | Confluence | Page title + direct URL (e.g., `https://company.atlassian.net/wiki/spaces/.../pages/...`) |
| (n) | Email | Thread subject line + sender |

When a single output bullet is supported by multiple data points, append all of their reference numbers (e.g., *(1, 3)*). When two data points say the same thing (exact duplicates), you may merge them under a single reference number and note the duplicate in the Sources table.

#### 1. Slack

- Channels whose **name contains** the project name (e.g., `#tpbank`, `#tpbank-eng`, `#co-sprint`)
- Messages in **any channel** whose **text contains** the project name
- Prioritize messages from: project leads, engineering leads, PMs, SRE members (identify by title, role mention, or channel pinned members list where available)
- Skip: emoji reactions, casual greetings, automated bot noise unrelated to status

#### 2. Jira

- Query the specified board (`board` parameter) for:
  - Tickets **updated** in the last `{days}` days
  - Tickets with status changes (e.g., moved to In Progress, Done, Blocked)
  - Tickets with new comments from the last `{days}` days
  - Open blockers or tickets flagged with impediments
  - Sprint goal and velocity (current sprint only)
- Note ticket IDs, assignees, and due dates where present

#### 3. Confluence

- Pages in spaces **linked to the project or board** (search by project name in space key or page title)
- Pages **created or updated** in the last `{days}` days
- Look for: meeting notes, decision logs, architecture docs, runbooks, status reports

#### 4. Email

- Threads whose **subject or body contains** the project name
- Prioritize threads from project-related distribution lists or direct stakeholder correspondence
- Skip: automated notifications, marketing, and out-of-office replies

---

## Analysis

After collecting raw data, extract the following signals:

| Signal | Description |
|--------|-------------|
| **Overall status** | Infer On-track / At Risk / Blocked based on open blockers, missed milestones, or explicit status statements |
| **Key updates** | Concrete progress made: features shipped, PRs merged, milestones hit, decisions confirmed |
| **Blockers / risks** | Explicit blockers, escalations, unresolved dependencies, or recurring concerns |
| **Decisions made** | Architectural decisions, scope changes, priority shifts, vendor choices |
| **Action items** | Tasks assigned to specific owners, follow-ups requested, commitments made |
| **Owners** | People explicitly named as responsible for key tasks or decisions |
| **Deadlines** | Dates, sprint end dates, release targets, or customer commitments mentioned |
| **Updated docs** | Confluence pages, design docs, runbooks, or shared materials updated recently |

When inferring overall status, apply these heuristics:

- **On-track**: No open blockers, sprint progress normal, no escalations
- **At Risk**: One or more blockers present, deadline concerns raised, or capacity issues flagged
- **Blocked**: Critical blocker explicitly named, escalation raised, or sprint halted

When signals conflict (e.g., a blocker exists but sprint velocity is normal), choose the more conservative status and note the ambiguity in the Executive Summary. For example: "Sprint velocity is healthy, but an unresolved dependency on the payments team introduces risk." Use 🟡 At Risk rather than 🟢 On-track when in doubt.

---

## Output Format

Produce the summary in this exact structure. Append the reference number(s) in parentheses at the end of each bullet to indicate which source(s) the information came from (e.g., *(1)*, *(2, 3)*).

---

### Project Status Summary — {project} ({board} board)
*Last updated: {YYYY-MM-DD} · Time range: last {days} days*

**Overall Status: 🟢 On-track / 🟡 At Risk / 🔴 Blocked**

#### 📋 Executive Summary
3–5 sentences covering: what the project is doing right now, its health, any notable wins or concerns, and the most important near-term focus.

#### ✅ Key Updates
- {Concrete progress item} *(1)*
- {Another update} *(2, 3)*
- …

#### 🚧 Blockers & Risks
- {Blocker or risk description, owner if known} *(4)*
- …
*(If none: "No active blockers identified in the last {days} days.")*

#### 📌 Action Items
- [ ] {Action item description} — **Owner**: {name or "Unassigned"} · *{Due date if mentioned}* *(5)*
- …
*(If none: "No explicit action items identified.")*

#### 💡 Decisions Made
- {Decision description, date if available, decision-maker if known} *(6)*
- …
*(If none: "No decisions recorded in this period.")*

#### 📅 Deadlines & Timelines
- {Milestone or deadline} — {Date} *(7)*
- …
*(If none: "No explicit deadlines mentioned.")*

#### 📄 Updated Docs & Materials
- [{Page or document title}]({link}) — updated {date} *(8)*
- …
*(If none: "No document updates found.")*

#### 🔗 Sources
| # | Source | Link / Reference |
|---|--------|-----------------|
| (1) | Slack #channel-name | [Message permalink]({slack_message_url}) |
| (2) | Jira CO-123 | [CO-123: Ticket title]({jira_ticket_url}) |
| (3) | Confluence | [Page title]({confluence_page_url}) |
| (4) | Email | "Thread subject line" from {sender} |
| … | … | … |

*Sources not accessible during this run: {list any inaccessible sources, or "none"}*

---

## Slack Delivery

After generating the summary, send it to the Slack destination specified by the `channel` parameter. If `channel` is not provided, send the summary as a DM to the user's own Slack account. Use Slack's text formatting (`*bold*`, `_italic_`, `•` bullets, `<url|label>` links).

**Slack message format:**

```
*📊 Project Status Summary — {project} ({board} board)*
_Last updated: {YYYY-MM-DD} · Time range: last {days} days_

*Overall Status:* 🟢 On-track / 🟡 At Risk / 🔴 Blocked

*Executive Summary*
{3–5 sentence summary}

*✅ Key Updates*
• {Update 1} _(1)_
• {Update 2} _(2, 3)_

*🚧 Blockers & Risks*
• {Blocker 1} _(4)_
_(or: No active blockers identified.)_

*📌 Action Items*
• ☐ {Action item} — *Owner:* {name} · _{due date}_ _(5)_
_(or: No explicit action items identified.)_

*💡 Decisions Made*
• {Decision} _(6)_

*📅 Deadlines & Timelines*
• {Milestone} — {date} _(7)_

*🔗 Sources*
_(1) <{slack_message_url}|Slack #channel-name>_
_(2) <{jira_url}|CO-123: Ticket title>_
_(3) <{confluence_url}|Confluence: Page title>_
_(4) Email: "Thread subject" from {sender}_
```

If the Slack integration is not available, display the summary in the chat window only and note: "ℹ️ Slack delivery is unavailable — displaying summary here instead."

## Guardrails

- **Do not fabricate information.** Only include details found in the actual sources. If a source is inaccessible, state that clearly in the Sources table.
- **Do not include casual or off-topic chat.** Filter out unrelated discussions, banter, and automated notifications.
- **Cite every bullet.** Every item in Key Updates, Blockers, Action Items, Decisions, Deadlines, and Updated Docs must have at least one reference number in parentheses.
- **Populate the Sources table fully.** For each reference number used in bullets, list the exact Slack message permalink, Jira ticket URL, Confluence page URL, or email subject line.
- **Respect privacy.** Do not include personal information beyond names and roles relevant to project work.
- **Be concise.** Each bullet should be one clear sentence. Avoid padding.
- **Preserve nuance.** If a status is unclear or conflicting signals exist, say so rather than forcing a definitive label.
- **If no data is found**, say so explicitly: "No relevant updates found for project '{project}' in the last {days} days across the sources searched."

---

## Handling Missing Integrations

If a source is not connected or the search returns no results:

1. Note it clearly in the "Sources" table: `Slack: not accessible`
2. Continue with the remaining sources
3. If **all** sources are inaccessible, inform the user:
   > "I was unable to access any of the configured sources (Slack, Jira, Confluence, email). Please ensure the relevant MCP integrations are connected and retry."

---

## Example Invocations

```
Please summarize the current project status for the project: "tpbank", on the CO board.
```

```
Please summarize the current project status for the project: "tpbank", on the CO board, for the last 14 days.
```

This skill is also invocable via the `/project-status-summary` slash command:

```
/project-status-summary tpbank CO
/project-status-summary tpbank CO 14
```

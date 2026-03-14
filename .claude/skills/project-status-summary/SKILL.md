---
name: project-status-summary
description: Summarize the current project status by searching Slack, email, Jira, and Confluence for recent updates, blockers, decisions, and action items. Produces a structured executive summary for any named project and Jira board.
license: MIT
compatibility: Requires MCP integrations or tool access for Slack, Jira, Confluence, and email.
metadata:
  author: awesome-ai-skills
  version: "1.0"
  generatedBy: "1.0.0"
---

Summarize the current project status. Search across all connected sources and produce a concise, structured executive summary.

---

## Overview

This skill collects recent project activity from multiple sources and distills it into a structured report. It is designed for project leads, engineering managers, and PMs who need a fast, reliable pulse on project health without manually trawling through channels and tickets.

---

## Inputs

The user must provide:

| Parameter | Description | Example |
|-----------|-------------|---------|
| `project` | Human-readable project name used to filter channels and messages | `tpbank` |
| `board` | Jira board key used to query tickets and sprints | `CO` |

If either is missing, ask the user before proceeding:

> "Which project and Jira board would you like to summarize? (e.g., project: tpbank, board: CO)"

---

## Search Strategy

### Time window

Search the **last 7 days**. Prioritize the most recent updates. Ignore anything older unless it is still referenced as an active blocker or open action item.

### Sources and filters

Search each source using the project name as the primary keyword. Apply the filters below.

#### 1. Slack

- Channels whose **name contains** the project name (e.g., `#tpbank`, `#tpbank-eng`, `#co-sprint`)
- Messages in **any channel** whose **text contains** the project name
- Prioritize messages from: project leads, engineering leads, PMs, SRE members (identify by title, role mention, or channel pinned members list where available)
- Skip: emoji reactions, casual greetings, automated bot noise unrelated to status

#### 2. Jira

- Query the specified board (`board` parameter) for:
  - Tickets **updated** in the last 7 days
  - Tickets with status changes (e.g., moved to In Progress, Done, Blocked)
  - Tickets with new comments from the last 7 days
  - Open blockers or tickets flagged with impediments
  - Sprint goal and velocity (current sprint only)
- Note ticket IDs, assignees, and due dates where present

#### 3. Confluence

- Pages in spaces **linked to the project or board** (search by project name in space key or page title)
- Pages **created or updated** in the last 7 days
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

Produce the summary in this exact structure:

---

### Project Status Summary — {project} ({board} board)
*Last updated: {YYYY-MM-DD} · Time range: last 7 days*

**Overall Status: 🟢 On-track / 🟡 At Risk / 🔴 Blocked**

#### Executive Summary
3–5 sentences covering: what the project is doing right now, its health, any notable wins or concerns, and the most important near-term focus.

#### Key Updates
- {Concrete progress item, with source reference where useful}
- {Another update}
- …

#### Blockers & Risks
- {Blocker or risk description, owner if known, source}
- …
*(If none: "No active blockers identified in the last 7 days.")*

#### Action Items
- [ ] {Action item description} — **Owner**: {name or "Unassigned"} · *{Due date if mentioned}*
- …
*(If none: "No explicit action items identified.")*

#### Decisions Made
- {Decision description, date if available, decision-maker if known}
- …
*(If none: "No decisions recorded in this period.")*

#### Deadlines & Timelines
- {Milestone or deadline} — {Date}
- …
*(If none: "No explicit deadlines mentioned.")*

#### Updated Docs & Materials
- [{Page or document title}]({link if available}) — updated {date}
- …
*(If none: "No document updates found.")*

#### Sources Searched
- Slack: {list of channels searched}
- Jira: {board key}, {number} tickets reviewed
- Confluence: {number} pages reviewed
- Email: {number} threads reviewed

---

## Guardrails

- **Do not fabricate information.** Only include details found in the actual sources. If a source is inaccessible, state that clearly under "Sources Searched".
- **Do not include casual or off-topic chat.** Filter out unrelated discussions, banter, and automated notifications.
- **Attribute sources when useful.** For significant updates or decisions, briefly note where the information came from (e.g., "per #tpbank-eng on Mon", "Jira CO-142").
- **Respect privacy.** Do not include personal information beyond names and roles relevant to project work.
- **Be concise.** Each bullet should be one clear sentence. Avoid padding.
- **Preserve nuance.** If a status is unclear or conflicting signals exist, say so rather than forcing a definitive label.
- **If no data is found**, say so explicitly: "No relevant updates found for project '{project}' in the last 7 days across the sources searched."

---

## Handling Missing Integrations

If a source is not connected or the search returns no results:

1. Note it clearly in the "Sources Searched" section: `Slack: not accessible`
2. Continue with the remaining sources
3. If **all** sources are inaccessible, inform the user:
   > "I was unable to access any of the configured sources (Slack, Jira, Confluence, email). Please ensure the relevant MCP integrations are connected and retry."

---

## Example Invocation

```
Please summarize the current project status for the project: "tpbank", on the CO board.
```

This skill is also invocable via the `/project-status-summary` slash command:

```
/project-status-summary tpbank CO
```

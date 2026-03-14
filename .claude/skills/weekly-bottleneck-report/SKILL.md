---
name: weekly-bottleneck-report
description: Generate an internal engineering report highlighting delays, bottlenecks, over-assignment, QA queue pressure, stale work items, and estimated release timelines from the active JIRA sprint. Searches Jira, Confluence, GitHub, and Calendar.
license: MIT
compatibility: Requires MCP integrations for Jira (and optionally Confluence, GitHub, Calendar).
metadata:
  author: awesome-ai-skills
  version: "1.0.0"
  generatedBy: "1.0.0"
---

Generate a comprehensive weekly sprint bottleneck and delay report. Search Jira for active sprint health signals, cross-reference Confluence for weekly updates, GitHub for PR/commit activity, and Calendar for upcoming deadlines.

> ⚠️ **Internal report only — not for Confluence publication.**
> This is a companion to the weekly Confluence update.

---

## Overview

This skill produces a deep engineering analysis for sprint leads and engineering managers. It goes beyond a simple status report to identify *why* things are delayed, *who* is blocked, and *when* work will realistically ship.

---

## Inputs

| Parameter | Description | Default | Example |
|-----------|-------------|---------|---------|
| `projects` | One or more Jira project keys, comma-separated | *(required)* | `AGI` or `AGI,CO` |
| `days` | Time window in days for stale-item detection | `7` | `14` |
| `board` | Jira board ID; auto-detected from the first project if omitted | auto | `337` |
| `recipient` | Slack destination for the condensed summary — a DM handle or channel name | user's own DM | `@john` or `#eng-leads` |

If `projects` is missing or ambiguous, ask:
> "Which project(s) would you like to analyse? (e.g., AGI, or AGI,CO for multiple). What time window in days? (default: 7)"

For multiple projects, run every JQL query for each project key and merge results.

---

## Step 1 — Fetch High-Priority Delayed Items

**JQL** (run per project key `{P}`):
```
project = {P} AND sprint in openSprints() AND priority in (High, Highest)
AND status not in ("Released / Done", "Done", "Closed")
ORDER BY priority ASC, updated ASC
```

**Fields**: `summary, status, priority, assignee, labels, created, updated, statuscategorychangedate, parent, subtasks`

**Analysis**:
- Count sprint labels per ticket (patterns: `{P}_ENG_20XX_SXX_PLANNED` / `_INJECTED`):
  - **1 label** → on track
  - **2 labels** → ⚠️ at risk
  - **3+ labels** → 🔴 significantly delayed
- Flag tickets where `statuscategorychangedate` is **>14 days ago** (stuck in current status)
- Note the assignee and parent epic for each delayed item

---

## Step 2 — Fetch Blocked Items

**JQL** (run per project key `{P}`):
```
project = {P} AND sprint in openSprints() AND status = "Blocked / On Hold"
ORDER BY priority ASC
```

**Fields**: `summary, status, priority, assignee, labels, created, updated, statuscategorychangedate, parent`

**Analysis**:
- Compute **days blocked** = today − `statuscategorychangedate`
- Group by parent story/epic to expose systemic blocks
- Classify blocker type from labels or summary keywords:
  - 🔗 External dependency
  - 🤝 Cross-team alignment
  - 📐 DS / design sign-off
  - 🔧 Technical blocker
  - ❓ Unknown

---

## Step 3 — Analyse Developer Workload

**JQL** (run per project key `{P}`):
```
project = {P} AND sprint in openSprints()
AND status in ("In Development", "In Code Review", "Testing", "In Progress", "Code Review")
ORDER BY priority ASC
```

**Fields**: `summary, status, priority, assignee, parent, labels`

**Analysis**:
- Count active tickets per assignee
- **3+ active tickets** → 🔴 over-assigned (apply 1.5× time multiplier in estimates)
- Multiple tickets in conflicting statuses (e.g. "In Progress" + "Testing") → ⚠️ context-switching risk
- For each over-assigned developer, list all their tickets and recommend which to pause or delegate

---

## Step 4 — Analyse QA Queue

**JQL** (run per project key `{P}`):
```
project = {P} AND sprint in openSprints() AND status = "Ready for Test"
ORDER BY priority ASC, updated ASC
```

**Fields**: `summary, status, priority, assignee, updated, customfield_10747`

**Analysis**:
- **Total queue depth** and breakdown by priority (Highest / High / Medium / Low)
- **Concentration risk**: many items from the same developer or feature area
- **QA queue position** for each item — used by Step 7 estimates
- Recommend prioritisation order: priority → Affirm pilot dependencies → batch-testable groups

---

## Step 5 — Identify Stale Issues

**JQL** (run per project key `{P}`):
```
project = {P} AND sprint in openSprints() AND updated <= -{days}d
AND status not in ("Released / Done", "Done", "Closed", "Cancelled")
ORDER BY updated ASC
```

**Fields**: `summary, status, priority, assignee, updated, comment`

**Analysis**:
- Items not updated for **`{days}`+ days** (uses the same `days` parameter; the stale-item JQL always reflects the configured window)
- Inspect last comment: unanswered questions, no reviewer assigned in Code Review, pending decisions
- Flag **Highest/High** stale items as critical
- Flag Code Review items with **zero comments** (no reviewer assigned)

---

## Step 6 — Cross-reference Confluence, GitHub & Calendar

Search these additional sources to enrich context:

**Confluence**
- Pages updated in the last `{days}` days whose title or space contains any of the project names
- Look for: weekly update pages, decision logs, architecture notes
- Use these to identify workstream groupings referenced in the report

**GitHub**
- PRs opened or updated in the last `{days}` days referencing any project ticket IDs
- Stale PRs (no activity for `{days}` days) that are still open
- Record: PR title, author, linked ticket ID, days since last activity

**Calendar** *(if accessible)*
- Upcoming sprint ceremonies, release dates, or milestone deadlines within the next 14 days
- External dependency meetings or review sessions
- Record: event title, date, attendees relevant to the project

---

## Step 7 — Estimate Release Timelines

For each major workstream (sourced from Confluence weekly update or inferred from epic groupings), apply the `/estimate-release` methodology:

**Base time by status**:
| Status | Base estimate |
|--------|--------------|
| In Progress / In Development | 5–10 business days |
| Code Review / In Code Review | 3–7 business days |
| Testing | 2–5 business days |
| Ready for Test | 3–7 business days (depends on QA queue position) |
| Ready For Release | 1–2 business days |

**Adjustment factors** (apply in order):
1. **QA queue position** → add 2–3 days per item ahead in queue
2. **Developer over-assignment** → multiply by 1.5× if dev has 3+ active items
3. **Sprint slippage** → add 30% buffer for 2 sprints slipped; 50% for 3+
4. **Blocked items** → add 5–10 days if external dependency, or mark **TBD**
5. **Complexity** → simple data fix vs. multi-subtask feature

Express all estimates in **business days** from today.

---

## Step 8 — Generate Recommendations

Produce three tiers:

**🚨 Immediate Actions (This Week)**
- Specific unblocking steps — name the ticket, the person, and the action
- Deployments ready to ship
- Reviewers to assign

**📅 This Sprint**
- Load balancing suggestions (who to delegate to)
- Items to deprioritise or carry over
- QA focus order

**🔄 Process Improvements**
- Systemic fixes for recurring patterns
- Definition-of-ready / definition-of-done gaps
- Handoff or communication gaps

---

## Step 9 — Save the Report

Write the generated report to:
```
local_files/Jira/weekly/<YYYY-MM-DD>-bottleneck-report.md
```

Where `<YYYY-MM-DD>` is today's date.

---

## Step 10 — Send Slack Summary

After saving the report, send a condensed summary to the `recipient` parameter (defaults to the user's own DM if not provided). Use Slack's text formatting (`*bold*`, `_italic_`, `•` bullets, `<url|label>` links).

**Slack message format:**

```
*🔍 Sprint Bottleneck & Delay Report — {YYYY-MM-DD}*
_Projects: {projects} · Time window: last {days} days_
⚠️ _Internal engineering report — not for Confluence publication._

*Sprint Health*
• 📦 Open sprint issues: {n}
• 🔴 Blocked: {n}  •  🧪 QA queue: {n}  •  🚀 Ready for release: {n}
• 👤 Over-assigned devs: {n}  •  ⏩ Slipped 2+ sprints: {n}  •  💤 Stale (>{days}d): {n}

*🔴 Top Delayed Items*
• <{url}|{TICKET-ID}>: {summary} — {n} sprints slipped, assigned to {assignee}
• _(list up to 5 highest-risk items)_

*🚧 Blocked Items*
• <{url}|{TICKET-ID}>: {summary} — blocked {n} days ({blocker type})
• _(or: No blocked items.)_

*👤 Over-Assigned Developers*
• {name}: {n} active tickets — _{recommendation}_
• _(or: No over-assigned developers.)_

*🧪 QA Queue*
• {n} items — top priority: <{url}|{TICKET-ID}> ({priority})
• Concentration risk: {description or "none"}

*✅ Immediate Actions*
• *<{url}|{TICKET-ID}>* — {specific action} — *Owner*: {name}
• …

*📄 Full Report*
Saved to `local_files/Jira/weekly/{YYYY-MM-DD}-bottleneck-report.md`
```

If Slack is not available, display the summary in the chat and note:
> "ℹ️ Slack delivery is unavailable — displaying summary here instead."

---

## Output Format

```markdown
# 🔍 Sprint Bottleneck & Delay Report — {YYYY-MM-DD}

> ⚠️ Internal engineering report — not for Confluence publication.
> Companion to the [Weekly Update]({confluence_link_if_available}).

---

## 1. Sprint Health Summary

| Metric | Value |
|--------|-------|
| 📦 Total open sprint issues | {n} |
| 🔴 Blocked items | {n} |
| 🧪 QA queue depth | {n} |
| 🚀 Ready for release | {n} |
| 👤 Over-assigned developers | {n} |
| ⏩ Items slipped 2+ sprints | {n} |
| 💤 Stale issues (>{days}d) | {n} |

---

## 2. 🔴 High-Priority Items With Significant Delays

| Ticket | Summary | Priority | Status | Sprints Slipped | Assignee | Risk |
|--------|---------|----------|--------|-----------------|----------|------|
| [AGI-XXX]({url}) | {summary} | 🔴 Highest | {status} | **{n}** | {name} | 🔴 High |
| … | … | … | … | … | … | … |

### Key Observations
- {Observation with specific ticket references}
- …

---

## 3. 🚧 Blocked Items

| Ticket | Summary | Priority | Blocked Since | Assignee | Blocker Type |
|--------|---------|----------|---------------|----------|--------------|
| [AGI-XXX]({url}) | {summary} | {priority} | **{n} days** | {name} | {type icon + label} |
| … | … | … | … | … | … |

### Impact Assessment
- {Impact with action needed — specific owner and ticket}
- …

---

## 4. 👤 Over-Assigned Developers

### {Developer Name} — 🔴 {n} Active Tickets

| Ticket | Summary | Status | Priority |
|--------|---------|--------|----------|
| [AGI-XXX]({url}) | {summary} | {status} | {priority} |
| … | … | … | … |

**Risk**: {Explanation}
**Recommendation**: {Specific action — which ticket to pause, who to delegate to}

---

## 5. 🧪 QA Queue Bottleneck

**Queue depth**: {n} items

| Priority | Count |
|----------|-------|
| 🔴 Highest | {n} |
| 🟠 High | {n} |
| 🟡 Medium | {n} |
| 🟢 Low | {n} |

**Concentration risk**: {description if any}

**Recommended QA order**:
1. [AGI-XXX]({url}) — {reason}
2. …

---

## 6. 💤 Stale Issues — Unanswered Questions & No-Activity Items

| Ticket | Priority | Status | Last Updated | Last Comment | Summary |
|--------|----------|--------|-------------|--------------|---------|
| [AGI-XXX]({url}) | {priority} | {status} | **{n} days ago** | {excerpt} | {summary} |
| … | … | … | … | … | … |

### Critical Staleness Observations
- {Observation — specific ticket and suggested action}
- …

---

## 7. 📅 Estimated Release Timelines

### {Workstream Name}

| Component | Status | Est. Release | Reasoning |
|-----------|--------|-------------|-----------|
| [AGI-XXX: {Feature}]({url}) | {status} | **{YYYY-MM-DD}** | {Brief reasoning with factors applied} |
| … | … | … | … |

*(Repeat per workstream)*

---

## 8. ✅ Recommendations

### 🚨 Immediate Actions (This Week)
- **[{Ticket}]({url})** — {specific action} — **Owner**: {name}
- …

### 📅 This Sprint
- {Recommendation with ticket reference and owner}
- …

### 🔄 Process Improvements
- {Systemic recommendation}
- …

---

## Appendix: How to Generate This Report

**Command**: `/weekly-bottleneck-report {projects} {days}`

**Skill file**: `.claude/skills/weekly-bottleneck-report/SKILL.md`

**Per-ticket deep-dive**: `/estimate-release <TICKET-ID>`
- Fetches ticket details, analyses subtasks, developer workload, QA queue position, and sprint slippage
- Produces a structured timeline estimate with blockers and risks

**MCP integrations required**:
- **Jira** — [mcp-atlassian](https://github.com/sooperset/mcp-atlassian)
- **Confluence** — same server as Jira (optional, enriches workstream context)
- **GitHub** — [github-mcp-server](https://github.com/github/github-mcp-server) (optional, surfaces stale PRs)
- **Calendar** — Google Calendar or Outlook MCP (optional, surfaces upcoming deadlines)
- **Slack** — [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack) (optional, delivers condensed summary via DM)
```

---

## Guardrails

- **Never fabricate data.** Only report information found in actual Jira, Confluence, GitHub, or Calendar sources.
- **Sprint label parsing** is the most important signal — always cross-reference label counts for slippage detection.
- **Recommendations must be specific**: name the ticket ID, the person responsible, and the exact action.
- **Release estimates must be conservative**: always apply sprint slippage buffers for items with 2+ sprint labels.
- **Express all time estimates in business days** (exclude weekends and public holidays).
- **QA queue is the #1 bottleneck** — always include queue depth, concentration risk, and recommended prioritisation.
- **Mark the report as internal** — include the `> ⚠️ Internal engineering report` disclaimer at the top.
- **Always send the Slack summary** after saving the report — default to the user's own DM when `recipient` is not provided.
- If an optional source (Confluence, GitHub, Calendar, Slack) is not accessible, note it and continue with remaining sources.
- If no data is found for a section, say so explicitly rather than omitting the section.
- Include the Appendix with `/estimate-release` skill documentation in every generated report.

---

## Handling Missing Integrations

If a source is not connected:
1. Note it in the report header: `> ℹ️ GitHub not connected — PR data unavailable.`
2. Continue with remaining sources
3. If **Jira is not connected**, stop and inform the user:
   > "Jira is required for this report. Please ensure the Jira MCP integration is connected and retry."
4. If **Slack is not connected**, display the condensed summary in the chat and note:
   > "ℹ️ Slack delivery is unavailable — displaying summary here instead."

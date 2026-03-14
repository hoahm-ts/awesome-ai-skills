---
name: "Weekly Bottleneck & Delay Report"
description: "Generate an internal engineering report highlighting delays, bottlenecks, over-assignment, QA queue pressure, stale work, and estimated release timelines from the active JIRA sprint"
category: Engineering
tags: [jira, bottleneck, sprint, delay, qa, estimation, weekly-report]
---

Generate the weekly sprint bottleneck and delay report for one or more projects.

**Usage**

```
/weekly-bottleneck-report <projects> [days] [board] [recipient]
```

Examples:
- `/weekly-bottleneck-report AGI` — single project, last 7 days, summary sent to your DM
- `/weekly-bottleneck-report AGI,CO` — multiple projects, last 7 days
- `/weekly-bottleneck-report AGI 14` — single project, last 14 days
- `/weekly-bottleneck-report "AGI, CO" 7 337` — with explicit board ID
- `/weekly-bottleneck-report AGI 7 337 @john` — send summary DM to @john
- `/weekly-bottleneck-report AGI 7 337 #eng-leads` — post summary to a channel

If no input is provided, ask:
> "Which project(s) would you like to analyse? (e.g., AGI, or AGI,CO for multiple). What time window? (default: 7 days)"

---

## Steps

Follow the full analysis and output format defined in `.claude/skills/weekly-bottleneck-report/SKILL.md`.

### Quick reference

1. **Parse input** — extract `projects` (comma-separated list), `days` (default `7`), optional `board`, and optional `recipient`
2. **Fetch high-priority delayed items** — JQL: `priority in (High, Highest)` + not done, ordered by `updated ASC`
3. **Fetch blocked items** — JQL: `status = "Blocked / On Hold"`
4. **Analyse developer workload** — JQL: active statuses, count per assignee, flag 3+ tickets as over-assigned
5. **Analyse QA queue** — JQL: `status = "Ready for Test"`, prioritise by priority + Affirm pilot dependencies
6. **Identify stale issues** — JQL: `updated <= -{days}d`, flag unanswered questions
7. **Estimate release timelines** — apply `/estimate-release` methodology per workstream
8. **Generate recommendations** — immediate actions, this sprint, process improvements
9. **Save report** — write to `local_files/Jira/weekly/<YYYY-MM-DD>-bottleneck-report.md`
10. **Send Slack summary** — deliver condensed summary to `recipient` via DM (default: your own DM)

---

## Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `projects` | ✅ | — | One or more Jira project keys, comma-separated (e.g., `AGI` or `AGI,CO`) |
| `days` | ❌ | `7` | Time window in days for stale-item detection and context |
| `board` | ❌ | auto-detect | Jira board ID; if omitted, auto-detected from the first project |
| `recipient` | ❌ | user's own DM | Slack destination for the summary — a channel (`#name`) or DM handle (`@name`) |

---

## Output

The full report is saved to `local_files/Jira/weekly/<YYYY-MM-DD>-bottleneck-report.md` and displayed in the chat.
A condensed summary is also sent to the `recipient` via Slack DM (or your own DM if not specified).

For the full report structure, formatting rules, and Slack delivery format, see `.claude/skills/weekly-bottleneck-report/SKILL.md`.


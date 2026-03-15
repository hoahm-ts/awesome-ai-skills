---
name: "Engineering Team Analysis"
description: "Deep-dive analysis of an engineering team's work — collects Jira tickets, Slack signals, and meeting transcripts then produces activity clusters, stakeholder maps, JTBD decomposition, AI readiness ratings, and a final report"
category: Engineering
tags: [jira, slack, engineering, analysis, jtbd, ai-readiness, team-health]
---

Run a full engineering team analysis for a given project, Jira board, and time window.

**Usage**

```
/eng-team-analysis [project] [board] [window]
```

Examples:
- `/eng-team-analysis` — uses all defaults (project=tpbank, board=CO, window=30)
- `/eng-team-analysis tpbank` — explicit project, default board and window
- `/eng-team-analysis tpbank CO` — explicit project and board, default window
- `/eng-team-analysis tpbank CO 60` — explicit project, board, and 60-day window
- `/eng-team-analysis payments ENG 90` — different project with 90-day window

If no input is provided, prompt:
> "Which project, board, and time window would you like to analyse? (defaults: project=tpbank, board=CO, window=30 days)"

---

## Steps

Follow the full analysis and output format defined in `.claude/skills/eng-team-analysis/SKILL.md`.

### Quick reference

1. **Parse input** — extract `project` (default: `tpbank`), `board` (default: `CO`), `window` (default: `30`)
2. **Collect Jira data** — paginate all tickets on `{board}` updated in the last `{window}` days; compute TAT metrics per ticket
3. **Search Slack** — channels containing `{project}` + messages referencing `{project}`, last `{window}` days
4. **Search Google Drive** — meeting transcripts and recordings involving the team (optional)
5. **Cluster activities** — classify tickets into 6–8 non-overlapping activity clusters named by what the work accomplishes
6. **Map stakeholders** — identify every requester by role, request volume, and implicit success measure
7. **Decompose JTBDs** — extract Jobs-To-Be-Done per cluster; re-cluster by demand-side structure; produce JTBD × Activity matrix
8. **Rate AI readiness** — HIGH / MEDIUM / LOW per JTBD cluster with three-phase roadmap
9. **Generate report** — save to `local_files/eng-analysis/{project}-{YYYY-MM-DD}-eng-team-analysis.md` and display in chat
10. **Send Slack summary** — deliver executive summary to user's DM (or specified channel)

---

## Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `project` | ❌ | `tpbank` | Project name used to filter Jira tickets, Slack channels, and Drive search |
| `board` | ❌ | `CO` | Jira board name or key |
| `window` | ❌ | `30` | Look-back window in days |

---

## Output

The full report is saved to `local_files/eng-analysis/{project}-{YYYY-MM-DD}-eng-team-analysis.md` and displayed in the chat.

Report sections:
1. Executive Summary
2. Activity Clusters (table: counts, TAT, top assignees)
3. Stakeholder Map
4. JTBD Clusters (functional + emotional + implicit dimensions, Good vs. Great, evidence)
5. JTBD × Activity Matrix
6. Synthesis: What The Jobs Tell Us
7. AI Transformation Roadmap (3-phase)
8. Appendix: TAT Distribution

For the full analysis methodology, clustering rules, and formatting requirements, see `.claude/skills/eng-team-analysis/SKILL.md`.

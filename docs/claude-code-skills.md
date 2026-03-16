# Using Claude Code Skills

Claude Code skills live in `.claude/skills/`. Each skill is a self-contained Markdown file that gives Claude a reusable, well-defined behaviour — like a saved workflow you can invoke by name.

## Available skills

| Skill | Slash command | Description |
|-------|---------------|-------------|
| `eng-team-analysis` | `/eng-team-analysis` | Deep-dive engineering team analysis — collects Jira tickets, Slack signals, and meeting transcripts then produces activity clusters, stakeholder maps, JTBD decomposition, AI readiness ratings, and a final report |
| `project-status-summary` | `/project-status-summary` | Aggregate project health signals from Slack, Jira, Confluence, and email into a structured executive summary, then post it to Slack |
| `weekly-bottleneck-report` | `/weekly-bottleneck-report` | Generate an internal engineering report highlighting sprint delays, bottlenecks, over-assignment, QA queue pressure, stale items, and release timeline estimates |
| `estimate-release` | `/estimate-release` | Deep-dive release estimation for a single Jira ticket — analyses subtasks, developer workload, QA queue position, and sprint slippage |
| `openspec-explore` | `/opsx:explore` | Enter explore mode — a thinking partner for ideas, problems, and requirements |
| `openspec-propose` | `/opsx:propose` | Propose a new change and generate all artifacts (proposal, design, tasks) in one step |
| `openspec-apply-change` | `/opsx:apply` | Implement tasks from an existing OpenSpec change |
| `openspec-archive-change` | `/opsx:archive` | Archive a completed change |

## How to use the `eng-team-analysis` skill

**Option A — natural language**

Ask Claude in plain English:

```
Run an engineering team analysis for the tpbank project on the CO board.
```

To extend the time window:

```
Run an engineering team analysis for the payments project on the ENG board for the last 60 days.
```

**Option B — slash command**

```
/eng-team-analysis                                # all defaults: tpbank, CO board, 30 days, summary to your DM
/eng-team-analysis tpbank                         # explicit project, default board and window
/eng-team-analysis tpbank CO                      # explicit project and board, default window
/eng-team-analysis tpbank CO 60                   # 60-day window
/eng-team-analysis payments ENG 90                # different project with 90-day window
/eng-team-analysis tpbank CO 30 @john             # send summary DM to @john
/eng-team-analysis tpbank CO 30 #eng-analysis     # post summary to a channel
```

**Parameters**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `project` | ❌ | `tpbank` | Project name used to filter Jira, Slack, and Drive |
| `board` | ❌ | `CO` | Jira board name or key |
| `window` | ❌ | `30` | Look-back window in days |
| `channel` | ❌ | user's own DM | Slack destination for the condensed summary — a channel (`#name`) or DM handle (`@name`) |

**What it does**

1. Pulls all Jira tickets on the `{board}` board updated in the last `{window}` days (paginated, no cap) and computes TAT metrics per ticket.
2. Searches Slack channels and messages referencing `{project}` to surface requests, decisions, and internal clients.
3. *(Optional)* Reads Google Drive meeting transcripts for context not captured in Jira or Slack.
4. Classifies tickets into 6–8 non-overlapping activity clusters named by what the work accomplishes.
5. Maps every stakeholder who requested work, grouped by role, with implicit success measures.
6. Extracts Jobs-To-Be-Done per cluster, re-clusters them by demand-side structure, and produces a JTBD × Activity matrix.
7. Rates each JTBD cluster HIGH / MEDIUM / LOW for AI automation potential and maps it to a three-phase roadmap.
8. Generates a full PDF-ready report saved to `local_files/eng-analysis/{project}-{YYYY-MM-DD}-eng-team-analysis.md`.
9. **Sends a condensed summary to Slack via the `channel` parameter** (defaults to the user's own DM if `channel` is not provided). If Slack is not connected, displays the summary in chat.

**Prerequisites**

The skill requires MCP integrations for the sources you want to search:
- **Jira** *(required)* — [Atlassian Jira MCP server](https://github.com/sooperset/mcp-atlassian)
- **Slack** *(optional)* — [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack); enriches stakeholder and request data
- **Google Drive** *(optional)* — Google Drive MCP; surfaces meeting transcript context
- **Google Calendar** *(optional)* — enriches deadline and milestone data

If Jira is not connected, the skill stops and asks you to connect it. All other sources are optional — if unavailable the skill notes the gap and continues.

## How to use the `project-status-summary` skill

**Option A — natural language**

Ask Claude in plain English:

```
Please summarize the current project status for the project: "tpbank", on the CO board.
```

To look back further than the default 7 days:

```
Please summarize the project "tpbank" on the CO board for the last 14 days.
```

**Option B — slash command**

```
/project-status-summary tpbank CO                     # last 7 days, send to your DM
/project-status-summary tpbank CO 14                  # last 14 days
/project-status-summary tpbank CO 7 #tpbank-updates   # post to a channel
/project-status-summary tpbank CO 7 @john             # DM to @john
/project-status-summary "my project" CO               # multi-word project name
```

**Parameters**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `project` | ✅ | — | Project name (use quotes for multi-word names) |
| `board` | ✅ | — | Jira board key (e.g., `CO`) |
| `days` | ❌ | `7` | Number of days to look back |
| `channel` | ❌ | user's own DM | Slack destination — a channel (`#name`) or DM handle (`@name`) |

**What it does**

1. Searches Slack (channels and messages), Jira (board tickets and sprint data), Confluence (updated pages), and email for activity matching the project name within the specified time window.
2. Produces a structured executive summary with numbered source references on every bullet.
3. Lists all sources in a table with direct links (Slack message permalinks, Jira ticket URLs, Confluence page URLs, email subject lines).
4. Sends the formatted summary to your Slack DM (or a channel you specify).

**Prerequisites**

The skill requires MCP integrations for the sources you want to search. Configure the relevant integrations in your Claude Code session before invoking the skill:
- **Slack** — [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack)
- **Jira** — [Atlassian Jira MCP server](https://github.com/sooperset/mcp-atlassian)
- **Confluence** — [Atlassian Confluence MCP server](https://github.com/sooperset/mcp-atlassian) (same server as Jira)
- **Email** — [Gmail MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/gmail) or [Outlook MCP server](https://github.com/modelcontextprotocol/servers)

If a source is not connected, the skill notes it clearly and continues with the remaining sources.

## How to use the `weekly-bottleneck-report` skill

**Option A — natural language**

Ask Claude in plain English:

```
Generate the weekly bottleneck report for the AGI project.
```

To analyse multiple projects or extend the time window:

```
Generate the weekly bottleneck report for AGI and CO for the last 14 days.
```

**Option B — slash command**

```
/weekly-bottleneck-report AGI                      # single project, last 7 days, DM to yourself
/weekly-bottleneck-report AGI,CO                   # multiple projects, last 7 days
/weekly-bottleneck-report AGI 14                   # last 14 days
/weekly-bottleneck-report AGI 7 337                # explicit board ID
/weekly-bottleneck-report AGI 7 337 @john          # send Slack DM to @john
/weekly-bottleneck-report AGI 7 337 #eng-leads     # post to a Slack channel
```

**Parameters**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `projects` | ✅ | — | One or more Jira project keys, comma-separated (e.g., `AGI` or `AGI,CO`) |
| `days` | ❌ | `7` | Time window in days for stale-item detection |
| `board` | ❌ | auto-detect | Jira board ID; auto-detected from the first project if omitted |
| `recipient` | ❌ | user's own DM | Slack destination for the condensed summary — a channel (`#name`) or DM handle (`@name`) |

**What it does**

1. Runs 5 targeted JQL queries against the active sprint (high-priority delayed items, blocked items, active development workload, QA queue, stale issues).
2. Cross-references Confluence for weekly update context, GitHub for stale PRs, and Calendar for upcoming deadlines.
3. Analyses sprint labels to detect slippage (2+ sprint labels = at risk; 3+ = significantly delayed).
4. Flags over-assigned developers (3+ concurrent active tickets) and context-switching risks.
5. Estimates release timelines per workstream using the `/estimate-release` methodology.
6. Produces tiered recommendations (immediate, this sprint, process improvements).
7. Saves the full report to `local_files/Jira/weekly/<YYYY-MM-DD>-bottleneck-report.md`.
8. Sends a condensed summary to the `recipient` via Slack DM (defaults to your own DM if not provided).

**Prerequisites**

The skill requires MCP integrations for the sources you want to search:
- **Jira** *(required)* — [Atlassian Jira MCP server](https://github.com/sooperset/mcp-atlassian)
- **Confluence** *(optional)* — same server as Jira; enriches workstream groupings
- **GitHub** *(optional)* — [github-mcp-server](https://github.com/github/github-mcp-server); surfaces stale PRs
- **Calendar** *(optional)* — Google Calendar or Outlook MCP; surfaces upcoming deadlines
- **Slack** *(optional)* — [Slack MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/slack); delivers the condensed summary

If Slack is not connected, the skill displays the summary in-chat. If any other optional source is not connected, the skill notes it and continues with remaining sources. Jira is required.

## How to use the `estimate-release` skill

**Usage**

```
/estimate-release AGI-5531
```

**What it does** (9 steps)

1. **Fetches the ticket** — status, priority, sprint labels, time in current status
2. **Analyses subtasks** — completion table, identifies blockers, flags unassigned work
3. **Assesses developer workload** — competing tickets, bandwidth for bug fixes
4. **Identifies QA assignee** — from `customfield_10747` or parent/linked tickets
5. **Assesses QA workload** — how many other tickets the QA is testing
6. **Checks QA queue position** — items ahead in the "Ready for Test" queue
7. *(Step 7 intentionally reserved)*
8. **Analyses sprint slippage** — number of sprints the ticket has been planned across
9. **Produces the estimate** — structured output with:
   - Current state summary table
   - Blocker list with owners and resolution estimates
   - Developer availability assessment (risk level)
   - Timeline estimate table (blocker resolution + QA + bug fixes = total)
   - Key risks ranked by impact
   - Actionable recommendations

**Prerequisites**

- **Jira** *(required)* — [Atlassian Jira MCP server](https://github.com/sooperset/mcp-atlassian)

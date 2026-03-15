---
name: eng-team-analysis
description: Deep-dive engineering team analysis — collects Jira tickets, Slack signals, and meeting transcripts for a configurable project and time window, then produces activity clusters, stakeholder maps, JTBD decomposition, AI readiness ratings, and a final PDF-ready report.
license: MIT
compatibility: Requires MCP integrations for Jira, Slack, and optionally Google Drive and Google Calendar.
metadata:
  author: awesome-ai-skills
  version: "1.0.0"
  generatedBy: "1.0.0"
---

Run a full engineering team analysis. Collect data from Jira, Slack, and meeting transcripts, then produce activity clusters, stakeholder maps, Jobs-To-Be-Done decomposition, AI transformation readiness ratings, and a final report.

---

## STYLE RULES

Be concise. Use plain words. No filler. Every sentence must carry insight or evidence. Do not summarise what you are about to do — just do it. Do not lose any insights in pursuit of brevity; compress the language, not the substance.

---

## Inputs

| Parameter | Description | Default | Example |
|-----------|-------------|---------|---------|
| `project` | Project name used to filter Jira, Slack, and Drive | `tpbank` | `payments` |
| `board` | Jira board name or key | `CO` | `ENG` |
| `window` | Look-back window in days | `30` | `60` |

If any required parameter is missing, ask before proceeding:

> "Which project, board, and time window would you like to analyse? (defaults: project=tpbank, board=CO, window=30 days)"

---

## PHASE 1: Data Collection

### 1.1 Jira — All tickets, no cap

Pull every ticket in the `{board}` board updated or created in the last `{window}` days. Do not cap results — paginate as needed.

JQL template:
```
project = "{project}" AND updated >= -{window}d ORDER BY created DESC
```

For each ticket extract: key, summary, status, type, priority, assignee, reporter, created date, each status-transition timestamp (e.g. To Do → In Progress → Review → Done/Cancelled), resolution date, linked issues, comment count, sprint labels.

### 1.2 Jira — Turnaround Time (TAT)

For every ticket, compute stage durations:

- **Ticket → First Status Change**: created date → first transition out of "To Do"
- **In Progress Duration**: cumulative time in any "In Progress" or "In Development" status
- **Review/QA Duration**: time in any Code Review, Ready for Test, or Testing status
- **Total Cycle Time**: created → resolved/closed
- **Sprint Slippage**: count of distinct sprint labels — 1 = on track, 2 = ⚠️ at risk, 3+ = 🔴 significantly delayed

Report TAT by cluster (median, P75, P90) and by assignee.

### 1.3 Slack — Project channels

Search Slack channels whose name contains `{project}` and any channel whose messages reference `{project}`, going back `{window}` days.

For each channel: capture message volume, key threads, who initiated requests, who responded, and any decisions or blockers surfaced. Identify the internal clients — the people asking the engineering team for something — and record their name and role.

### 1.4 Google Drive — Meeting transcripts *(optional)*

Search Google Drive for meeting recordings or transcripts involving team members in the last `{window}` days. Keywords: project name, "engineering", "sprint", "architecture", "release", plus known team member names if available.

Read full transcripts rather than auto-summaries. Extract: decisions made, action items assigned to the team, stakeholder requests and the language they used, and strategic context not captured in Jira or Slack.

If Google Drive is not accessible, note the gap and continue.

---

## PHASE 2: Activity Clustering

Classify all tickets into activity clusters. Aim for **6–8 non-overlapping clusters** where each ticket belongs to exactly one and the cluster boundaries are crisp.

Cluster naming rules:
- Name clusters by **what the work accomplishes**, not the tool or technology used.
- Merge clusters that share the same core activity — do not create separate clusters for the same underlying job done for different clients.
- Split clusters that contain fundamentally different jobs even if they involve the same codebase.

Suggested starting clusters for engineering teams (adapt to what the data actually shows):

| Cluster | Core activity |
|---------|--------------|
| **Feature Delivery** | Building new product capabilities requested by Product or business stakeholders |
| **Bug Fixing & Reliability** | Resolving defects and preventing recurrence; keeping the system trustworthy |
| **Data & Integration Pipelines** | Ingesting, transforming, and exposing data across systems |
| **Platform & Infrastructure** | Improving developer tooling, CI/CD, observability, and deployment reliability |
| **Tech Debt & Refactoring** | Reducing complexity, improving maintainability, unblocking future delivery |
| **Security & Compliance** | Hardening access controls, auditing, and meeting regulatory requirements |
| **Analytics & Reporting** | Producing dashboards, metrics, or data products consumed by internal or external stakeholders |
| **Knowledge & Process** | Documentation, onboarding, runbooks, process improvements |

For each cluster provide:
- Ticket count (`{window}` days) and ticket count (last 30 days if `window` > 30)
- Input → Process → Output description
- TAT distribution (median, P75, P90)
- Primary assignees and their share of total cluster tickets
- Top 3 requesters (from Jira reporter field + Slack thread initiators)

---

## PHASE 3: Stakeholder Segmentation

Identify every person who requested work from the engineering team. Sources: Jira reporter field, Slack thread initiators, meeting transcript action items.

Group requesters by role:

| Stakeholder | Role | Example requests | Implicit need |
|-------------|------|-----------------|---------------|
| CEO / Exec | Strategic oversight | "When will X be live?" | Board-ready narrative, confidence |
| CTO | Technical leadership | Architecture decisions, trade-off questions | Infrastructure confidence |
| Product Manager | Feature ownership | "Can we add Y?" | Predictable delivery, clear scope |
| QA / Test | Quality gate | Bug reports, regression concerns | Stable releases, fast turnaround |
| Data / Analytics | Internal consumer | Data access, schema changes | Reliable, documented data contracts |
| External clients | Business partners | Integration questions, SLA concerns | Trust, uptime, responsiveness |
| Other engineering teams | Internal | API changes, dependency questions | Continuity, backward compatibility |
| Operations / SRE | Reliability | Incident response, runbook gaps | Fast resolution, proactive alerting |

For each stakeholder group:
- How many requests in `{window}` days?
- What types of work do they request? (map to clusters from Phase 2)
- What is the typical TAT for their requests?
- What language do they use? (reveals whether they want speed, accuracy, narrative, or autonomy)
- What is the implicit success measure? (the quality bar they would name if forced to articulate what "great" looks like)

---

## PHASE 4: Jobs To Be Done Decomposition

### 4.1 Extract JTBDs from activity clusters

For each activity cluster from Phase 2, identify every distinct JTBD using the format:

> **When** [situation], **I want to** [motivation], **so I can** [outcome].

One activity cluster may contain multiple JTBDs. One JTBD may span multiple clusters. That is expected.

For each JTBD provide:
- **Functional dimension**: what needs to happen, concretely
- **Emotional dimension**: what the requester is feeling — urgency, credibility pressure, anxiety about breakage, desire for autonomy
- **Implicit success measure**: the quality bar the requester cannot articulate. What separates output they accept from output that makes them trust the team more. Name what was previously only felt.
- **Good vs. Great**: one sentence each. What does competent look like? What does exceptional look like?
- **Evidence**: specific Jira ticket keys and Slack threads that ground this JTBD in real data

### 4.2 Re-cluster JTBDs

After extracting all JTBDs, group them into JTBD clusters — independent of the original activity clusters. These represent the true demand-side structure of the engineering team's work.

Expected: 5–8 distinct JTBD clusters. Map each back to which activity clusters and which stakeholder groups it serves.

Produce a matrix: JTBD clusters (rows) × Activity clusters (columns), with ticket counts in cells. This shows where the same job is fulfilled through different activities, and where different jobs share the same activity.

---

## PHASE 5: AI Transformation Readiness

For each JTBD cluster (not activity cluster):
- Rate: **HIGH / MEDIUM / LOW** for AI automation potential
- Specify what AI can do (the mechanical, repetitive, or pattern-matching parts)
- Specify what remains human (judgment, taste, context, and relationship)
- Estimate % of current effort that could be automated or augmented
- Identify the implementation phase:
  - **Phase 1 (months 1–3)**: automate repetitive, well-defined jobs
  - **Phase 2 (months 3–6)**: augment judgment-heavy jobs with AI-assisted recommendations
  - **Phase 3 (months 6–12)**: elevate strategic jobs — AI handles the floor so humans can reach for the ceiling

---

## PHASE 6: Report Generation

Generate a PDF-ready report. The narrative arc is: **evidence → meaning → action**.

### Structure

1. **Executive Summary** — 1 page max. Key numbers, biggest insight, top risk.
2. **Activity Clusters** — table with ticket counts, TAT, top assignees. No prose unless it adds insight the table cannot convey.
3. **Stakeholder Map** — who asks for what, how often, implicit needs.
4. **JTBD Clusters** — the core section. Each JTBD gets functional, emotional, and implicit dimensions + Good vs. Great + evidence.
5. **JTBD × Activity Matrix** — crosswalk showing how jobs map to activities.
6. **Synthesis: What The Jobs Tell Us** — step back from individual clusters. Name the meta-patterns: What is the engineering team's core function? Where is effort misallocated relative to value? Which stakeholders are well-served vs. underserved? Which jobs are done three different ways because nobody standardised them? This is the section the CTO and CEO will actually read.
7. **From JTBD to AI Transformation Roadmap** — the action section. The three-phase plan (automate repetitive → augment judgment → elevate strategic) is earned because the synthesis explained why. Map each JTBD cluster to a phase. End with: what does the engineering team become when AI handles the floor so the team can reach for the ceiling?
8. **Appendix: TAT Distribution** — tables for cycle time by cluster and by assignee.

### Style rules

- No sentence longer than 25 words unless quoting evidence.
- No paragraph longer than 4 sentences.
- Use tables over prose whenever possible.
- Bold only for emphasis, not decoration.
- Every claim backed by a Jira ticket key, Slack thread reference, or meeting transcript citation.
- Compress the language, not the substance. Never lose an insight for the sake of brevity.

---

## Output

Save the report to:
```
local_files/eng-analysis/{project}-{YYYY-MM-DD}-eng-team-analysis.md
```

Where `{project}` is the project parameter and `{YYYY-MM-DD}` is today's date in ISO 8601 format (e.g., `2026-03-15`).

Display the report in the chat window. If the Slack MCP is connected, send an executive summary to the user's own DM unless a `channel` parameter was provided.

---

## Guardrails

- **Never fabricate data.** Only report information found in actual sources. If a source is inaccessible, state that clearly and continue.
- **Paginate Jira queries.** Do not stop at the default page size — retrieve all tickets.
- **Cite every claim.** Every bullet in the report must reference at least one Jira ticket key, Slack thread, or transcript.
- **Crisp cluster boundaries.** Each ticket belongs to exactly one activity cluster. If a ticket spans two clusters, assign it to the one that represents the majority of the work and note the overlap.
- **Be specific in recommendations.** Name the ticket, the person, and the exact action — never generic advice.
- **Preserve nuance.** If data is ambiguous or conflicting, say so rather than forcing a clean narrative.
- **Respect privacy.** Only include names and roles relevant to engineering work. Do not include personal information.
- If **Jira is not connected**, stop and inform the user before proceeding.
- If optional sources (Google Drive, Slack, Calendar) are inaccessible, note the gap and continue with remaining sources.

---

## Handling Missing Integrations

| Source | If unavailable |
|--------|---------------|
| Jira | Stop — this is required. Ask the user to connect the Jira MCP and retry. |
| Slack | Note the gap. Skip Phase 1.3. Continue with Jira data only. |
| Google Drive | Note the gap. Skip Phase 1.4. Continue without transcript data. |
| Google Calendar | Note the gap. Omit calendar-derived deadlines from the report. |
| Slack delivery | Display the summary in chat. Note delivery was unavailable. |

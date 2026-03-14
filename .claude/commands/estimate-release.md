---
name: "Estimate Release"
description: "Deep-dive release estimation for a single Jira ticket — analyses subtasks, developer workload, QA queue position, and sprint slippage to produce a data-backed release date estimate"
category: Engineering
tags: [jira, estimation, release, sprint, qa]
---

Produce a structured release estimate for a single Jira ticket.

**Usage**

```
/estimate-release <TICKET-ID>
```

Examples:
- `/estimate-release AGI-5531`
- `/estimate-release CO-1234`

If no ticket ID is provided, ask:
> "Which Jira ticket would you like to estimate? (e.g., AGI-5531)"

---

## Steps

### Step 1 — Fetch the ticket

Retrieve the ticket using the Jira MCP. Capture:
- `summary`, `status`, `priority`, `assignee`
- `labels` — sprint labels (patterns like `AGI_ENG_20XX_SXX_PLANNED/INJECTED`)
- `created`, `updated`, `statuscategorychangedate`
- `parent`, `subtasks`
- `customfield_10747` — QA assignee field

Compute:
- **Days in current status** = today − `statuscategorychangedate`
- **Sprint count** = number of distinct sprint labels on the ticket

### Step 2 — Analyse subtasks

For each subtask, fetch `summary`, `status`, `assignee`.

Build a completion table:

| Subtask | Status | Assignee |
|---------|--------|----------|
| {subtask summary} | {status} | {assignee} |

Flag:
- 🚫 Subtasks still **In Progress** or **Blocked**
- ⚠️ Subtasks with **no assignee**
- Overall completion: `{done}/{total}` subtasks complete

### Step 3 — Assess developer workload

JQL for the developer's active tickets:
```
project = {project} AND sprint in openSprints()
AND assignee = "{assignee}"
AND status in ("In Development", "In Code Review", "Testing", "In Progress", "Code Review")
```

- Count active tickets (including this one)
- **3+ active tickets** = over-assigned → apply **1.5× time multiplier** to estimates
- Flag any items the developer is unlikely to reach before this ticket

### Step 4 — Identify QA assignee

1. Check `customfield_10747` on the ticket
2. If empty, check parent ticket's `customfield_10747`
3. If still empty, check linked tickets for a QA name

If QA is unknown, note the risk and continue.

### Step 5 — Assess QA workload

JQL for the QA's active testing load:
```
project = {project} AND sprint in openSprints()
AND assignee = "{qa_assignee}"
AND status in ("Testing", "Ready for Test", "In Test")
```

Flag QA as **over-loaded** if they have 3+ concurrent testing items.

### Step 6 — Check QA queue position

JQL for the "Ready for Test" queue:
```
project = {project} AND sprint in openSprints()
AND status = "Ready for Test"
ORDER BY priority ASC, updated ASC
```

Find this ticket's position in the queue (or the expected position once it reaches "Ready for Test"). Add **2–3 days per item ahead** in the queue.

### Step 8 — Analyse sprint slippage

*(Step 7 is intentionally reserved.)*

Count distinct sprint labels on the ticket:
- **1 sprint label** = no slippage
- **2 sprint labels** = at risk — add **30% time buffer**
- **3+ sprint labels** = significantly delayed — add **50% time buffer**

### Step 9 — Produce the estimate

#### Current State

| Field | Value |
|-------|-------|
| Status | {status} |
| Days in current status | {n} days |
| Sprint count | {n} sprint(s) |
| Developer | {name} ({n} active tickets) |
| QA | {name or "Unknown"} |
| Subtasks | {done}/{total} complete |

#### Blockers

| Blocker | Owner | Est. Resolution |
|---------|-------|-----------------|
| {description} | {owner} | {days or "Unknown"} |

#### Developer Availability

| Risk Level | Reason |
|------------|--------|
| 🟢 Low / 🟡 Medium / 🔴 High | {explanation — e.g., "2 other active items, normal load"} |

#### Timeline Estimate

| Phase | Estimated Days | Notes |
|-------|----------------|-------|
| Remaining dev / review | {n} days | Based on current status |
| QA queue wait | {n} days | {n} items ahead at 2–3 days each |
| QA testing | {n} days | Includes bug-fix cycles |
| Buffer (slippage / workload) | {n} days | {reason} |
| **Total** | **{n} business days** | **Est. release: {YYYY-MM-DD}** |

#### 🔑 Key Risks (ranked by impact)

1. {Risk 1 — e.g., "QA unknown — may delay testing start"}
2. {Risk 2}
3. …

#### ✅ Recommendations

- {Specific, actionable recommendation — name the ticket, person, and action}
- …

---

## Guardrails

- Use only real data from Jira — never fabricate assignees, statuses, or dates
- Express all estimates in **business days** (exclude weekends)
- When QA is unknown, mark the QA phase as **TBD** and flag it as a high risk
- When a blocker has no resolution timeline, note **"Unknown"** — do not guess
- Apply the developer 1.5× multiplier only when they have 3+ active tickets
- Apply the sprint slippage buffer only for items with 2+ sprint labels
- State confidence level (High / Medium / Low) with the final estimate

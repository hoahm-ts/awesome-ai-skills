# Claude Desktop — MCP Integration & Custom Skills Guide

This guide is for **non-technical users** who want to get the most out of Claude Desktop by connecting it to the tools you already use every day, and then teaching it new skills tailored to your role.

## Table of Contents

- [What is MCP?](#what-is-mcp)
- [Part 1 — Prerequisites: Connect Claude Desktop to Your Tools](#part-1--prerequisites-connect-claude-desktop-to-your-tools)
  - [Before You Start](#before-you-start)
  - [Where to Edit the Configuration File](#where-to-edit-the-configuration-file)
  - [1. Slack](#1-slack)
  - [2. GitHub](#2-github)
  - [3. Gmail](#3-gmail)
  - [4. Google Calendar](#4-google-calendar)
  - [5. Google Drive](#5-google-drive)
  - [6. Figma](#6-figma)
  - [7. Jira](#7-jira)
  - [8. Confluence](#8-confluence)
  - [Verify Your Connections](#verify-your-connections)
- [Part 2 — Create Custom Skills by Asking Claude](#part-2--create-custom-skills-by-asking-claude)
  - [What is a Skill?](#what-is-a-skill)
  - [How to Ask Claude to Create a Skill](#how-to-ask-claude-to-create-a-skill)
  - [Sample Prompts](#sample-prompts)
    - [Daily Briefing Skill](#daily-briefing-skill)
    - [End-of-Day Summary Skill](#end-of-day-summary-skill)
    - [Meeting Prep Skill](#meeting-prep-skill)
    - [Weekly Report Skill](#weekly-report-skill)
    - [Inbox Triage Skill](#inbox-triage-skill)
  - [Using Your New Skill](#using-your-new-skill)
- [Troubleshooting](#troubleshooting)

---

## A Quick Note on Terminology

This guide uses two related terms:

- **Claude Desktop** — the desktop application you open on your computer. It provides a familiar chat interface and uses Claude Code under the hood.
- **Claude Code** — the underlying technology (available as both a CLI tool and the engine powering Claude Desktop) that supports skills, slash commands, and MCP integrations. When this guide says "Claude Code skill", it means a skill that works in Claude Desktop and the Claude Code CLI alike.

---

## What is MCP?

**MCP (Model Context Protocol)** is the standard way to give Claude access to external tools and services. Think of it as a set of "plug-ins" — once you connect a service, Claude can read from and write to it on your behalf.

---

## Part 1 — Prerequisites: Connect Claude Desktop to Your Tools

### Before You Start

1. **Install Claude Desktop** — download it from [claude.ai/download](https://claude.ai/download) and sign in.
2. **Install Node.js** — most MCP servers require it. Download the LTS version from [nodejs.org](https://nodejs.org/).
3. **Gather your API tokens** — each service below needs an API key or access token. The steps tell you exactly where to get each one.

> **Tip:** You only need to set up the services you actually use. Skip any that don't apply to your role.

---

### Where to Edit the Configuration File

All MCP servers are declared in a single JSON file that Claude Desktop reads on startup.

| Operating System | File location |
|---|---|
| **macOS** | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| **Windows** | `%APPDATA%\Claude\claude_desktop_config.json` |

**How to open and edit it:**

1. Open **Finder** (macOS) or **File Explorer** (Windows).
2. Navigate to the path above.
3. Open `claude_desktop_config.json` with any text editor (e.g. Notepad on Windows, TextEdit on macOS, or [VS Code](https://code.visualstudio.com/)).
4. The file looks like this to start with (create it if it doesn't exist):

```json
{
  "mcpServers": {}
}
```

Each service you add goes inside the `"mcpServers"` block. After saving the file, **quit and relaunch Claude Desktop** for changes to take effect.

---

### 1. Slack

**What you need:** A Slack Bot Token.

**Step 1 — Create a Slack app:**
1. Go to [api.slack.com/apps](https://api.slack.com/apps) and click **Create New App → From scratch**.
2. Name it (e.g. "Claude MCP") and choose your workspace.

**Step 2 — Add permissions:**
1. In the left sidebar click **OAuth & Permissions**.
2. Under **Bot Token Scopes** add:
   - `channels:history`, `channels:read`
   - `groups:history`, `groups:read`
   - `im:history`, `im:read`
   - `mpim:history`, `mpim:read`
   - `users:read`
   - `chat:write` (needed for Claude to send messages)

**Step 3 — Install the app:**
1. Click **Install to Workspace** and approve.
2. Copy the **Bot User OAuth Token** (starts with `xoxb-`).

**Step 4 — Add to Claude:**

Add the following block inside `"mcpServers"`:

```json
"slack": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-slack"],
  "env": {
    "SLACK_BOT_TOKEN": "xoxb-your-token-here",
    "SLACK_TEAM_ID": "T0XXXXXXXXX"
  }
}
```

Replace `xoxb-your-token-here` with your token and `T0XXXXXXXXX` with your Slack workspace ID (visible in the URL when you open Slack in a browser: `app.slack.com/client/T0XXXXXXXXX/`).

---

### 2. GitHub

**What you need:** A GitHub Personal Access Token.

**Step 1 — Create a token:**
1. Go to [github.com/settings/tokens](https://github.com/settings/tokens) → **Generate new token (classic)**.
2. Give it a descriptive name (e.g. "Claude Desktop").
3. Select scopes: `repo`, `read:org`, `read:user`.
4. Click **Generate token** and copy it immediately — you won't see it again.

**Step 2 — Add to Claude:**

```json
"github": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_your-token-here"
  }
}
```

---

### 3. Gmail

**What you need:** Google Cloud credentials with Gmail API access.

**Step 1 — Create a Google Cloud project:**
1. Go to [console.cloud.google.com](https://console.cloud.google.com/) and create a new project.
2. In the left menu go to **APIs & Services → Library**, search for **Gmail API**, and click **Enable**.

**Step 2 — Create OAuth credentials:**
1. Go to **APIs & Services → Credentials → Create Credentials → OAuth client ID**.
2. Choose **Desktop app**, name it, and click **Create**.
3. Download the credentials JSON file and save it somewhere safe (e.g. `~/claude-credentials/gmail.json`).

**Step 3 — Add to Claude:**

```json
"gmail": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-gmail"],
  "env": {
    "GMAIL_CREDENTIALS_PATH": "/Users/yourname/claude-credentials/gmail.json",
    "GMAIL_TOKEN_PATH": "/Users/yourname/claude-credentials/gmail-token.json"
  }
}
```

> The first time you ask Claude to access your Gmail, a browser window will open asking you to sign in and authorize access. This only happens once.

---

### 4. Google Calendar

**What you need:** The same Google Cloud project as Gmail (or a new one).

**Step 1 — Enable the Calendar API:**
1. In your Google Cloud project go to **APIs & Services → Library**, search for **Google Calendar API**, and click **Enable**.

**Step 2 — Create credentials** (if you haven't already for Gmail):
Follow the same steps as Gmail Step 2, saving the file as `calendar.json`.

**Step 3 — Add to Claude:**

```json
"google-calendar": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-google-calendar"],
  "env": {
    "GOOGLE_CREDENTIALS_PATH": "/Users/yourname/claude-credentials/calendar.json",
    "GOOGLE_TOKEN_PATH": "/Users/yourname/claude-credentials/calendar-token.json"
  }
}
```

---

### 5. Google Drive

**What you need:** The same Google Cloud project (enable Drive API).

**Step 1 — Enable the Drive API:**
1. In your Google Cloud project go to **APIs & Services → Library**, search for **Google Drive API**, and click **Enable**.

**Step 2 — Add to Claude:**

```json
"google-drive": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-gdrive"],
  "env": {
    "GDRIVE_CREDENTIALS_PATH": "/Users/yourname/claude-credentials/drive.json",
    "GDRIVE_TOKEN_PATH": "/Users/yourname/claude-credentials/drive-token.json"
  }
}
```

---

### 6. Figma

**What you need:** A Figma Personal Access Token.

**Step 1 — Create a token:**
1. In Figma click your **profile picture → Settings**.
2. Scroll to **Personal access tokens** and click **Generate new token**.
3. Name it (e.g. "Claude Desktop") and copy the token.

**Step 2 — Add to Claude:**

```json
"figma": {
  "command": "npx",
  "args": ["-y", "figma-developer-mcp", "--stdio"],
  "env": {
    "FIGMA_API_KEY": "figd_your-token-here"
  }
}
```

---

### 7. Jira

**What you need:** An Atlassian API token and your Jira Cloud URL.

**Step 1 — Create an API token:**
1. Go to [id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens).
2. Click **Create API token**, name it (e.g. "Claude Desktop"), and copy the token.

**Step 2 — Find your Jira URL:**
Your Jira URL looks like: `https://yourcompany.atlassian.net`

**Step 3 — Add to Claude:**

```json
"jira": {
  "command": "npx",
  "args": ["-y", "mcp-atlassian"],
  "env": {
    "JIRA_URL": "https://yourcompany.atlassian.net",
    "JIRA_USERNAME": "your.email@company.com",
    "JIRA_API_TOKEN": "your-api-token-here"
  }
}
```

---

### 8. Confluence

**What you need:** The same Atlassian API token as Jira.

> **Good news:** Confluence uses the same MCP server as Jira (`mcp-atlassian`). If you've already added Jira, just extend that block:

```json
"jira-confluence": {
  "command": "npx",
  "args": ["-y", "mcp-atlassian"],
  "env": {
    "JIRA_URL": "https://yourcompany.atlassian.net",
    "JIRA_USERNAME": "your.email@company.com",
    "JIRA_API_TOKEN": "your-api-token-here",
    "CONFLUENCE_URL": "https://yourcompany.atlassian.net/wiki",
    "CONFLUENCE_USERNAME": "your.email@company.com",
    "CONFLUENCE_API_TOKEN": "your-api-token-here"
  }
}
```

---

### Verify Your Connections

After saving the config file and relaunching Claude Desktop:

1. Open a new chat and type:

   ```
   Which tools and integrations do you have access to right now?
   ```

2. Claude will list all connected MCP servers. Confirm the ones you configured appear in the list.

3. Test each connection with a quick question, for example:
   - **Slack:** `What are the most recent messages in the #general channel?`
   - **GitHub:** `List my open pull requests.`
   - **Gmail:** `What are my unread emails from today?`
   - **Google Calendar:** `What meetings do I have tomorrow?`
   - **Jira:** `Show me tickets assigned to me.`

---

## Part 2 — Create Custom Skills by Asking Claude

### What is a Skill?

A **skill** is a saved, reusable workflow for Claude — like a recipe it can follow whenever you need it. Once created, you can trigger it with a simple slash command (e.g. `/briefing-my-day`) instead of typing a long prompt every time.

Skills live in `.claude/skills/` inside your project folder and are written in plain Markdown, so you can read and tweak them yourself.

---

### How to Ask Claude to Create a Skill

The easiest way is to describe what you want in plain English and ask Claude to turn it into a skill file. Here is a template you can follow:

```
I want you to create a Claude Code skill for me.

The skill should:
- [Describe what it does, step by step]
- [Mention which tools it uses, e.g. Slack, Gmail, Google Calendar]
- [Describe what the final output should look like]

Please save it as `.claude/skills/<skill-name>/<skill-name>.md`
and register it in `.claude/commands/<skill-name>.md` so I can
trigger it with the slash command `/<skill-name>`.
```

---

### Sample Prompts

Copy any of the prompts below directly into Claude Desktop and press Enter. Claude will create the skill files for you automatically.

---

#### Daily Briefing Skill

**What it does:** Every morning, Claude checks your calendar, emails, and Slack messages then gives you a structured summary of your day.

**Prompt to give Claude:**

```
I want you to create a Claude Code skill called "briefing-my-day".

The skill should:
1. Check Google Calendar for all events scheduled for today and tomorrow.
2. Check Gmail for unread emails received in the last 24 hours and highlight
   anything marked urgent or from my manager.
3. Check Slack for any messages that mention my name or are sent directly
   to me in the last 12 hours.
4. Produce a structured daily briefing with three sections:
   - "📅 My Schedule Today" — list every meeting with time, title,
     and a one-sentence agenda if available.
   - "📬 Emails to Action" — bullet list of emails that need a reply
     or decision, with sender, subject, and a one-line summary.
   - "💬 Slack Highlights" — bullet list of important direct messages
     or mentions, with sender and a one-line summary.
5. End with a "⚡ Top 3 Priorities" section — suggest the three most
   important things I should do first today based on urgency and deadlines.

Please save it as:
- `.claude/skills/briefing-my-day/briefing-my-day.md`
- `.claude/commands/briefing-my-day.md`

Make the slash command `/briefing-my-day`.
```

---

#### End-of-Day Summary Skill

**What it does:** At the end of your work day, Claude compiles what you accomplished and what needs attention tomorrow.

**Prompt to give Claude:**

```
Create a Claude Code skill called "end-of-day-summary".

The skill should:
1. Check Google Calendar for all events that occurred today and note which
   ones actually happened (were not cancelled).
2. Check Gmail for emails I sent today to track what I communicated.
3. Check Slack for messages I sent today in public channels.
4. Check Jira for any tickets I updated or commented on today.
5. Produce an end-of-day summary with these sections:
   - "✅ What I Accomplished Today" — based on calendar, Jira activity,
     and emails/Slack sent.
   - "📋 Open Items Carrying Forward" — any tasks or emails that were
     not resolved and need attention tomorrow.
   - "📅 What's Coming Tomorrow" — a preview of tomorrow's calendar events.
6. Send this summary to my Slack DM as a formatted message.

Save it as:
- `.claude/skills/end-of-day-summary/end-of-day-summary.md`
- `.claude/commands/end-of-day-summary.md`

Slash command: `/end-of-day-summary`
```

---

#### Meeting Prep Skill

**What it does:** Before a meeting, Claude pulls together everything you need — attendees, agenda, relevant emails, and linked Jira tickets.

**Prompt to give Claude:**

```
Create a Claude Code skill called "meeting-prep".

The skill accepts one input: a meeting title or calendar event name.

The skill should:
1. Search Google Calendar for the next upcoming event matching the given
   meeting title.
2. List all attendees and their names.
3. Check Gmail for recent email threads with those attendees (last 7 days).
4. Search Confluence for any pages related to the meeting topic.
5. Search Jira for open tickets related to the meeting topic.
6. Produce a one-page meeting brief with:
   - "📋 Meeting Overview" — date, time, attendees, and the calendar description.
   - "📧 Recent Email Context" — key points from recent email exchanges.
   - "📄 Relevant Documents" — links to Confluence pages found.
   - "🎫 Related Jira Tickets" — list of open tickets with status.
   - "❓ Suggested Questions or Talking Points" — 3–5 questions Claude
     thinks are worth raising based on the context found.

Save it as:
- `.claude/skills/meeting-prep/meeting-prep.md`
- `.claude/commands/meeting-prep.md`

Slash command: `/meeting-prep`

Usage example: `/meeting-prep "Q2 Planning"`
```

---

#### Weekly Report Skill

**What it does:** Generates a weekly status report you can share with your manager or team by pulling data from Jira, Confluence, Slack, and your calendar.

**Prompt to give Claude:**

```
Create a Claude Code skill called "weekly-report".

The skill should cover the past 7 days and:
1. Search Jira for all tickets I worked on or that changed status this week.
2. Search Confluence for any pages I created or edited this week.
3. Search Slack for key announcements I made or that were made in project
   channels I follow.
4. Check Google Calendar for major meetings I attended.
5. Produce a weekly report with:
   - "🏆 Highlights This Week" — biggest achievements, decisions made,
     or milestones reached.
   - "🎫 Jira Progress" — table of tickets: title, status change, and
     a one-line note.
   - "📝 Documents & Decisions" — Confluence pages created/updated.
   - "🚧 Blockers & Risks" — anything that slowed me down or is at risk.
   - "📅 Next Week's Focus" — inferred from open Jira tickets and
     upcoming calendar events.
6. Save the report as a Markdown file at
   `local_files/weekly-reports/YYYY-MM-DD-weekly-report.md`.
7. Send a condensed version (Highlights + Next Week's Focus) to my
   Slack DM.

Save it as:
- `.claude/skills/weekly-report/weekly-report.md`
- `.claude/commands/weekly-report.md`

Slash command: `/weekly-report`
```

---

#### Inbox Triage Skill

**What it does:** Sorts your unread emails into priority buckets and drafts suggested replies for the urgent ones.

**Prompt to give Claude:**

```
Create a Claude Code skill called "inbox-triage".

The skill should:
1. Fetch all unread emails from Gmail.
2. Categorise each email into one of:
   - 🔴 Urgent — needs a reply today
   - 🟡 Important — needs a reply this week
   - 🟢 FYI — no reply needed, just informational
   - 🗑️ Archive — newsletters, notifications, or irrelevant items
3. For each 🔴 Urgent email, draft a short suggested reply (2–4 sentences).
4. Present the triage results as a structured list grouped by category.
5. Ask me: "Shall I send any of the drafted replies?" and wait for
   my confirmation before sending anything.

Save it as:
- `.claude/skills/inbox-triage/inbox-triage.md`
- `.claude/commands/inbox-triage.md`

Slash command: `/inbox-triage`
```

---

### Using Your New Skill

Once Claude has created the skill files, you can invoke your skill two ways:

**Option A — slash command** (fastest):

```
/briefing-my-day
```

**Option B — plain English**:

```
Please run my daily briefing.
```

> **Tip:** You can always ask Claude to refine a skill. For example: *"Update the `/briefing-my-day` skill to also check Google Drive for files shared with me in the last 24 hours."*

---

## Troubleshooting

| Problem | What to check |
|---|---|
| Claude says a tool is not available | Check `claude_desktop_config.json` for typos and make sure you saved and relaunched Claude Desktop. |
| "Authentication failed" error | Your API token may have expired or be missing required scopes. Re-generate and update `claude_desktop_config.json`. |
| A slash command is not recognised | Confirm the command file exists in `.claude/commands/` and restart Claude Desktop. |
| Google OAuth browser window doesn't open | Make sure Node.js is installed and the credentials JSON path in the config is the correct absolute path. |
| Claude sends a message without asking | Review the skill file and add an explicit instruction: *"Always ask for my confirmation before sending anything."* |

---

> **Need help?** Ask Claude directly: *"I'm having trouble connecting [service name]. Can you help me debug my MCP configuration?"* — Claude can read your config and suggest fixes.

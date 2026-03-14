#!/usr/bin/env bash
# init-repo-settings.sh — Apply standard repository settings for a new repository.
#
# Usage:
#   bash .github/init-repo-settings.sh
#
# Prerequisites:
#   - GitHub CLI (gh) installed and authenticated
#   - Run from inside the repository you want to initialise

set -euo pipefail

REPO=$(gh repo view --json nameWithOwner -q '.nameWithOwner')

echo "Configuring repository settings for: $REPO"

# Pull Request settings
gh repo edit "$REPO" \
  --allow-merge-commit \
  --merge-commit-title "PR_TITLE" \
  --allow-squash-merge \
  --squash-merge-commit-title "PR_TITLE" \
  --allow-rebase-merge \
  --allow-update-branch \
  --delete-branch-on-merge

# Require contributors to sign off on web-based commits
gh api \
  --method PATCH \
  -H "Accept: application/vnd.github+json" \
  "/repos/$REPO" \
  -f web_commit_signoff_required=true

echo "✅ Repository settings configured successfully."

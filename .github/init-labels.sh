#!/usr/bin/env bash
# init-labels.sh — Create standard GitHub issue labels for a new repository.
#
# Usage:
#   bash .github/init-labels.sh
#
# Prerequisites:
#   - GitHub CLI (gh) installed and authenticated
#   - Run from inside the repository you want to initialise
#
# Note: --force updates an existing label's colour/description in place.
# Any custom colour or description changes you have made to a same-named label
# will be overwritten when this script is re-run.

set -euo pipefail

gh label create feature            --color 0075ca --description "New feature or functionality"      --force
gh label create spec               --color e4e669 --description "Specification or requirement"       --force
gh label create chore              --color ededed --description "Routine maintenance task"           --force
gh label create fix                --color d73a4a --description "Bug fix"                            --force
gh label create docs               --color 0075ca --description "Documentation change"               --force
gh label create refactor           --color c5def5 --description "Code refactoring, no behaviour change" --force
gh label create test               --color bfd4f2 --description "Adding or updating tests"           --force
gh label create spec-archive       --color f9d0c4 --description "Archived specification"             --force
gh label create enhancement        --color a2eeef --description "Improvement to existing feature"    --force
gh label create security           --color ff0000 --description "Security vulnerability or hardening" --force
gh label create migration/database --color 8a2be2 --description "Database migration required"        --force

echo "✅ All standard labels created successfully."

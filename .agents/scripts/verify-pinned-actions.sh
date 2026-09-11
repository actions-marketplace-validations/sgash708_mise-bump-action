#!/usr/bin/env bash
# Verifies that every `uses: owner/repo@SHA # vX` line in .github/workflows/
# still has a SHA that GitHub actually resolves the comment's tag to. Catches
# the case where the tag comment and the SHA drift out of sync (e.g. someone
# edits one but not the other by hand). Requires `gh` to be authenticated.
#
# Usage: .agents/scripts/verify-pinned-actions.sh
set -euo pipefail

cd "$(dirname "$0")/../.."

status=0
while IFS= read -r line; do
  file="${line%%:*}"
  rest="${line#*:}"
  repo=$(echo "$rest" | sed -E 's|.*uses: ([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)@.*|\1|')
  sha=$(echo "$rest" | sed -E 's|.*@([a-f0-9]{40}) .*|\1|')
  tag=$(echo "$rest" | sed -E 's|.*# (v[0-9.]+)$|\1|')

  resolved=$(gh api "repos/${repo}/git/ref/tags/${tag}" --jq '.object.sha' 2>/dev/null || true)
  obj_type=$(gh api "repos/${repo}/git/ref/tags/${tag}" --jq '.object.type' 2>/dev/null || true)
  if [ "$obj_type" = "tag" ]; then
    resolved=$(gh api "repos/${repo}/git/tags/${resolved}" --jq '.object.sha' 2>/dev/null || true)
  fi

  if [ -z "$resolved" ]; then
    echo "WARN  ${file}: could not resolve ${repo}@${tag} via GitHub API (skipped)"
    continue
  fi
  if [ "$resolved" != "$sha" ]; then
    echo "FAIL  ${file}: ${repo}@${tag} is pinned to ${sha} but that tag currently resolves to ${resolved}"
    status=1
  else
    echo "OK    ${file}: ${repo}@${tag} -> ${sha}"
  fi
done < <(grep -rnoE "uses: [A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+@[a-f0-9]{40} # v[0-9.]+" .github/workflows/*.yml)

exit $status

#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Unreleased
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🤖
# @raycast.argument1 { "type": "dropdown", "placeholder": "Repo", "data": [{ "title": "Backend", "value": "augusthealth/august-backend" }, { "title": "Frontend", "value": "augusthealth/august-frontend" }]  }

# Documentation:
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

set -euxo pipefail

main() {
  repo=$1
  if ! command -v gh &> /dev/null; then
    echo "gh could not be found"
    exit 1
  fi

  latest_release=$(gh release list --exclude-drafts --exclude-pre-releases --json tagName -L 1 -R "$repo" | jq '.[0].tagName' | sed s/\"//g)

  open "https://github.com/$repo/compare/$latest_release...master"
}

main $1

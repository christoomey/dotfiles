#!/bin/bash

set -eo

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title GitHub PRs
# @raycast.mode silent

# Optional parameters:
# @raycast.icon icons/github-logo.png

# Documentation:
# @raycast.description Open filtered GitHub PRs search
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

date_range="$(gdate --date='5 days ago' +%Y-%m-%d)..$(gdate +%Y-%m-%d)"

open "https://github.com/augusthealth/august-backend/pulls?q=\
  is:pr+is:open+\
  -is:draft+\
  -author:christoomey+\
  -reviewed-by:@me+\
  -author:app%2Fscala-steward-augusthealth+\
  created:$date_range"

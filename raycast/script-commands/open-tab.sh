#!/usr/bin/env bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Open Tab
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🤖

# Documentation:
# @raycast.description Open the specified URL in Chrome
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com
# @raycast.argument1 { "type": "text", "placeholder": "URL" }


~/bin/open-tab "$1"

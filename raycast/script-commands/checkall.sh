#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Checkall
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🔄

# Documentation:
# @raycast.description Check all work sources
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

# open -a Terminal ~/code/things-report/bin/checkall

open -a iTerm
tmux display-popup -d "$HOME/code/things-report" -xC -yC -w80% -h80% "bin/checkall"

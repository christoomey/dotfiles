#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title August
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🔄

# Documentation:
# @raycast.description August helper scripts
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

open -a iTerm

tmux display-popup -d "$HOME/code/work/august/july" -xC -yC -w60% -h70% -EE "ls bin; zsh"

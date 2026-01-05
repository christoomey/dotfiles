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

tmux switch-client -t july
tmux display-popup -d "$HOME/code/work/august/july" -xC -yC -w70% -h80% -EE "ls bin | fzf --reverse --preview 'bat --color=always bin/{}' | xargs -I {} -o zsh -c './bin/{}'"

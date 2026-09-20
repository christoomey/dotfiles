#!/bin/zsh
set -e
print -r -- "started $2" >>"$1/started.log"
sleep 0.2
print -r -- "done $2" >>"$1/done.log"

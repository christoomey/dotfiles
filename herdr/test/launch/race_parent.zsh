#!/bin/zsh
# Stands in for the popup: detach a worker, log, exit. herdr SIGHUPs this
# process group the moment it is gone.
out=$1
id=$2
"${0:A:h}/../../../bin/herdr-detach" --log "$out/worker.out" -- "${0:A:h}/race_worker.zsh" "$out" "$id"
print -r -- "$(date '+%H:%M:%S') forked $id" >>"$out/parent.log"
exit 0

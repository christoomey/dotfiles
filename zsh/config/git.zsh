compdef g=git
function g {
  if [[ $# -gt 0 ]]; then
    git "$@"
  else
    git status --short --branch
  fi
}

function gitup {
    if git rev-parse --git-dir > /dev/null 2>&1; then
        cd "./$(git rev-parse --show-cdup)"
    else
        return
    fi
}

function list-all-git-refs() {
  find .git/refs  -type f -print -exec cat {} \;
}

_hub() { reply=(browse compare pull-request ci-status) }
compctl -K _hub hub

edit-modified-files-in-tabs() {
  vim -O $(git status --porcelain | sed s/^...//)
}

alias gish=gitsh

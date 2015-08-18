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

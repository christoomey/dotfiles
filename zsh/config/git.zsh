compdef g=git
function g {
  if [[ $# -gt 0 ]]; then
    if [ $1 = "." ]; then
      git status --short --branch .
    else
      git "$@"
    fi
  else
    if behind_master; then yellow_msg 'behind master'; fi
    # behind upstream warning
    git status --short --branch
  fi
}

function yellow_msg {
  echo "\033[33m$1\033[0m"
}

function behind_master {
  test $(git merge-base HEAD master) != $(git rev-parse master)
}

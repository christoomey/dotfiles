_no-untracked-files() {
  [[ -z $(git status --untracked --short) ]]
}

_no-modified-files() {
  git diff --exit-code > /dev/null
}

_git-repo-is-clean() {
  _no-modified-files && _no-untracked-files
}

update-upcase-repos() {
  if _git-repo-is-clean; then
    git checkout master
    git pull
    git fetch production
    git fetch staging
    git checkout -
  else
    echo "Repo has local changes"
    return 1
  fi
}

# if git-repo-is-clean; then
#   echo "CLEAN"
# else
#   echo "it be dirty"
# fi

current_ticket_number() {
  git rev-parse --abbrev-ref HEAD | grep -Eo "\d+"
}

booster() {
  server=$(cat ~/.ssh/config | grep '^Host booster' | cut -d ' ' -f 2 | sed 's/booster_//' | grep -v '*' | fzf --reverse)
  if [ -z "$server" ]; then
    >&2 echo "No server selected"
    return 1
  else
    echo "connecting to $server"
    ssh "booster_$server"
  fi
}

ball() {
  for app in $(printf "%s\n" dev_proxy fulcrum donation admin cas_server backend payment_service); do
    (
      echo "\n----- running in $app"
      cd "$HOME/code/work/current/$app"
      $*
    )
  done
}

bbranches() {
  for app in $(printf "%s\n" dev_proxy fulcrum donation admin cas_server backend payment_service); do
    (
      cd "$HOME/code/work/current/$app"
      branch=$(git rev-parse --abbrev-ref HEAD)
      if [ "$branch" != "master" ]; then
        echo "$app is on $branch"
      fi
    )
  done
}

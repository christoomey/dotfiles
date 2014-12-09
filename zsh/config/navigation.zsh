unsetopt auto_cd # with cdpath enabled, auto_cd gives too many false positives
cdpath=($HOME/code $HOME/code/work $HOME/code/work/current $HOME/code/vim $HOME/code/alfred $HOME)

_cdpath_directories() {
  modified_in_last_days=${1:-999}
  echo "${CDPATH//:/\n}" | while read dir; do
    find -L "$dir" \
      -not -path '*/\.*' \
      -type d \
      -mtime -"$modified_in_last_days" \
      -maxdepth 1
  done
}

_is_a_git_repo() {
  while read dir; do
    if [[ -d "$dir/.git" ]]; then
      basename "$dir"
    fi
  done
}

projects() {
  _cdpath_directories $1 | _is_a_git_repo
}


# Do not autoupdate tmux window titles
export DISABLE_AUTO_TITLE="true"

alias tls="tmux list-sessions"

tm-select-session() {
  project=$(projects | fzf --reverse)
  if [ ! -z "$project" ]; then
    (cd "$project" && tat)
  fi
}

twlist() {
  pgrep -lf tagwatch | cut -d ' ' -f 5 | sed -E "s#$HOME#\~#" | uniq
}

_not_inside_tmux() { [[ -z "$TMUX" ]] }

ensure_tmux_is_running() {
  if _not_inside_tmux; then
    tat
  fi
}

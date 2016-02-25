# Do not autoupdate tmux window titles
log() {
  echo "$*" >> ~/foo.log
}

export DISABLE_AUTO_TITLE="true"

alias tls="tmux list-sessions"

tm-select-session() {
  project=$(projects | fzf --reverse)
  if [ ! -z "$project" ]; then
    (cd "$project" && tat)
  fi
}

spin-up-tagwatchers() {
  for project in $(projects 2); do
    (
      cd "$project" &> /dev/null
      if ! pgrep -q "tagwatch -d $(pwd)"; then
        tagwatch -d "$(pwd)" &
      fi
    )
  done
}

twlist() {
  pgrep -lf tagwatch | cut -d ' ' -f 5 | sed -E "s#$HOME#\~#" | uniq
}

twb() {
  nohup tagwatch -v -d "$(pwd)" &> ~/.tagwatch.log &
}

_not_inside_tmux() {
  [[ -z "$TMUX" ]]
}

ensure_tmux_is_running() {
  if _not_inside_tmux; then
    tat
  fi
}

# Do not autoupdate tmux window titles
export DISABLE_AUTO_TITLE="true"

alias tls="tmux list-sessions"

# Fuzzy select dir from CDPATH, then switch to tmux session for it
function t {
  local project
  local session
  local sessions
  project=$(echo "${CDPATH//:/\n}" | while read dir; do find -L "$dir" -not -path '*/\.*' -type d -maxdepth 1 -exec basename {} \;; done | fzf --reverse)
  sessions=$(tmux list-sessions | awk -F ':' '{print $1}')
  current_session=$(tmux display-message -p '#S')
  if echo $sessions | grep -q "$project"; then
    tmux switch-client -t "$project"
  else
    (cd "$project" && TMUX= tmux new-session -d -s "$project")
    tmux switch-client -t "$project"
  fi
}

# Kill current tmux session and move to next
function tk {
  current_session=$(tmux display-message -p '#S')
  tmux switch-client -n
  tmux kill-session -t "$current_session"
}

# Fuzzy select from other tmux sessions and switch to selection
function ta {
  if _not_inside_tmux; then
    tmux attach
  else
    local current_session
    local session
    local sessions
    sessions=$(tmux list-sessions | awk -F ':' '{print $1}')
    current_session=$(tmux display-message -p '#S')
    session=$(echo "$sessions" | grep -v "^$current_session$" | \
      fzf --select-1 --exit-0 --reverse) &&
    tmux switch-client -t "$session"
  fi
}

# Creat a new tmux session. Use directory name for session if none provided
function tn {
  if [ -z "$1" ]; then
    session_name=$(basename `pwd`)
  else
    session_name=$1
  fi
  if _not_inside_tmux; then
    tmux new-session -d -s $session_name -n vim
    tmux attach-session -t $session_name
  else
    (TMUX= tmux new-session -d -s "$session_name")
    tmux switch-client -t "$session_name"
  fi
}

_not_inside_tmux() { [[ -z "$TMUX" ]] }
_tmux_sessions_list_is_empty() { [[ -z $(tmux ls) ]] }

_ensure_tmux_has_a_session() {
  if _tmux_sessions_list_is_empty; then
    tmux new -d -s $(basename `pwd`)
  fi
}

ensure_tmux_is_running() {
  if _not_inside_tmux; then
    _ensure_tmux_has_a_session
    tmux attach
  fi
}

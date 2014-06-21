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
  local session
  local current_session
  sessions=$(tmux list-sessions | awk -F ':' '{print $1}')
  current_session=$(tmux display-message -p '#S')
  session=$(echo "$sessions" | grep -v "$current_session" | \
    fzf --query="$1" --select-1 --exit-0 --reverse) &&
  tmux switch-client -t "$session"
}

# Creat a new tmux session. Use directory name for session if none provided
function tn {
  if [ -z "$1" ]; then
    session_name=$(basename `pwd`)
  else
    session_name=$1
  fi
  tmux new-session -d -s $session_name -n vim
  tmux attach-session -t $session_name
}

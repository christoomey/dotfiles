_tmux_orientation() {
  width=$(tmux display -p '#{pane_width}')
  height=$(tmux display -p '#{pane_height}')
  normalized_height=$( echo "$height * 2.2" | bc | xargs printf "%.0f")

  if (( normalized_height > width )); then
    echo 'portrait'
  else
    echo 'landscape'
  fi
}

tmux-smart-pane() {
  [[ $(_tmux_orientation) = 'portrait' ]] && orient='-v' || orient='-h'
  eval "tmux split-window $orient $@"
}

_not_inside_tmux() {
  [[ -z "$TMUX" ]]
}

ensure_tmux_is_running() {
  if _not_inside_tmux; then
    tat
  fi
}

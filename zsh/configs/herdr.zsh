# Shells opened by herdr-new-thread carry HERDR_THREAD=1. Once the claude
# session that herdr started in this shell exits, exit the shell too so the
# pane (and with it the single-pane tab) closes instead of dropping to a prompt.
if [[ -n "$HERDR_THREAD" ]]; then
  _herdr_thread_preexec() { [[ "$1" == claude* ]] && _herdr_thread_agent_ran=1 }
  _herdr_thread_precmd()  { [[ -n "$_herdr_thread_agent_ran" ]] && exit }
  autoload -Uz add-zsh-hook
  add-zsh-hook preexec _herdr_thread_preexec
  add-zsh-hook precmd  _herdr_thread_precmd
fi

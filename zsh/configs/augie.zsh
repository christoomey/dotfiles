# On-demand loader for augie session-command completion (bin/open · bin/close ·
# bin/down). This file only DEFINES the command below — it registers nothing and
# shadows nothing until you actually run `augie-complete` in a shell. Run it once
# per shell where you want the completion (e.g. when you start working in the
# fullstack repo).
augie-complete() {
  source /Users/christoomey/code/work/august/fullstack/bin/sessions.sh \
    && print "augie completion loaded for this shell"
}

# Completion for `af` (the augie dispatcher). First word: the subcommands.
# For open/close/down, reuse the subcommand's own `--candidates` (routed through
# af) so the feature list stays in lockstep with the tooling — open completes
# CLOSED features, close/down complete OPEN ones.
_af() {
  if (( CURRENT == 2 )); then
    compadd -- up review open close down ls status ui
    return
  fi
  case ${words[2]} in
    open|close|down)
      local -a cands chosen
      cands=(${(f)"$(af ${words[2]} --candidates 2>/dev/null)"})
      chosen=(${words[3,-1]})
      compadd -- ${cands:|chosen}
      ;;
  esac
}
# completion.zsh (which runs compinit) sorts after this file, so compdef isn't
# defined yet at load time. Register once on the first prompt, then unhook.
autoload -Uz add-zsh-hook
_af_register_completion() {
  compdef _af af 2>/dev/null
  add-zsh-hook -d precmd _af_register_completion
}
add-zsh-hook precmd _af_register_completion

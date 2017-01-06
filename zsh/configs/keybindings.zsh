# handy keybindings
bindkey "^a" beginning-of-line
bindkey "^e" end-of-line
bindkey "^f" forward-char
bindkey "^b" backward-char
bindkey "^k" kill-line
bindkey "^d" delete-char
bindkey "^p" history-search-backward
bindkey "^n" history-search-forward
bindkey "^y" accept-and-hold
bindkey "^w" backward-kill-word
bindkey "^u" backward-kill-line

# Open current command in Vim
autoload -z edit-command-line
zle -N edit-command-line
bindkey "^x^e" edit-command-line

# Copy the most recent command to the clipboard
function _pbcopy_last_command(){
  history | tail -1 | sed 's/ *[0-9]* *//' | pbcopy
}
zle -N pbcopy-last-command _pbcopy_last_command
bindkey '^x^y' pbcopy-last-command

# Fuzzy match against history, edit selected value
_uniqe_without_sort() { awk '!x[$0]++' }
_fuzzy_history() {
  zle -U "$(
    history | \
    tail -2000 | \
    sed 's/ *[0-9]* *//' | \
    _uniqe_without_sort | \
    fzf-tmux --tac --reverse --no-sort
  )"
}
zle -N fuzzy-history _fuzzy_history
bindkey '^r' fuzzy-history

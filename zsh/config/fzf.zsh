# Fuzzy match against history, edit selected value
fhist() {
  print -z $(([ -n "$ZSH_NAME" ] && fc -l 1 || history) | \
    tail -2000 | \
    fzf --tac --reverse --no-sort | \
    sed 's/ *[0-9]* *//')
}
bindkey -s '^r' 'fhist\n'

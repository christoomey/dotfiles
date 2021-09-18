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
  history | tail -1 | sed 's/ *[0-9]* *//' | pbcopy && \
    tmux display-message "Previous command coppied to clipboard"
}
zle -N pbcopy-last-command _pbcopy_last_command
bindkey '^x^y' pbcopy-last-command

# Git branches
_fuzzy_git_branches() {
  zle -U "$(
    git branch --color=always --sort=-committerdate | \
    grep -v '^* ' | \
    grep -v '^\s\+master' | \
    grep -v '^\s\+develop' | \
    grep -v '^\s\+development' | \
    fzf --reverse --ansi --select-1 --multi | \
    xargs echo | \
    sed -E 's/^[ \t]*//'
  )"
}
zle -N fuzzy-git-branches _fuzzy_git_branches
bindkey '^g^b' fuzzy-git-branches

# cat package.json| jq '.scripts | keys[]' | tr -d '"' |
#   fzf-tmux --reverse --ansi --preview 'cat package.json | jq ".scripts" | jq .["{}"]'
_yarn_scripts() {
  zle -U "$(
    cat package.json| jq '.scripts | keys[]' | tr -d '"' |
      fzf --reverse --ansi
  )"
}
zle -N yarn-scripts _yarn_scripts
bindkey '^g^y' yarn-scripts

# Git files
_fuzzy_git_status_files() {
  zle -U "$(
    git -c color.status=always status --short | \
    fzf --ansi --reverse --no-sort --multi | \
    cut -d ' ' -f 3
  )"
}
zle -N fuzzy-git-status-files _fuzzy_git_status_files
bindkey '^g^f' fuzzy-git-status-files

# Git files
_fuzzy_git_shalector() {
  commit=$(
    git log --color=always --oneline --decorate -35 | \
    fzf --ansi --reverse --no-sort
  )
  zle -U "$(echo $commit | cut -d ' ' -f 1)"
  zle -M "$commit"
}
zle -N fuzzy-git-shalector _fuzzy_git_shalector
bindkey '^g^g' fuzzy-git-shalector

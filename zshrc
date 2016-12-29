source ~/.zplug/init.zsh

zplug 'mafredri/zsh-async'
zplug 'sindresorhus/pure'
zplug 'zsh-users/zsh-syntax-highlighting', nice:10
zplug 'zsh-users/zsh-completions', nice:10

zplug load

# handy keybindings
bindkey "^a" beginning-of-line
bindkey "^e" end-of-line
bindkey "^f" forward-char
bindkey "^b" backward-char 
bindkey "^k" kill-line
bindkey "^d" delete-char
bindkey "^r" history-incremental-search-backward
bindkey "^p" history-search-backward
bindkey "^n" history-search-forward
bindkey "^y" accept-and-hold
bindkey "^q" push-line-or-edit

autoload -z edit-command-line
zle -N edit-command-line
bindkey "^x^e" edit-command-line

export EDITOR="vim"

for zsh_source in $HOME/.zsh/configs/*.zsh; do
  source $zsh_source
done

# makes color constants available
autoload -U colors
colors

# enable colored output from ls, etc. on FreeBSD-based systems
export CLICOLOR=1

function pbcopy-buffer(){
  print -rn $BUFFER | pbcopy
  zle -M "  pbcopy-ed: ${BUFFER}"
}
zle -N pbcopy-buffer
bindkey '^x^y' pbcopy-buffer

sz() { source ~/.zshrc }

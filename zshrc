# source "${HOME}/.zgen/zgen.zsh"
#
# zgen load mafredri/zsh-async
# zgen load sindresorhus/pure
# zgen load zsh-users/zsh-syntax-highlighting
# zgen load zsh-users/zsh-completions

source ~/.zplug/init.zsh

zplug 'mafredri/zsh-async'
zplug 'sindresorhus/pure'
zplug 'zsh-users/zsh-syntax-highlighting', nice:10
zplug 'zsh-users/zsh-completions', nice:10

zplug load

# handy keybindings
bindkey "^A" beginning-of-line
bindkey "^E" end-of-line
bindkey "^F" forward-char
bindkey "^B" backward-char 
bindkey "^K" kill-line
bindkey "^D" delete-char
bindkey "^R" history-incremental-search-backward
bindkey "^P" history-search-backward
bindkey "^Y" accept-and-hold
bindkey "^N" insert-last-word
bindkey "^Q" push-line-or-edit

for zsh_source in $HOME/.zsh/configs/*.zsh; do
  source $zsh_source
done

# makes color constants available
autoload -U colors
colors

# enable colored output from ls, etc. on FreeBSD-based systems
export CLICOLOR=1

# compdef pdn=heroku
function pdn {
  if [[ $# -gt 0 ]]; then
    production "$@"
  else
    production console
  fi
}
# compdef sgn=heroku
function sgn {
  if [[ $# -gt 0 ]]; then
    staging "$@"
  else
    staging console
  fi
}
function dev {
  if [[ $# -gt 0 ]]; then
    bundle exec rails "$@"
  else
    bundle exec rails console
  fi
}

copy-production-to() {
  if [ "$1" != "staging" ] && [ "$1" != "development" ]; then
    echo >&2 "Usage: copy-production-to <staging|development>"
    return 1
  else
    production backup && "$1" restore production
  fi
}

_copy-production-to() { reply=(development staging) }
compctl -K _copy-production-to copy-production-to

_bin-deploy() { reply=(production staging) }
compctl -K _bin-deploy bin/deploy


export BIND_OPTION="emacs"

function pbcopy-buffer(){
  print -rn $BUFFER | pbcopy
  zle -M "  pbcopy-ed: ${BUFFER}"
}
zle -N pbcopy-buffer
bindkey '^x^y' pbcopy-buffer


alias nombom='npm cache clear && bower cache clean && rm -rf node_modules bower_components && npm install && bower install'

sz() { source ~/.zshrc }

export PATH="$HOME/.yarn/bin:$PATH"

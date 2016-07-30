# source "${HOME}/.zgen/zgen.zsh"
#
# zgen load mafredri/zsh-async
# zgen load sindresorhus/pure
# zgen load zsh-users/zsh-syntax-highlighting
# zgen load zsh-users/zsh-completions

export ZPLUG_HOME=/usr/local/opt/zplug
source $ZPLUG_HOME/init.zsh

zplug 'mafredri/zsh-async'
zplug 'sindresorhus/pure'
zplug 'zsh-users/zsh-syntax-highlighting', nice:10
zplug 'zsh-users/zsh-completions', nice:10

# Install plugins if there are plugins that have not been installed
if ! zplug check --verbose; then
    printf "Install? [y/N]: "
    if read -q; then
        echo; zplug install
    fi
fi

zplug load

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


sz() { source ~/.zshrc }
function rbinstall() {
  command ruby-install ruby $(cat .ruby-version) > /dev/null
}

_chruby() { compadd $(chruby | tr -d '* ') }
compdef _chruby chruby

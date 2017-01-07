function rbinstall() {
  command ruby-install ruby $(cat .ruby-version) > /dev/null
  gem install bundler pry
}

if [[ -e /usr/local/share/chruby ]]; then
  source '/usr/local/share/chruby/chruby.sh'
  source '/usr/local/share/chruby/auto.sh'
  if [[ -f ~/.ruby-version ]]; then
    chruby $(cat ~/.ruby-version)
  fi
fi

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
    if [[ $1 == 'migrate' ]]; then
      bin/rake db:migrate db:test:prepare
    else
      bin/rails "$@"
    fi
  else
    bin/rails console
  fi
}
compdef production=heroku
compdef pdn=heroku

compdef staging=heroku
compdef sgn=heroku

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

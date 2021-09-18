unsetopt auto_cd # with cdpath enabled, auto_cd gives too many false positives
cdpath=(
  $HOME/code \
  $HOME/code/work/current \
  $HOME/code/work/current-two \
  $HOME/code/work \
  $HOME/code/vim \
  $HOME/code/alfred \
  $HOME
)

itree() {
  if [ -f .gitignore ]; then
    tree -I "$(cat .gitignore | paste -s -d'|' -)" -C | less -R
  else
    tree -I node_modules -C
  fi
}

function cdup() {
  cd "$(git rev-parse --show-toplevel)"
}

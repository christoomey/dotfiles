# Tab completion for the worktree lifecycle scripts. Replaces the old
# `workmux completions zsh` (zshrc) now that bin/up + bin/rm-worktree own the
# worktree flow. bin/down takes no arguments, so it needs no completion.

# Directory names of the repo's checked-out worktrees, excluding the primary
# checkout — i.e. the open worktrees bin/rm-worktree can trash.
_open_worktrees() {
  local -a names
  names=(${(f)"$(git worktree list --porcelain 2>/dev/null \
    | awk '/^worktree /{ print $2 }' \
    | tail -n +2 \
    | sed 's#.*/##')"})
  compadd -a names
}
compdef _open_worktrees bin/rm-worktree

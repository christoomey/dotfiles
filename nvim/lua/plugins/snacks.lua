-- Startup dashboard.
--
-- Shown on a bare `nvim` launch. The "Recent in Worktree" list is scoped to
-- the cwd, which in the worktree workflow means it is effectively scoped to
-- the current branch (each branch is checked out in its own directory). Key
-- actions drive the existing Telescope setup rather than snacks' own picker.

return {
  'folke/snacks.nvim',
  priority = 1000,
  lazy = false,
  opts = {
    dashboard = {
      width = 90,
      preset = {
        keys = {
          { icon = ' ', key = 'f', desc = 'Find File', action = ':Telescope find_files' },
          { icon = ' ', key = 'r', desc = 'Recent (cwd)', action = ':Telescope oldfiles cwd_only=true' },
          { icon = ' ', key = 'g', desc = 'Grep', action = ':Telescope live_grep' },
          { icon = ' ', key = 'b', desc = 'Branch Files', action = ":lua require('git_branch').files()" },
          { icon = ' ', key = 'n', desc = 'New File', action = ':ene | startinsert' },
          {
            icon = ' ',
            key = 'c',
            desc = 'Config',
            action = ":lua require('telescope.builtin').find_files({ cwd = vim.fn.stdpath('config') })",
          },
          { icon = '󰒲 ', key = 'L', desc = 'Lazy', action = ':Lazy' },
          { icon = ' ', key = 'q', desc = 'Quit', action = ':qa' },
        },
      },
      sections = {
        { section = 'header' },
        { section = 'keys', gap = 1, padding = 1 },
        { icon = ' ', title = 'Recent in Worktree', section = 'recent_files', cwd = true, limit = 8, padding = 1 },
        {
          icon = ' ',
          title = 'Branch Status',
          section = 'terminal',
          -- In a normal worktree the cwd is itself a git repo, so show its
          -- status. In the multi-repo investigation layout the cwd is NOT a
          -- repo (the repos are nested one level down) -- show each nested
          -- repo's status instead of erroring with `not a git repository`.
          cmd = [[
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git -c color.status=always status --short --branch
else
  for d in */(N); do
    git -C "$d" rev-parse --is-inside-work-tree >/dev/null 2>&1 || continue
    printf '\033[1;34m%s\033[0m\n' "${d%/}"
    git -C "$d" -c color.status=always status --short --branch
    echo
  done
fi
]],
          height = 8,
          ttl = 0,
          padding = 1,
        },
        { section = 'startup' },
      },
    },
  },
}

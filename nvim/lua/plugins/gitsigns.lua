-- Branch-scoped git annotations.
--
-- Rather than gitsigns' default (diff against the index, i.e. uncommitted
-- changes only), we diff against the merge-base with the default branch. That
-- reframes the gutter as "everything this branch changed since it forked" —
-- committed work included — which is what's useful when jumping around a
-- feature branch. The base is applied once the first buffer's diff has settled
-- and refreshed after any repo change. On open, the cursor lands on the first
-- such change.

return {
  'lewis6991/gitsigns.nvim',
  event = { 'BufReadPre', 'BufNewFile' },
  config = function()
    local gs = require 'gitsigns'

    -- Resolve the branch's fork point: merge-base(HEAD, <default branch>).
    -- Prefers the remote's default branch, falling back to master then main.
    local function default_base_ref()
      local origin = vim.fn.systemlist({ 'git', 'rev-parse', '--abbrev-ref', 'origin/HEAD' })[1]
      if vim.v.shell_error == 0 and origin and origin ~= '' and not origin:find 'fatal' then
        return origin
      end
      for _, ref in ipairs { 'master', 'main' } do
        vim.fn.system { 'git', 'rev-parse', '--verify', '--quiet', ref }
        if vim.v.shell_error == 0 then
          return ref
        end
      end
      return nil
    end

    local function branch_merge_base()
      local base_ref = default_base_ref()
      if not base_ref then
        return nil
      end
      local mb = vim.fn.systemlist({ 'git', 'merge-base', 'HEAD', base_ref })[1]
      if vim.v.shell_error == 0 and mb and mb ~= '' then
        return mb
      end
      return nil
    end

    -- Auto jump-to-first-change state. `jumped` ensures we only reposition once
    -- per buffer, not on every edit.
    local auto_jump = true
    local jumped = {}

    local function maybe_jump(buf)
      if not auto_jump or jumped[buf] then
        return
      end
      if not vim.api.nvim_buf_is_valid(buf) or vim.bo[buf].buftype ~= '' then
        return
      end
      local win = vim.fn.bufwinid(buf)
      if win == -1 then
        return
      end
      -- Only reposition an untouched open; if anything already moved the cursor
      -- (`nvim file +42`, quickfix, a restore plugin), leave it be.
      if vim.api.nvim_win_get_cursor(win)[1] ~= 1 then
        jumped[buf] = true
        return
      end
      local hunks = gs.get_hunks(buf)
      if not hunks or #hunks == 0 then
        return
      end
      jumped[buf] = true
      vim.api.nvim_win_call(win, function()
        gs.nav_hunk('first', { foldopen = true })
        vim.cmd 'normal! zz'
      end)
    end

    -- Re-point all buffers at the branch fork point. gitsigns only reliably
    -- adopts a non-index base via change_base() once a buffer's git object is
    -- ready, so this is driven from the first GitSignsUpdate (see below) rather
    -- than the setup `base` option. global=true covers all current + future
    -- buffers; the callback form ensures the re-diff has completed before we
    -- reposition the cursor.
    local base_applied = false
    local function refresh_branch_base()
      local mb = branch_merge_base()
      if not mb then
        return
      end
      gs.change_base(mb, true, function()
        vim.g.gitsigns_branch_base = mb
        vim.schedule(function()
          maybe_jump(vim.api.nvim_get_current_buf())
        end)
      end)
    end

    gs.setup {
      on_attach = function(bufnr)
        local function map(mode, lhs, rhs, opts)
          opts = opts or {}
          opts.buffer = bufnr
          vim.keymap.set(mode, lhs, rhs, opts)
        end

        -- ]c / [c navigate branch changes, but defer to native diff-mode
        -- behavior when this window is part of a diff.
        map('n', ']c', function()
          if vim.wo.diff then
            return ']c'
          end
          vim.schedule(function()
            gs.nav_hunk 'next'
          end)
          return '<Ignore>'
        end, { expr = true, desc = 'Next branch change' })

        map('n', '[c', function()
          if vim.wo.diff then
            return '[c'
          end
          vim.schedule(function()
            gs.nav_hunk 'prev'
          end)
          return '<Ignore>'
        end, { expr = true, desc = 'Prev branch change' })

        map('n', ']C', function()
          gs.nav_hunk 'last'
        end, { desc = 'Last branch change' })
        map('n', '[C', function()
          gs.nav_hunk 'first'
        end, { desc = 'First branch change' })

        map('n', '<leader>hp', gs.preview_hunk, { desc = 'Preview hunk' })
        map('n', '<leader>hb', function()
          gs.blame_line { full = true }
        end, { desc = 'Blame line' })
      end,
    }

    local grp = vim.api.nvim_create_augroup('gitsigns-branch-base', { clear = true })

    -- The first GitSignsUpdate signals a buffer attached and its initial diff
    -- settled — the point at which the git object is ready and change_base
    -- reliably re-diffs against the branch base. The small defer exits the
    -- autocmd context first. Thereafter, reposition buffers onto their first
    -- branch change as signs land.
    vim.api.nvim_create_autocmd('User', {
      pattern = 'GitSignsUpdate',
      group = grp,
      callback = function(args)
        if not base_applied then
          base_applied = true
          vim.defer_fn(refresh_branch_base, 50)
        end
        local buf = args.data and args.data.buffer
        if buf then
          vim.schedule(function()
            maybe_jump(buf)
          end)
        end
      end,
    })

    -- Keep the fork point accurate after commits, checkouts, etc.
    vim.api.nvim_create_autocmd('User', {
      pattern = 'GitSignsChanged',
      group = grp,
      callback = function()
        vim.schedule(refresh_branch_base)
      end,
    })

    vim.api.nvim_create_autocmd('BufDelete', {
      group = grp,
      callback = function(args)
        jumped[args.buf] = nil
      end,
    })

    vim.api.nvim_create_user_command('GitsignsBranchBase', refresh_branch_base, {
      desc = 'Recompute branch diff base (merge-base with default branch)',
    })
    vim.api.nvim_create_user_command('GitsignsBranchJumpToggle', function()
      auto_jump = not auto_jump
      vim.notify('Branch jump-on-open: ' .. (auto_jump and 'on' or 'off'))
    end, { desc = 'Toggle auto jump-to-first-branch-change on open' })
  end,
}

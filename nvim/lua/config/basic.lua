vim.g.mapleader = ' '
vim.g.maplocalleader = ' '

-- Set to true if you have a Nerd Font installed and selected in the terminal
vim.g.have_nerd_font = true

-- Make line numbers default
vim.opt.number = true
-- You can also add relative line numbers, to help with jumping.
--  Experiment for yourself to see if you like it!
vim.opt.relativenumber = true

-- Enable mouse mode, can be useful for resizing splits for example!
vim.opt.mouse = 'a'

-- Don't show the mode, since it's already in the status line
vim.opt.showmode = false

-- Enable break indent
vim.opt.breakindent = true

-- Save undo history
vim.opt.undofile = true

-- Automatically reload files changed outside of Neovim
vim.opt.autoread = true

-- Case-insensitive searching UNLESS \C or one or more capital letters in the search term
vim.opt.ignorecase = true
vim.opt.smartcase = true

-- Keep signcolumn on by default
vim.opt.signcolumn = 'yes'

-- Decrease update time
vim.opt.updatetime = 250

-- Decrease mapped sequence wait time
-- Displays which-key popup sooner
vim.opt.timeoutlen = 300

-- Configure how new splits should be opened
vim.opt.splitright = true
vim.opt.splitbelow = true

-- Sets how neovim will display certain whitespace characters in the editor.
--  See `:help 'list'`
--  and `:help 'listchars'`
vim.opt.list = true
vim.opt.listchars = { tab = '» ', trail = '·', nbsp = '␣' }

-- Preview substitutions live, as you type!
vim.opt.inccommand = 'split'

-- Show which line your cursor is on
vim.opt.cursorline = true

-- Minimal number of screen lines to keep above and below the cursor.
vim.opt.scrolloff = 10

-- Set highlight on search, but clear on pressing <Esc> in normal mode
vim.opt.hlsearch = true
vim.keymap.set('n', '<Esc>', '<cmd>nohlsearch<CR>')

-- Diagnostic keymaps
vim.keymap.set('n', '[d', vim.diagnostic.goto_prev, { desc = 'Go to previous [D]iagnostic message' })
vim.keymap.set('n', ']d', vim.diagnostic.goto_next, { desc = 'Go to next [D]iagnostic message' })
vim.keymap.set('n', '<leader>e', vim.diagnostic.open_float, { desc = 'Show diagnostic [E]rror messages' })
vim.keymap.set('n', '<leader>q', vim.diagnostic.setloclist, { desc = 'Open diagnostic [Q]uickfix list' })

vim.keymap.set('n', 'j', 'gj')
vim.keymap.set('n', 'k', 'gk')

-- Highlight when yanking (copying) text
--  Try it with `yap` in normal mode
--  See `:help vim.highlight.on_yank()`
vim.api.nvim_create_autocmd('TextYankPost', {
  desc = 'Highlight when yanking (copying) text',
  group = vim.api.nvim_create_augroup('kickstart-highlight-yank', { clear = true }),
  callback = function()
    vim.highlight.on_yank()
  end,
})

-- Check for external file changes when focusing back on Neovim or navigating windows/tabs
vim.api.nvim_create_autocmd({ 'FocusGained', 'TermClose', 'TermLeave', 'BufEnter', 'WinEnter' }, {
  desc = 'Check for external file changes',
  group = vim.api.nvim_create_augroup('checktime', { clear = true }),
  callback = function()
    if vim.o.buftype ~= 'nofile' then
      vim.cmd 'checktime'
    end
  end,
})

-- Session management: automatically save and restore window/tab layout per directory/branch
vim.opt.sessionoptions = 'buffers,curdir,tabpages,winsize,winpos'

-- Function to get session file path based on cwd and git branch
local function get_session_file()
  local cwd = vim.fn.getcwd()
  -- Replace path separators with underscores to create a valid filename
  local cwd_encoded = cwd:gsub('/', '_'):gsub('\\', '_')

  -- Try to get git branch name
  local branch = vim.fn.system('git -C ' .. vim.fn.shellescape(cwd) .. ' rev-parse --abbrev-ref HEAD 2>/dev/null'):gsub('\n', '')

  -- If in a git repo and got a valid branch, include it in the filename
  local session_name = cwd_encoded
  if vim.v.shell_error == 0 and branch ~= '' then
    session_name = cwd_encoded .. '_' .. branch:gsub('/', '_')
  end

  -- Ensure sessions directory exists
  local sessions_dir = vim.fn.stdpath('data') .. '/sessions'
  vim.fn.mkdir(sessions_dir, 'p')

  return sessions_dir .. '/' .. session_name .. '.vim'
end

-- Save session on exit
vim.api.nvim_create_autocmd('VimLeavePre', {
  desc = 'Save session on exit',
  group = vim.api.nvim_create_augroup('auto-session', { clear = true }),
  callback = function()
    -- Only save if we're in a real directory (not empty buffer or special buffer)
    if vim.fn.argc() > 0 or vim.fn.bufname() ~= '' then
      local session_file = get_session_file()
      vim.cmd('mksession! ' .. vim.fn.fnameescape(session_file))
    end
  end,
})

-- Restore session on startup (only if no files were specified)
if vim.fn.argc() == 0 then
  vim.api.nvim_create_autocmd('VimEnter', {
    desc = 'Restore session on startup',
    group = vim.api.nvim_create_augroup('auto-session-restore', { clear = true }),
    nested = true,
    callback = function()
      local session_file = get_session_file()
      if vim.fn.filereadable(session_file) == 1 then
        vim.cmd('source ' .. vim.fn.fnameescape(session_file))
      end
    end,
  })
end

vim.keymap.set('n', '0', '^')

-- Function to close the hover window and clear search highlighting
local function enhanced_ctrl_c()
  -- Close floating windows (e.g., hover)
  for _, win in ipairs(vim.api.nvim_tabpage_list_wins(0)) do
    local config = vim.api.nvim_win_get_config(win)
    if config.relative ~= '' then
      vim.api.nvim_win_close(win, true)
    end
  end

  -- Clear search highlighting
  vim.cmd 'nohlsearch'

  -- Perform the default <C-c> behavior
  -- This is typically a no-op in normal mode, but you can map it to <Esc> if needed
  vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes('<Esc>', true, false, true), 'n', true)
end

-- Keymap to execute the enhanced <C-c> functionality
vim.keymap.set('n', '<C-c>', enhanced_ctrl_c, { noremap = true, silent = true })

vim.keymap.set('n', '<leader>x', '<cmd>.lua<CR>', { desc = 'Execute the current line' })
vim.keymap.set('n', '<leader><leader>x', '<cmd>source %<CR>', { desc = 'Execute the current file' })

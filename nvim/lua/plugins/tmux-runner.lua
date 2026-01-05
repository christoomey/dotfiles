local function attachAndHighlightFn(pane)
  return function()
    vim.cmd('VtrAttachToPane ' .. pane)
    os.execute('tmux clock-mode -t ' .. pane .. ' && sleep 0.1 && tmux send-keys -t ' .. pane .. ' q')
  end
end

return {
  'christoomey/vim-tmux-runner',
  dir = '~/code/vim/tmux-runner/',
  lazy = false,
  keys = {
    { '<leader>v0', attachAndHighlightFn(0) },
    { '<leader>v1', attachAndHighlightFn(1) },
    { '<leader>v2', attachAndHighlightFn(2) },
    { '<leader>v3', attachAndHighlightFn(3) },
    { '<leader>fr', '<cmd>VtrFocusRunner<cr>' },
    { '<leader>\\s', '<cmd>VtrSendCommandToRunner<cr>' },
    { '<leader>\\', '<Plug>VtrSend' },
    { '<leader>\\', '<Plug>VtrSend', mode = 'x' },
    { '<leader>\\\\', '<Plug>VtrSendLine' },
    { '<leader>sc', '<cmd>VtrSendKeysRaw ^c<cr>' },
    { '<leader>sd', '<cmd>VtrSendKeysRaw ^d<cr>' },
    { '<leader>sl', '<cmd>VtrSendKeysRaw ^l<cr>' },
  },
}

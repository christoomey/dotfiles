return {
  'gabebw/vim-spec-runner',
  init = function()
    vim.g.spec_runner_dispatcher = 'call VtrSendCommand("bin/{command}")'
  end,
  keys = {
    { '<leader>t', '<Plug>RunFocusedSpec' },
    { '<leader>a', '<Plug>RunCurrentSpecFile' },
  },
}

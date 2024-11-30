return {
  'metalelf0/jellybeans-nvim',
  dependencies = { 'rktjmp/lush.nvim' },
  priority = 1000,
  init = function()
    vim.cmd.colorscheme 'jellybeans-nvim'
  end,
}

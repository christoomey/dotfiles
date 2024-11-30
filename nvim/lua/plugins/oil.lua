return {
  'stevearc/oil.nvim',
  config = function()
    require('oil').setup()
    vim.keymap.set('n', '-', '<cmd>Oil<cr>', { desc = 'Open parent directory' })
  end,
  dependencies = { { 'echasnovski/mini.icons', opts = {} } },
}

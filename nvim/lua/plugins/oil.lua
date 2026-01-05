return {
  'stevearc/oil.nvim',
  config = function()
    require('oil').setup {
      delete_to_trash = true,
      skip_confirm_for_simple_edits = true,
    }
    vim.keymap.set('n', '-', '<cmd>Oil<cr>', { desc = 'Open parent directory' })
  end,
  dependencies = { { 'echasnovski/mini.icons', opts = {} } },
}

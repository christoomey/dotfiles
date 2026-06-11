return {
  'vim-scripts/ReplaceWithRegister',
  lazy = false,
  dependencies = { 'vim-scripts/ingo-library' },
  keys = { { '<leader>gr', '"*gr', remap = true } },
  init = function()
    -- Neovim 0.11+ ships gri/gra/grn/grt as default LSP maps,
    -- which shadow the gr operator's text-object motions.
    vim.api.nvim_create_autocmd('VimEnter', {
      callback = function()
        for _, lhs in ipairs { 'gri', 'gra', 'grn', 'grt' } do
          pcall(vim.keymap.del, 'n', lhs)
        end
        pcall(vim.keymap.del, 'x', 'gra')
      end,
    })
  end,
}

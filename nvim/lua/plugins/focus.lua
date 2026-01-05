return {
  'nvim-focus/focus.nvim',
  version = '*',
  config = function()
    require('focus').setup {
      ui = {
        hybridnumber = true,
        cursorline = true,
        cursorcolumn = true,
        absolutenumber_unfocussed = true,
      },
    }
  end,
}

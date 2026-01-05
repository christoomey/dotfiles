return {
  'tpope/vim-fugitive',
  config = function()
    vim.g.fugitive_legacy_commands = 0
    vim.api.nvim_create_user_command('Gbrowse', function(opts)
      if opts.range > 0 then
        vim.cmd(string.format('GBrowse %d,%d', opts.line1, opts.line2))
      else
        vim.cmd 'GBrowse'
      end
    end, { range = true })
  end,
}

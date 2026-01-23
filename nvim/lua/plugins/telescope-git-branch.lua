vim.keymap.set({ 'n', 'v' }, '<leader>gb', function()
  require('git_branch').files()
end)

vim.keymap.set({ 'n', 'v' }, '<leader>GB', function()
  require('telescope.builtin').git_status()
end)

return {
  'mrloop/telescope-git-branch.nvim',
}

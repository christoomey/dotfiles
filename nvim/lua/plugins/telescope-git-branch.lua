vim.keymap.set({ 'n', 'v' }, '<leader>gp', function()
  require('git_branch').files()
end)

vim.keymap.set({ 'n', 'v' }, '<leader>GP', function()
  require('telescope.builtin').git_status()
end)

return {
  'mrloop/telescope-git-branch.nvim',
}

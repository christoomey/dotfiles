local paths = vim.split(vim.fn.glob '~/.config/nvim/rcfiles/*', '\n')

for _, file in pairs(paths) do
  vim.cmd('source ' .. file)
end

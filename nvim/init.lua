require 'config.basic'

require 'config.lazy'

local paths = vim.split(vim.fn.glob '~/.config/nvim/rcfiles/*', '\n')

for _, file in pairs(paths) do
  vim.cmd('source ' .. file)
end

-- vim: ts=2 sts=2 sw=2 et

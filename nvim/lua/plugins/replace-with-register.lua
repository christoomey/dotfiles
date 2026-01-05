return {
  {
    'vim-scripts/ReplaceWithRegister',
    lazy = false,
    dependencies = { 'vim-scripts/ingo-library' },
    keys = { { '<leader>gr', '"*gr', remap = true } },
  },
  { 'vim-scripts/ReplaceWithSameIndentRegister', dependencies = { 'vim-scripts/ingo-library' } },
}

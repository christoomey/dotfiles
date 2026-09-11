return {
  'scalameta/nvim-metals',
  dependencies = {
    'nvim-lua/plenary.nvim',
    'hrsh7th/cmp-nvim-lsp',
    -- Displays LSP progress (import-build, compile) so long waits are visible
    -- instead of looking like a hang.
    { 'j-hui/fidget.nvim', opts = {} },
  },
  ft = { 'scala', 'sbt' },
  opts = function()
    local metals_config = require('metals').bare_config()

    metals_config.settings = {
      serverVersion = 'latest.release',
      -- Give the server enough heap for the loquat build; default GC-crawls
      serverProperties = { '-Xmx2G', '-Xss4m' },
      -- Show implicit arguments and conversions inline
      inlayHints = {
        implicitArguments = { enable = true },
        implicitConversions = { enable = true },
      },
      -- Filter out common unwanted packages from completions
      excludedPackages = { 'akka.actor.typed.javadsl', 'com.github.swagger.akka.javadsl' },
    }

    -- Route metals status through LSP progress/messages (picked up by fidget)
    -- rather than requiring statusline integration.
    metals_config.init_options.statusBarProvider = 'off'

    -- Pass cmp-nvim-lsp capabilities for richer completions (snippets, etc.)
    metals_config.capabilities = require('cmp_nvim_lsp').default_capabilities()

    metals_config.on_attach = function(_, bufnr)
      vim.lsp.inlay_hint.enable(true, { bufnr = bufnr })
    end

    return metals_config
  end,
  config = function(self, metals_config)
    local nvim_metals_group = vim.api.nvim_create_augroup('nvim-metals', { clear = true })
    vim.api.nvim_create_autocmd('FileType', {
      pattern = self.ft,
      callback = function()
        require('metals').initialize_or_attach(metals_config)
      end,
      group = nvim_metals_group,
    })
  end,
}

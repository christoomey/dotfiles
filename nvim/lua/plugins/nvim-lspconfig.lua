return { -- LSP Configuration & Plugins
  'neovim/nvim-lspconfig',
  dependencies = {
    { 'mason-org/mason.nvim', opts = {} },
    {
      'mason-org/mason-lspconfig.nvim',
      opts = {
        ensure_installed = { 'lua_ls' },
        -- automatic_enable = true (the default) calls vim.lsp.enable()
        -- for all Mason-installed servers automatically
      },
    },
    'WhoIsSethDaniel/mason-tool-installer.nvim',

    -- Useful status updates for LSP.
    { 'j-hui/fidget.nvim', opts = {} },

    -- `lazydev` configures Lua LSP for your Neovim config, runtime and plugins
    {
      'folke/lazydev.nvim',
      ft = 'lua',
      opts = {
        library = {
          { path = '${3rd}/luv/library', words = { 'vim%.uv' } },
        },
      },
    },
  },
  config = function()
    -- Rounded borders on all floating windows (hover, diagnostics, etc.)
    vim.o.winborder = 'rounded'

    vim.diagnostic.config {
      virtual_text = false,
      update_in_insert = false,
    }

    vim.api.nvim_create_autocmd('LspAttach', {
      group = vim.api.nvim_create_augroup('kickstart-lsp-attach', { clear = true }),
      callback = function(event)
        local map = function(keys, func, desc)
          vim.keymap.set('n', keys, func, { buffer = event.buf, desc = 'LSP: ' .. desc })
        end

        -- Native go-to-definition: jumps directly on single result (pushing the
        -- tagstack so <C-t> works), opens quickfix on multiple. Avoids
        -- telescope 0.1.x's deprecated jump_to_location / supports_method paths.
        map('gd', vim.lsp.buf.definition, '[G]oto [D]efinition')
        map('<C-]>', vim.lsp.buf.definition, '[G]oto [D]efinition')
        map('<leader>]', vim.lsp.buf.definition, '[G]oto [D]efinition')

        map('<leader>[', function()
          require('telescope.builtin').lsp_references { include_current_line = false, include_declaration = false, show_line = false }
        end, '[G]oto [R]eferences')

        map('gI', require('telescope.builtin').lsp_implementations, '[G]oto [I]mplementation')
        map('<leader>D', require('telescope.builtin').lsp_type_definitions, 'Type [D]efinition')
        map('<leader>ds', require('telescope.builtin').lsp_document_symbols, '[D]ocument [S]ymbols')
        map('<leader>ws', require('telescope.builtin').lsp_dynamic_workspace_symbols, '[W]orkspace [S]ymbols')

        map('<leader>rn', vim.lsp.buf.rename, '[R]e[n]ame')

        map('<leader>ca', vim.lsp.buf.code_action, '[C]ode [A]ction')
        map('<leader>do', vim.lsp.buf.code_action, '(do) Code Action')

        vim.keymap.set('n', ']r', function()
          vim.diagnostic.jump { count = 1 }
        end, { desc = '[G]oto Next E[r]ror' })
        vim.keymap.set('n', '[r', function()
          vim.diagnostic.jump { count = -1 }
        end, { desc = '[G]oto Prev E[r]ror' })

        map('K', vim.lsp.buf.hover, 'Hover Documentation')
        map('<leader>k', vim.lsp.buf.hover, 'Hover Documentation')

        map('gD', vim.lsp.buf.declaration, '[G]oto [D]eclaration')

        local client = vim.lsp.get_client_by_id(event.data.client_id)
        if client and client.server_capabilities.documentHighlightProvider then
          local highlight_augroup = vim.api.nvim_create_augroup('kickstart-lsp-highlight', { clear = false })
          vim.api.nvim_create_autocmd({ 'CursorHold', 'CursorHoldI' }, {
            buffer = event.buf,
            group = highlight_augroup,
            callback = vim.lsp.buf.document_highlight,
          })

          vim.api.nvim_create_autocmd({ 'CursorMoved', 'CursorMovedI' }, {
            buffer = event.buf,
            group = highlight_augroup,
            callback = vim.lsp.buf.clear_references,
          })

          vim.api.nvim_create_autocmd('LspDetach', {
            group = vim.api.nvim_create_augroup('kickstart-lsp-detach', { clear = true }),
            callback = function(event2)
              vim.lsp.buf.clear_references()
              vim.api.nvim_clear_autocmds { group = 'kickstart-lsp-highlight', buffer = event2.buf }
            end,
          })
        end

        if client and client.server_capabilities.inlayHintProvider and vim.lsp.inlay_hint then
          map('<leader>th', function()
            vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled())
          end, '[T]oggle Inlay [H]ints')
        end

        -- Workaround for Svelte language server not picking up TS/JS file changes
        -- See: https://github.com/sveltejs/language-tools/issues/2008
        if client and client.name == 'svelte' then
          vim.api.nvim_create_autocmd({ 'BufWritePost', 'BufReadPost', 'FileChangedShellPost' }, {
            pattern = { '*.js', '*.ts' },
            group = vim.api.nvim_create_augroup('svelte-on-ts-change', { clear = true }),
            callback = function(ctx)
              local file = ctx.file or vim.api.nvim_buf_get_name(ctx.buf)
              client.notify('$/onDidChangeTsOrJsFile', { uri = vim.uri_from_fname(file) })
            end,
          })
        end
      end,
    })

    -- Global LSP capabilities (shared by all servers)
    local capabilities = vim.lsp.protocol.make_client_capabilities()
    capabilities = vim.tbl_deep_extend('force', capabilities, require('cmp_nvim_lsp').default_capabilities())

    vim.lsp.config('*', {
      capabilities = capabilities,
    })

    -- Per-server configs
    vim.lsp.config('lua_ls', {
      settings = {
        Lua = {
          completion = {
            callSnippet = 'Replace',
          },
        },
      },
    })

    vim.lsp.config('ruby_lsp', {
      init_options = {
        formatter = 'none',
        linters = { 'none' },
      },
    })

    require('mason-tool-installer').setup {
      ensure_installed = {
        'stylua',
      },
    }
  end,
}

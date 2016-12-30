" ======================================================================
" ================= START thoughtbot dotfiles import ===============
" ======================================================================

set nocompatible

" Need to set the leader before defining any leader mappings
let mapleader = "\<Space>"

function! s:SourceConfigFilesIn(directory)
  let directory_splat = '~/.vim/' . a:directory . '/*'
  for config_file in split(glob(directory_splat), '\n')
    if filereadable(config_file)
      execute 'source' config_file
    endif
  endfor
endfunction

call s:SourceConfigFilesIn('rcplugins')
call s:SourceConfigFilesIn('rcfiles')

" Switch syntax highlighting on, when the terminal has colors
" Also switch on highlighting the last used search pattern.
if (&t_Co > 2 || has("gui_running")) && !exists("syntax_on")
  syntax on
endif

if filereadable(expand("~/.vimrc.bundles"))
  source ~/.vimrc.bundles
endif

" Load matchit.vim, but only if the user hasn't installed a newer version.
if !exists('g:loaded_matchit') && findfile('plugin/matchit.vim', &rtp) ==# ''
  runtime! macros/matchit.vim
endif

filetype plugin indent on

augroup vimrcEx
  autocmd!

  " When editing a file, always jump to the last known cursor position.
  " Don't do it for commit messages, when the position is invalid, or when
  " inside an event handler (happens when dropping a file on gvim).
  autocmd BufReadPost *
    \ if &ft != 'gitcommit' && line("'\"") > 0 && line("'\"") <= line("$") |
    \   exe "normal g`\"" |
    \ endif

  " Set syntax highlighting for specific file types
  autocmd BufRead,BufNewFile *.md set filetype=markdown
augroup END

" When the type of shell script is /bin/sh, assume a POSIX-compatible
" shell for syntax highlighting purposes.
let g:is_posix = 1

" Softtabs, 2 spaces
set tabstop=2
set shiftwidth=2
set shiftround
set expandtab

" Display extra whitespace
set list listchars=tab:»·,trail:·,nbsp:·

" Use one space, not two, after punctuation.
set nojoinspaces

" Use The Silver Searcher https://github.com/ggreer/the_silver_searcher
if executable('ag')
  " Use Ag over Grep
  set grepprg=ag\ --nogroup\ --nocolor

  " Use ag in CtrlP for listing files. Lightning fast and respects .gitignore
  let g:ctrlp_user_command = 'ag -Q -l --nocolor --hidden -g "" %s'

  " ag is fast enough that CtrlP doesn't need to cache
  let g:ctrlp_use_caching = 0

  if !exists(":Ag")
    command -nargs=+ -complete=file -bar Ag silent! grep! <args>|cwindow|redraw!
    nnoremap \ :Ag<SPACE>
  endif
endif

" Make it obvious where 80 characters is
set textwidth=80
set colorcolumn=+1

" Numbers
set number
set numberwidth=5

" Tab completion
" will insert tab at beginning of line,
" will use completion if not at beginning
set wildmode=list:longest,list:full
function! InsertTabWrapper()
    let col = col('.') - 1
    if !col || getline('.')[col - 1] !~ '\k'
        return "\<tab>"
    else
        return "\<c-p>"
    endif
endfunction
inoremap <Tab> <c-r>=InsertTabWrapper()<cr>
inoremap <S-Tab> <c-n>

" Switch between the last two files
nnoremap <leader><leader> <c-^>

" Treat <li> and <p> tags like the block tags they are
let g:html_indent_tags = 'li\|p'

set hidden

nmap j gj
nmap k gk

" Open new split panes to right and bottom, which feels more natural
set splitbelow
set splitright

" configure syntastic syntax checking to check on open as well as save
let g:syntastic_check_on_open=1
let g:syntastic_html_tidy_ignore_errors=[" proprietary attribute \"ng-"]
let g:syntastic_eruby_ruby_quiet_messages =
    \ {"regex": "possibly useless use of a variable in void context"}

" Set spellfile to location that is guaranteed to exist, can be symlinked to
" Dropbox or kept in Git and managed outside of thoughtbot/dotfiles using rcm.
set spellfile=$HOME/.vim-spell-en.utf-8.add

" Autocomplete with dictionary words when spell check is on
set complete+=kspell

" Always use vertical diffs
set diffopt+=vertical

" ======================================================================
" ================= END thoughtbot dotfiles import ===============
" ======================================================================




colorscheme jellybeans

" Use C-Space to Esc out of any mode
nnoremap <C-Space> <Esc>:noh<CR>
vnoremap <C-Space> <Esc>gV
onoremap <C-Space> <Esc>
cnoremap <C-Space> <C-c>
inoremap <C-Space> <Esc>
" Terminal sees <C-@> as <C-space> WTF, but ok
nnoremap <C-@> <Esc>:noh<CR>
vnoremap <C-@> <Esc>gV
onoremap <C-@> <Esc>
cnoremap <C-@> <C-c>
inoremap <C-@> <Esc>

nnoremap <leader>sop :source %<cr>

set relativenumber

nnoremap <leader>df :Files ~/code/dotfiles<cr>

nnoremap <leader>; :

" automatically rebalance windows on vim resize
autocmd VimResized * :wincmd =

" zoom a vim pane, <C-w>= to re-balance
nnoremap <leader>- :wincmd _<cr>:wincmd \|<cr>
nnoremap <leader>= :wincmd =<cr>



nnoremap <leader>dot :Files ~/dotfiles-local<cr>

nmap <leader>bi :PlugInstall<cr>



nnoremap Q @q

nnoremap 0 ^


set hlsearch
set incsearch
set ignorecase
set smartcase
nnoremap <leader>sub :%s///g<left><left>
vnoremap <leader>sub :s///g<left><left>





" START headers

function! s:SmartLevelThreeHeader()
  call s:DeleteExistingUnderline()
  call s:DeleteExistingLeadingHeaderMarks()
  s/^/### /
  silent! call repeat#set("\<Plug>SmartLevelThreeHeader")
endfunction

function! s:OnLastLineOfFile()
  return line('.') == line('$')
endfunction

function! s:DeleteExistingLeadingHeaderMarks()
  silent! s/^#\{1,6} //
endfunction

function! s:DeleteExistingUnderline()
  if !s:OnLastLineOfFile()
    let saved_cursor = getpos(".")
    +1g/\v^[-=]+$/d
    call setpos('.', saved_cursor)
  endif
endfunction

function! s:SmartUnderline(char)
  call s:DeleteExistingUnderline()
  call s:DeleteExistingLeadingHeaderMarks()
  let underline = repeat(a:char, len(getline('.')))
  call append(line('.'), underline)
  if a:char == '='
    silent! call repeat#set("\<Plug>UnderlineH1", v:count)
  else
    silent! call repeat#set("\<Plug>UnderlineH2", v:count)
  end
endfunction

nnoremap <silent> <Plug>UnderlineH1 :call <sid>SmartUnderline('=')<cr>
nnoremap <silent> <Plug>UnderlineH2 :call <sid>SmartUnderline('-')<cr>
nnoremap <silent> <Plug>SmartLevelThreeHeader :call <sid>SmartLevelThreeHeader()<cr>
nmap <leader>u1 <Plug>UnderlineH1
nmap <leader>u2 <Plug>UnderlineH2
nmap <leader>u3 <Plug>SmartLevelThreeHeader


function! s:EslintAutofix()
  let saved_cursor = getpos(".")
  %! eslint_d --stdin --fix-to-stdout
  write
  call setpos('.', saved_cursor)
endfunction
command! EslintAutofix call <sid>EslintAutofix()
nnoremap <leader>rf :EslintAutofix<cr>

command! EslintDStop echo system('eslint_d stop')
nnoremap <leader>es :EslintDStop<cr>

function! s:CopyFilename()
  let @" = expand('%') . "\n"
  echohl String | echo 'Filepath copied' | echohl None
endfunction
command! CopyFilename call <sid>CopyFilename()
nnoremap <leader>yp :CopyFilename<cr>


" END headers

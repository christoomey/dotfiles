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

nnoremap <leader>df :Files ~/dotfiles-local<cr>

nnoremap <leader>; :

" automatically rebalance windows on vim resize
autocmd VimResized * :wincmd =

" zoom a vim pane, <C-w>= to re-balance
nnoremap <leader>- :wincmd _<cr>:wincmd \|<cr>
nnoremap <leader>= :wincmd =<cr>



nnoremap <leader>dot :Files ~/dotfiles-local<cr>

nmap <leader>bi :PlugInstall<cr>



nnoremap Q @q


set hlsearch
set incsearch
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

" END headers

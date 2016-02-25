" Trusted Local
"--------------

autocmd VimEnter * call s:LoadTrustedVimrcLocal()
autocmd BufWrite .vimrc.local call s:TrustVimrcLocal()

let s:GIT_SAFE_DIR = ".git/safe"
let s:VIMRC_LOCAL = ".vimrc.local"

function! s:TrustVimrcLocal()
  if isdirectory('.git') && !isdirectory(s:GIT_SAFE_DIR)
    call mkdir(s:GIT_SAFE_DIR)
  endif
endfunction

function! s:LoadTrustedVimrcLocal()
  let s:trused_local_path = s:GIT_SAFE_DIR . "/../../" . s:VIMRC_LOCAL
  if filereadable(s:trused_local_path)
    execute "source " . s:trused_local_path
  endif
endfunction

function! s:EditVimrcLocal()
  execute "edit " . s:VIMRC_LOCAL
endfunction

function! s:SmartSplit()
  let split_cmd = (winwidth(0) >= 100) ? 'vsplit' : 'split'
  execute split_cmd . " " . s:VIMRC_LOCAL
endfunction

command! EditVimrcLocal call <sid>SmartSplit()

nnoremap <leader>el :EditVimrcLocal<cr>

" vim:ft=vim

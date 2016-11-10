" Project Notes
" -------------

" Quick access to a local notes file for keeping track of things in a given
" project. Add `.project-notes` to global ~/.gitignore

let s:PROJECT_NOTES_FILE = '.project-notes'

command! EditProjectNotes call <sid>SmartSplit(s:PROJECT_NOTES_FILE)
nnoremap <leader>ep :EditProjectNotes<cr>

autocmd BufEnter .project-notes call <sid>LoadNotes()

function! s:SmartSplit(file)
  let split_cmd = (winwidth(0) >= 100) ? 'vsplit' : 'split'
  execute split_cmd . " " . a:file
endfunction

function! s:LoadNotes()
  setlocal filetype=markdown
  nnoremap <buffer> q :wq<cr>
endfunction

" vim:ft=vim

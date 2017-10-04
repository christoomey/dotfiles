" fzf - the fuzzy finder of all the things

Plug 'junegunn/fzf', { 'dir': '~/.fzf', 'do': './install --all' }
Plug 'junegunn/fzf.vim'
let g:fzf_files_options =
  \ '--reverse ' .
  \ '--preview "(coderay {} || cat {}) 2> /dev/null | head -'.&lines.'"'
nnoremap <C-p> :Files<cr>
let $FZF_DEFAULT_COMMAND = 'ag -g "" --hidden'

let branch_files_options = { 'source': 'branch_files' }
let uncommited_files_options = { 'source': 'branch_files -w' }

let s:diff_options =
  \ '--reverse ' .
  \ '--preview "(mdiff {} | tail -n +5 || cat {}) 2> /dev/null | head -'.&lines.'"'
command! BranchFiles call fzf#run(fzf#wrap('BranchFiles',
      \ extend(branch_files_options, { 'options': s:diff_options }), 0))
command! UncommitedFiles call fzf#run(fzf#wrap('UncommitedFiles',
      \ extend(uncommited_files_options, { 'options': s:diff_options }), 0))
nnoremap <silent> <leader>gp :BranchFiles<cr>
nnoremap <silent> <leader>GP :UncommitedFiles<cr>


nnoremap <leader>ga :Files app/<cr>
nnoremap <leader>gm :Files app/models/<cr>
nnoremap <leader>gv :Files app/views/<cr>
nnoremap <leader>gc :Files app/controllers/<cr>
nnoremap <leader>gy :Files app/assets/stylesheets/<cr>
nnoremap <leader>gj :Files app/assets/javascripts/<cr>
nnoremap <leader>gs :Files spec/<cr>

function! s:all_help_files()
  return join(map(split(&runtimepath, ','), 'v:val ."\/doc\/tags"'), ' ')
endfunction
let full_help_cmd = "cat ". s:all_help_files() ." 2> /dev/null \| grep -i '^[a-z]' \| awk '{print $1}' \| sort"

nnoremap <silent> <leader>he :Helptags<cr>

" vim:ft=vim

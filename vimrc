"===========================================================================
"                                 ~My vimrc~
"===========================================================================
" Author:        Chris Toomey <christopherjtoomey at the google mails>
" Source:        https://github.com/christoomey/dotfiles
"
" My vimrc, mostly for python development. Did my best to document
" inline for anyone who may end up reading this (especially me+=1year)
"---------------------------------------------------------------------

"---- PATHOGEN FOR PLUGIN ----
    " Pathogen lets each plugin have its own subdir for easy mgmt
    " Made by Tim Pope. Found here: github.com/tpope/vim-pathogen
    if !exists('g:vimrc_loaded')
        filetype off
        call pathogen#runtime_append_all_bundles()
        call pathogen#helptags()

        syntax on
        filetype plugin indent on
    endif

    " Search here for other plugins (github mirror of offical vim-scripts)
    " http://vim-scripts.org/scripts.html

    " Using this ruby script to install plugin bundles via vimrc comments
    " http://github.com/bronson/vim-update-bundles/

    " Map to bundles listing which is also a TOC for plugin helpfiles
    nnoremap <LEADER>hb :h bundles<CR>5j7l

"---- GENERAL SETUP STUF ----
    set hidden                      " Allow buffer change w/o saving
    set autoread                    " Load file from disk, ie for git reset
    set nocompatible		        " Not concerned with vi compatibility
    set backspace=indent,eol,start	" Sane backspace behavior
    set nobackup      		        " Do not automatically backup files
    set history=1000  		        " Remember last 1000 commands
    set nu                          " Show line numbers
    set scrolloff=4                 " Keep at least 4 lines below cursor
    set expandtab                   " Convert <tab> to spaces (2 or 4)
    set tabstop=4                   " Four spaces per tab as default
    set shiftwidth=4                "     then override with per filteype
    set softtabstop=4               "     specific settings via autocmd

    " Interacting with system clipboard
    " TODO get system clipboard interaction going
        " Map for pasting from clipboard
        " Toggles paste mode
        " Paste from clipboard
        " Toggle back

    " Toggle paste mode
    nmap <silent> ,p :set invpaste<CR>:set paste?<CR>

    " Disable sound/visual bell on errors
    set vb t_vb=

    " Auto change the directory to the current file I'm working on
    if version >= 703
        set autochdir " Need to use this vs 'lcd %:p:h' for fugitive to work
    else
        autocmd BufEnter * lcd %:p:h " Fights fugitive
    endif

    " If this is terminal vim, we should work from the current dir. Macvim ~/code
    " if has('mac')
        " if !exists('g:vimrc_loaded')
            " cd ~/work
        " endif
    " endif

    " Configure easy toggle between single and double width
    nnoremap <leader>7 :set columns=170<cr>
    nnoremap <leader>8 :set columns=85<cr>

    " Allow command line editing like emacs
    cnoremap <C-A>      <Home>
    cnoremap <C-B>      <Left>
    cnoremap <C-E>      <End>
    cnoremap <C-F>      <Right>
    cnoremap <C-N>      <End>
    cnoremap <C-P>      <Up>

"---- GUI SETTING ----
    if has('gui')
        colorscheme ctwombat
        set guioptions-=r "Hide the right side scrollbar
        set guioptions-=L "Hide the left side scrollbar
        set guioptions-=T "Hide toolbars...this is vim for craps sake
        set guioptions-=m "Hide the menu, see above

        " Size and position the window well (only perform on startup)
        if !exists('g:vimrc_loaded')
            set columns=85
            set lines=999
            winpos 999 5
        endif

        " Display a crosshair to highlight the cursor (hi row & col)
        set cursorline cursorcolumn

        " Version 7.3 of the vim, with things like margins
        if version >= 703
            " Add a margin on the left to highlight past 79
            hi ColorColumn guibg=#2d2d2d
            set colorcolumn=80,81,82
        endif

        " MacVim is very pretty
        if has('gui_macvim')
            set transparency=8

            " Fullscreen options
            set fuoptions=maxvert
            " au GUIEnter * set fullscreen
        endif

        " Set a pretty font
        if has('win32')
            set guifont=Consolas:h10
        elseif has('mac')
            if !exists('g:vimrc_loaded')
                set guifont=Consolas:h12
            endif
        endif
    else
        set nocursorline nocursorcolumn
        colorscheme ctslate
    endif


    " Use +/- directly to resize a window
    nnoremap - 10<C-w><
    nnoremap = 10<C-w>>
    " Easy access to maximizing
    nnoremap <C-_> <C-w>_

"---- SPEEDUP ----
    "Keep the leader close for many custom maps
    let mapleader = ","

    " I have almost no need for semicolons, but the colon, I love
    nnoremap ; :
    vnoremap ; :
    cnoremap ; :

    " Use j/k to scroll through autocomplete options
    inoremap <expr> <C-j> ((pumvisible())?("\<C-n>"):("j"))
    inoremap <expr> <C-k> ((pumvisible())?("\<C-p>"):("k"))
    " FIXME Fire off Supertab with C-j
    let g:SuperTabMappingForward = '<nul>'
    let g:SuperTabMappingBackward = '<s-nul>'
    let g:SuperTabMappingBackward = '<C-k>'
    let g:SuperTabMappingForward = '<C-j>'

    " For opening up HTML files in chrome for previewing
    nmap <leader>op :silent !open %<CR>

    " Do not use <Ctrl-c> to break out to normal mode
    " Use C-Space to Esc out of any mode
    nnoremap <C-Space> <Esc>:noh<CR>
    vnoremap <C-Space> <Esc>gV
    onoremap <C-Space> <Esc>
    cnoremap <C-Space> <C-c>
    inoremap <C-Space> <Esc>`^
    " Terminal sees <C-@> as <C-space> WTF, but ok
    nnoremap <C-@> <Esc>:noh<CR>
    vnoremap <C-@> <Esc>gV
    onoremap <C-@> <Esc>
    cnoremap <C-@> <C-c>
    inoremap <C-@> <Esc>`^

    " Move current line up
    nnoremap <A-k> :m-2<cr>
    " Move current line down
    nnoremap <A-j> :m+<cr>
    " Move visual selection up
    vnoremap <A-k> :m-2<cr>gv
    " Move visual selection down
    vnoremap <A-j> :m'>+<cr>gv

    " Setup nice command tab completion
    set wildmenu
    set wildmode=list:longest,full
    set wildignore+=*.pyc

    " Easy vimrc moding ",v" loads vimrc for edit, ",V" sources it
    map <leader>rc :e ~/.vimrc<CR>
    map <silent> <leader>src :source %<CR>

    " Help File funtimes, <enter> to follow tag, delete for back
    au filetype help nnoremap <buffer><cr> <c-]>
    au filetype help nnoremap <buffer><bs> <c-T>
    au filetype help nnoremap <buffer>q :q<CR>
    au filetype help set nonumber
    set splitbelow " Split windows, ie Help, make more sense to me below
    au filetype help wincmd _ " Maximze the help on open

    " Better edit mapping. Includes current path, ie ":e /Users/<user>/path/"
    nmap <Leader>e :e <C-R>=expand("%:p:h") . "/" <CR>

    " Speedy window jumping
    nnoremap <C-j> <C-w>j
    nnoremap <C-h> <C-w>h
    nnoremap <C-k> <C-w>k
    nnoremap <C-l> <C-w>l

    "Make Y yank to end of line (like D, or C)
    nmap Y y$

    "Allow Ctrl-f to paste current register in insert mode
    imap <C-f> <C-r>"

    " Easy access to the start of the line
    nnoremap 0 ^

    " Quickly open a second window to view files side by side
    nmap <LEADER>vn :vnew<CR>
        \:set columns=170<CR>
        \<C-w>=

    " Thesaurus lookup for the |word| under the cursor
    nnoremap <LEADER>th :silent !open
        \ http://thesaurus.com/browse/<cword><CR>

    " Move one line at a time, aka 'fine ajdustment'
    nmap j gj
    nmap k gk

    " Move four lines at a time, 'medium adjustment'
    nnoremap <C-j> 10j
    nnoremap <C-k> 10k

    " Better visual indentation control (reslect region after shift)
    vnoremap > >gv
    vnoremap < <gv

    " Reselect pasted text. Mnem: 'Get pasted'
    nnoremap gp `[v`]

    " Go to position of last edit. Mnem: 'Go to Edit'
    nnoremap ge `.

    " Quick insertion of a datetime stamp for daily log
    nnoremap <LEADER>dt :r! date +"\%Y-\%m-\%d \%H:\%M:\%S"

"---- DIFF SETTINGS ----
    " TODO link the filetypes (buffer[2].&ft = buffer[1].&ft)
    " TODO set the statusbar for the original to say '---ORIGINAL---'
    " au FilterWritePre * if &diff | echo 'Hello' | endif
    if &diff
        set columns=170
        vert res=85
        nnoremap <C-d> ]c
        nnoremap <C-u> [c
        nnoremap q :qa!<cr>
        if !has('gui')
            hi diffadd ctermbg=0
            hi diffdelete ctermbg=0
            hi diffchange ctermbg=0
            hi difftext ctermbg=1
            hi folded ctermbg=8
        endif
    endif

"---- SEARCH STUFF ----
    " Searching stuff
    set hlsearch                    " hilight searches, map below to clear
    nohlsearch                      " kill highliting on vimrc reload
    set incsearch                   " do incremental searching
    set ignorecase                  " Case insensitive...
    set smartcase                   " ...except if you use UCase
    nmap <F4> :silent noh<CR>
    nnoremap <LEADER>rh :silent noh<CR>

    " Replace settings
    set gdefault "Replace all occurences of pattern in a line

    " Mappings for quick search & replace. Global set to default
    " In order, they are `all`, `visual`, `confirm`, `word`
    map <Leader>rpa :%s///<Left>
    map <Leader>rpv :s//<Left>
    map <Leader>rpc :%s//c<Left><Left>
    map <Leader>rpw :%s/<C-r><C-w>//c<Left><Left>

    " TODO get the awk search functional for moi

    " Vimgrep for the last search
    nmap <silent> ,gs
         \ :vimgrep /<C-r>// %<CR>:ccl<CR>:cwin<CR><C-W>J:set nohls<CR>

    " Vimgrep for the word under the cursor
    nmap <silent> ,gw
         \ :vimgrep /<C-r><C-w>/ %<CR>:ccl<CR>:cwin<CR><C-W>J:set nohls<CR>

    " Vimgrep for the current WORD under cursor
    nmap <silent> ,gW
         \ :vimgrep /<C-r><C-a>/ %<CR>:ccl<CR>:cwin<CR><C-W>J:set nohls<CR>

    nnoremap <LEADER>rdb :g/import\ ipdb/normal dd

"---- PYTHON SETTINGS ----
    "TODO Get PEP8.vim and or pylint
    au FileType python setl sw=4 sts=4 ts=4 " Four spaces per tab

    "Run the current file
    nnoremap <LEADER>prun :! python ./%<CR>

    " TODO get ahold of a better syntax file. Ref gary B. Spec, hilite '%s'

"---- RUBY SETTINGS ----
    "Ruby is new, and sometimes funtimes
    au FileType ruby setl sw=2 sts=2 ts=2 " Two spaces per tab

    " Set .erb html files
    au FileType ebury setl setl sw=2 sts=2 ts=2 " Two spaces per tab

    " System `touch` to have autotest run
    nnoremap <LEADER>touch :silent !touch %<CR>

"---- FOLDING AND INDENT OPTION ----
    "Enable indent folding
    set foldenable
    set fdm=indent

    " So I never use s, and I want a single key map to toggle folds, thus
    " lower s = toggle <=> upper S = toggle recursive
    nnoremap s za
    nnoremap S zA

    "Maps for folding, unfolding all
    nnoremap <LEADER>fu zM<CR>
    nnoremap <LEADER>uf zR<CR>

    "Maps for indenting / outdenting
    nnoremap <RIGHT> >>
    nnoremap <LEFT> <<
    vmap <RIGHT> >
    vmap <LEFT> <

"---- OS SPECIFIC ----
    "May need to use the following line for linux
    "let uname = substitute(system("uname"),"\n","","g")
    "
    "Hopefully can just use this switch sequence
    "if has("win32")
    "...
    "elseif has("unix")
    "...
    "elseif has("macunix")
    "...
    "elseif has("amiga")
    "...
    "endif

"---- TESTING / REFACTORING RELATED ----
    " Create a nice long test name without having to enter the undersocre
    function! Test_Function()
        let test_name = getline('.')
        if test_name != ""
            let new_test_name = substitute(test_name, "\\s", "_", "g")
            execute "normal ccdef " . new_test_name . "():"
        endif
    endfunction
    nmap <LEADER>tf :call Test_Function()<CR>

    " Borrowed from Gary Bernhardt's vimrc
    function! ExtractVariable()
        echohl String
        let name = input("Variable name: ")
        if name == ''
            return
        endif

        " Enter visual mode (input() takes us out of it)
        normal! gv

        " Replace selected text with the variable name
        exec "normal c" . name
        " Define the variable on the line above
        exec "normal! O" . name . " = "
        " Paste the original selected text to be the variable value
        normal! $p
    endfunction
    vnoremap <LEADER>ev :call ExtractVariable()<cr>

    function! Git_Repo_Cdup() " Get the relative path to repo root
        "Ask git for the root of the git repo (as a relative '../../' path)
        let git_top = system('git rev-parse --show-cdup')
        let git_fail = 'fatal: Not a git repository'
        if strpart(git_top, 0, strlen(git_fail)) == git_fail
            " Above line says we are not in git repo. Ugly. Better version?
            return ''
        else
            " Return the cdup path to the root. If already in root,
            " path will be empty, so add './'
            return './' . git_top
        endif
    endfunction

    function! RunDjangoTests()
        let cdup = Git_Repo_Cdup()
        if cdup != ''
            execute ":cd " . cdup
            let test_results = system('python manage.py test --verbosity 0 -x')
            echo test_results
        else
            echohl ErrorMsg
            echo 'Not in a git repo'
        endif
    endfunction
    nmap <LEADER>ts :call RunDjangoTests()<cr>

    function! FeedbackBar(type, msg)
        " Produce a red bar on error with ErrorMsg, else green bar
        " TODO set a global var g:last_good_time each time we hit green bar
        " then create a function to revert (:earlier g:last_good_time) so I
        " will always have the fallback of a known good configuration, kinda
        " like a rock climber doing lead climbing
        hi GreenBar guifg=black guibg=green
        hi YellowBar guifg=black guibg=lightgray
        hi RedBar guifg=black guibg=red
        if a:type == "red"
            echohl RedBar
        elseif a:type == "green"
            let g:last_green_time = strftime('%s') "seconds since epoch
            echohl GreenBar
        else
            echohl YellowBar
        endif
        echon a:msg repeat(" ", &columns - strlen(a:msg) - 1)
        echohl None
    endfunction

"---- PLUGIN OPTIONS ----
    " A__BUNDLE: git://github.com/scrooloose/nerdcommenter.git
        let g:NERDCreateDefaultMappings = 0
        nmap <LEADER>cm <plug>NERDCommenterToggle
        vmap <LEADER>cm <plug>NERDCommenterToggle
        let NERDSpaceDelims=1 " Add a space before the comment
    " A__BUNDLE: git://github.com/scrooloose/nerdtree.git
        nnoremap <LEADER>nt :NERDTreeToggle<CR>
        " TODO remap the keys for better motion in the tree (space=>fold)

    " A__BUNDLE: git://github.com/msanders/snipmate.vim.git
        let g:snips_author = 'Chris Toomey'
        let g:snippets_dir = '$HOME/.vim/snippets'
        "XXX consider a switch to xptemplate

    " A__BUNDLE: git://github.com/vim-scripts/TaskList.vim.git
        "Need to remap this before Command-T, or it barks
        map <leader>tl <Plug>TaskList

    " A__BUNDLE: git://github.com/vim-scripts/ScrollColors.git

    " A__BUNDLE: git://github.com/edsono/vim-bufexplorer.git
        nmap <LEADER>be :BufExplorer<CR>
        let g:bufExplorerDefaultHelp=1       " Show default help.
        let g:bufExplorerDetailedHelp=0      " Don't show detailed help.
        let g:bufExplorerShowRelativePath=1  " Show relative paths.
        let g:bufExplorerSortBy='mru'        " Sort by most recently used.
        let g:bufExplorerSplitOutPathName=1  " Split the path and file

    " A__BUNDLE: git://github.com/bronson/vim-trailing-whitespace.git
        nmap <LEADER>wht :FixWhitespace<CR>

    " A__BUNDLE: git://github.com/vim-scripts/kwbdi.vim.git
        " Use <LEADER>bd to dump buffer w/o closing window

    " A__BUNDLE: git://github.com/vim-scripts/YankRing.vim.git
        " Use <Ctrl-p> to cycle back to earlier yanks after a paste
        let g:yankring_history_file = '.yankring_history'

    " A__BUNDLE: git://github.com/mileszs/ack.vim.git
        "TODO stop vim from jumping to the first match
        nmap <LEADER>a :call AckProject()<CR>
        nmap <LEADER>\ack :Ack<space>
        function! AckProject()
            let cdup = Git_Repo_Cdup()
            if cdup != ''
                " Move working dir to root of repo, then Ack
                execute ":cd " . cdup
                normal ,\ack
            else
                " Not in a git repo, work from cwd
                normal ,\ack
            endif
        endfunction

    " A__BUNDLE: git://github.com/ervandew/supertab.git
        let g:SuperTabDefaultCompletionType = 'context'

    " A__BUNDLE: git://github.com/tpope/vim-surround.git

    " A__BUNDLE: git://github.com/vim-scripts/IndexedSearch.git

    " A__BUNDLE: git://github.com/christoomey/vim-space.git
        " TODO Need to reclaim the ; normal map from this
        " Unfortunately I have to unhook ';' in the plugin itself
        " to avoid conflict with how I `nnoremap ; :`

    " A__BUNDLE: git://github.com/tpope/vim-fugitive.git
        nmap <LEADER>gs :Gstatus<CR>
        nmap <LEADER>gd :Gdiff<CR>

    " A__BUNDLE: git://github.com/wincent/Command-T.git
        " Ref wildignore setting above for filtering, relative path setting
        let g:CommandTCancelMap='<C-space>'
        let g:CommandTMatchWindowAtTop=1
        function! Command_T_Local()
            let cdup = Git_Repo_Cdup()
            if cdup != ''
                " Move working dir to root of repo, then CommandT
                execute ":cd " . cdup
                CommandTFlush
                CommandT
            else
                " Not in a git repo, default to ~/work
                call Command_T_Work()
            endif
        endfunction
        function! Command_T_Work() "Go from the ~/work repo
            cd ~/work
            CommandT
        endfunction
        nmap <LEADER>fow :call Command_T_Work()<CR>
        nmap <LEADER>fop :call Command_T_Local()<CR>
        nmap <LEADER>ctb :CommandTBuffer<CR>
        nmap <LEADER>ctf :CommandTFlush<CR>

    " A__BUNDLE: git://github.com/nathanaelkane/vim-indent-guides.git
        let g:indent_guides_guide_size = 1
        let g:indent_guides_enable_on_vim_startup = 1
        let g:indent_guides_start_level = 2

    " Can't figure out the issue here. Had to load the old
    " fahioned way by putting the script into a plugin dir
    " #BUNDLE: git://github.com/vim-scripts/MRU.git
        " nmap <LEADER>mru :MRU<SPACE>


    " For me to try
    "--------------
    " For speedy HTMLing (using zen coding)
    " #BUNDLE: git://github.com/rstacruz/sparkup.git
    "
    " TODO create a plugin that combines LustyJuggler with BufExplorer
    " --Display in Ctrl-D windlmenu style, not new buffer (ref snipmate)
    " --Display very similar to BufExplorer (columns, path, etc)
    " --Use asdfghjkl lusty style selection (with color confirm)
    " --http://www.vim.org/scripts/script.php?script_id=2050
    "
    " TODO try out this updated version of rainbow parens
    " http://bitbucket.org/sjl/dotfiles/src/tip/vim/bundle/rainbow/
    "
    " TODO pyflakes for code checking. Needs vim:Python == sys:Python
    "
    " TODO get scrooloses's syntastic checking in there for statusline
    " git://github.com/scrooloose/syntastic.git
    "
    " TODO pep8.vim or pylint
    "
    " TODO Launch unit tests from vim, errors in quickfix
    " #BUNDLE: git://github.com/nvie/vim-pyunit.git
    "
    " TODO rope.vim or similar refactoring plugin.
    "
    " TODO surround.vim coupled with repeat.vim
    "
    " TODO XP Template to replace snipmate as the best in class snippeter
    "
    " TODO Check out this for replacing pager in *nix environment
    " #BUNDLE: git://github.com/vim-scripts/vimpager.git


    " Cancelled
    "-----------
    " Didn't seem to work, so I got rid of it
    " #BUNDLE: git://github.com/vim-scripts/Rainbow-Parenthesis.git

"---- CUSTOM FUNCTION ----
    "Underline a vimrc section title
    nmap <LEADER>urc yypVr-r"

    " TODO Create a func to generate a table of contents with tags for vimrc
    " Ref: http://www.vim.org/scripts/script.php?script_id=2014

    " Simple wrapper around system mkdir cmd
    function! Mkdir()
        let dir = input('Dir name: ')
        let dir_exists = isdirectory(dir)

        " Check to see if the dir exists, abort if so, but inform user
        if dir_exists
            echon 'Error, dir ['
            echohl WarningMsg
            echon dir
            echohl None
            echon '] already exists in path ['
            echohl vimString
            echon getcwd()
            echohl None
            echon ']'
        else
            let retval = system('mkdir ' . dir)
            echon 'Created directory ['
            echohl vimString
            echon dir
            echohl None
            echon '] under root ['
            echohl vimString
            echon getcwd()
            echohl None
            echon ']'
        endif
    endfunction
    nnoremap <LEADER>md :call Mkdir()<CR>

    " List out the syntax for the word under the cursor
    function! SyntaxGroup()
        let hl = synIDattr(synID(line("."),col("."),1),"name")
        let trans = synIDattr(synID(line("."),col("."),0),"name")
        let low = synIDattr(synIDtrans(synID(line("."),col("."),1)),"name")
        " for id in synstack(line("."), col("."))
           " echo synIDattr(id, 'name')
        " endfor
        echo 'hl<' . hl . '> trans<' . trans . '>low<' . low . '>'
    endfunction
    nnoremap <LEADER>syn :call SyntaxGroup()<CR>

    function! ListMapShaddows()
         silent! redir @a
         silent! nmap <LEADER>
         silent! redir END
         silent! new
         silent! put! a
         silent! g/^s*$/d
         silent! %s/^.*,//
         silent! normal ggVg
         silent! sort
         silent! let lines = getline(1,"$")
         silent! call append("$", "--------------------")
         silent! call append("$", lines)
        " let lines = getline(1,"$")
        " for line in lines
    endfunction

"---- STATUSLINE ----
    " For general info on statusline, start with the :h, then see
    " http://got-ravings.blogspot.com/2008/08/vim-pr0n-making-statuslines-that-own.html
    " NOTE: NSFW, but very good overview of statusling configuration

    set laststatus=2 " Always show the statusline

    "define 3 custom highlight groups
    hi User1 ctermbg=lightgray ctermfg=yellow guifg=orange guibg=#444444 cterm=bold gui=bold
    hi User2 ctermbg=lightgray ctermfg=red guifg=#dc143c guibg=#444444 gui=none
    hi User3 ctermbg=lightgray ctermfg=red guifg=#ffff00 guibg=#444444 gui=bold

    set statusline= " Clear the statusline for vimrc reloads

    set stl=%*                       " Normal statusline highlight
    set stl+=%{InsertSpace()}        " Put a leading space in
    set stl+=%1* 				     " Red highlight
    set stl+=%{HasPaste()}           " Red show paste

    set stl+=%*                      " Return to normal stl hilight
    set stl+=%t                      " Filename

    set stl+=%2* 				     " Red highlight
    set stl+=%m                      " Modified flag

    set stl+=%*                      " Return to normal stl hilight
    set stl+=%r                      " Readonly flag
    set stl+=%h                      " Help file flag

    set stl+=%*                      " Set to 3rd highlight
    set stl+=\ %y                    " Filetype

    " Visual feedback of time since last diff
    set stl+=\ \ \ \ Last\ Green:
    set stl+=%*
    set stl+=\ %{TimeSinceGreen('low')}
    set stl+=%1*
    set stl+=%{TimeSinceGreen('medium')}
    set stl+=%2*
    set stl+=%{TimeSinceGreen('high')}
    set stl+=%3*
    set stl+=%{TimeSinceGreen('na')}
    set stl+=%*

    set stl+=%=                      " Right align from here on
    set statusline+=%{SlSpace()}     " Vim-space plugin current setting
    set stl+=\ \ Col:%c              " Column number
    set stl+=\ \ Line:%l/%L          " Line # / total lines
    set stl+=\ \ %P%{InsertSpace()}  " Single space buffer

    function! SlSpace()
        if exists("*GetSpaceMovement")
            return "[" . GetSpaceMovement() . "]"
        else
            return ""
        endif
    endfunc

    function! InsertSpace()
        " For adding trailing spaces onto statusline
        return ' '
    endfunction

    function! HasPaste()
        if &paste
            return '[PASTE]'
        else
            return ''
        endif
    endfunction

    function! CurDir()
        let curdir = substitute(getcwd(), '/Users/nation/', "~/", "g")
        return curdir
    endfunction

    "TODO Define equivalent hilights for the terminal usage
    "TODO get this hooked up with run_tests func and FeedbackBar('green')
    function! TimeSinceGreen(var)
        " Really ugly, but it works. See the associated ugliness in the
        " statusline section for where this gets used

        "Setup the time difference value for use in responses
        if exists('g:last_green_time')
            let current_time = strftime('%s') " Using seconds since epoch
            let diff = current_time - g:last_green_time
            let interval = 5
            let diff_minutes = diff / 6
        endif

        "Many cases for return values
        if a:var == 'low' "Check for normal time, 0-5 minutes
            if exists('g:last_green_time')
                if diff_minutes <= interval
                    return diff_minutes
                else
                    return ''
                endif
            else
                return ''
            endif
        elseif a:var == 'medium' "Start to get worried
            if exists('g:last_green_time')
                if diff_minutes > interval && diff_minutes <= interval * 2
                    return diff_minutes
                else
                    return ''
                endif
            else
                return ''
            endif
        elseif a:var == 'high' " high warning level
            if exists('g:last_green_time')
                if diff_minutes > interval * 2
                    return diff_minutes
                else
                    return ''
                endif
            else
                return ''
            endif
        elseif a:var == 'na'
            if exists('g:last_green_time')
                return ''
            else
                return '--'
            endif
        endif
    endfunction

"---- REFERENCES ----
    " A number of sources I used when compiling my vimrc

    " Swaroop's great, free, intro to vim. Where I started
    " http://www.swaroopch.com/notes/Vim
    "
    " Probably the best blog post I have read on vim.
    " http://stevelosh.com/blog/2010/09/coming-home-to-vim/
    "
    " An insanely long list of lesser known/used vimery
    " http://rayninfo.co.uk/vimtips.html
    "
    " A great stack overflow bit on groking vi
    " http://stackoverflow.com/questions/1218390/what-is-your-most-productive-shortcut-with-vim/1220118#1220118
    "
    " The best commented vimrc I have seen
    " http://www.vi-improved.org/vimrc.php
    "
    " Another dense well commented vimrc
    " http://www.derekwyatt.org/vim/the-vimrc-file/my-vimrc-file/
    "
    " Amix the lucky stiff's huge vimrc
    " http://amix.dk/vim/vimrc.html
    "
    " Another good one from the creator of vimcasts
    " https://github.com/nelstrom/dotfiles/raw/master/vimrc

"---- LOADED VARIABLE ----
    "Use this to prevent some settings from reloading
    let g:vimrc_loaded = 1

"---- GTD RELATED STUFF ----
    function! InsertDate()
        let date = strftime('%Y-%m-%d') " Grab date as 2010-09-17 format
        exe 'normal A [' . date . ']'
    endfunction
    nnoremap <LEADER>ad :call InsertDate()<cr>

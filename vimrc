"===========================================================================
"                                 ~My vimrc~
"===========================================================================
" Author:        Chris Toomey <christopherjtoomey at the google mails>
" Source:        http://bitbucket.org/christoomey/dotfiles
"
" My vimrc, mostly for python development. Did my best to document
" inline for anyone who may ended up reading this (especially me+=1year)
"---------------------------------------------------------------------

"---- PATHOGEN FOR PLUGIN ----
    " Pathogen lets each plugin have its own subdir for easy mgmt
    " Made by Tim Pope. Found here: github.com/tpope/vim-pathogen
    filetype off
    call pathogen#runtime_append_all_bundles()
    call pathogen#helptags()

    syntax on
    filetype plugin indent on

    " Search here for other plugins (github mirror of offical vim-scripts)
    " http://vim-scripts.org/scripts.html

    " Using this ruby script to install plugin bundles via vimrc comments
    " http://github.com/bronson/vim-update-bundles/

    " Map to bundles listing which is also a TOC for plugin helpfiles
    nnoremap <LEADER>hb :h bundles<CR>5j7l

"---- GENERAL SETUP STUF ----
    set hidden                      " Allow buffer change w/o saving
    set nocompatible		        " Not concerned with vi compatibility
    set backspace=indent,eol,start	" Sane backspace behavior
    set nobackup      		        " Do not automatically backup files
    set history=1000  		        " Remember last 1000 commands
    set nu                          " Show line numbers
    set scrolloff=4                 " Keep at least 4 lines below cursor

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
    autocmd BufEnter * lcd %:p:h

    " Allow command line editing like emacs
    cnoremap <C-A>      <Home>
    cnoremap <C-B>      <Left>
    cnoremap <C-E>      <End>
    cnoremap <C-F>      <Right>
    cnoremap <C-N>      <End>
    cnoremap <C-P>      <Up>

"---- GUI SETTING ----
    if has('gui')
        colorscheme wombatMine
        set guioptions-=r "Hide the right side scrollbar
        set guioptions-=L "Hide the left side scrollbar
        set guioptions-=T "Hide toolbars...this is vim for craps sake
        set guioptions-=m "Hide the menu, see above

        " Size and position the window well
        set columns=85
        set lines=999
        winpos 999 5

        " Display a crosshair to highlight the cursor (hi row & col)
        set cursorline! cursorcolumn!

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
            au GUIEnter * set fullscreen
        endif

        " Set a pretty font
        if has('win32')
            set guifont=Consolas:h10
        elseif has('mac')
            set guifont=Consolas:h12
            nnoremap <LEADER>ext :set guifont=Consolas:h14<CR>
            nnoremap <LEADER>lap :set guifont=Consolas:h12<CR>
        endif
    else
        colorscheme slate
    endif


    " Use +/- directly to resize a window
    nnoremap - <C-w>-
    nnoremap = <C-w>+
    " Easy access to maximizing
    nnoremap <C-_> <C-w>_

"---- SPEEDUP ----
    "Keep the leader close for many custom maps
    let mapleader = ","

    " I have almost no need for semicolons, but the colon, I love
    nnoremap ; :
    vnoremap ; :
    inoremap ; :
    cnoremap ; :

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

    " Setup nice command tab completion
    set wildmenu
    set wildmode=list:longest,full
    set wildignore+=*.pyc

    " Easy vimrc moding ",v" loads vimrc for edit, ",V" sources it
    map <leader>rc :e ~/.vimrc<CR>
    map <silent> <leader>src :source ~/.vimrc<CR>
            \:filetype detect<CR>:exe ":echo 'vimrc reloaded'"<CR>

    " Help File funtimes, <enter> to follow tag, delete for back
    au filetype help nnoremap <buffer><cr> <c-]>
    au filetype help nnoremap <buffer><bs> <c-T>
    au filetype help nnoremap <buffer>q :q<CR>
    au filetype help set nonumber
    set splitbelow " Split windows, ie Help, make more senset to me below
    " TODO Get help buffers to maximize on open
    " au filetype help resize 999
    " Mappings for quick maximizing or equaling btwn windows
    nmap <LEADER>max <C-w>_
    nmap <LEADER>eq <C-w>=

    " Speedy window jumping
    nnoremap <C-j> <C-w>j
    nnoremap <C-h> <C-w>h
    nnoremap <C-k> <C-w>k
    nnoremap <C-l> <C-w>l

    "Make Y yank to end of line (like D, or C)
    nmap Y y$

    " And while we are at, vv for visual line, V for visual to end of line
    nnoremap vv V
    nnoremap V v$

    "Allow Ctrl-f to paste current register in insert mode
    imap <C-f> <C-r>"

    " Easy access to the start of the line
    nnoremap 0 ^

    " Quickly open a second window to view files side by side
    nmap <LEADER>vn :vnew<CR>
        \:set columns=170<CR>
        \<C-w>=

    " Shell `Open` the WORD under the cursor. Mostly for URLs
    nnoremap <LEADER>op :silent !open <cWORD><CR>
    "TODO wrap the open map above in a function to handle windows
    "
    " Thesaurus lookup for the |word| under the cursor
    nnoremap <LEADER>th :silent !open
        \http://thesaurus.com/browse/<cword><CR>

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

    " Quick insertion of a datetime stamp for daily log
    nnoremap <LEADER>dt :r! date +"\%Y-\%m-\%d \%H:\%M:\%S"

"---- DIFF SETTINGS ----
    " TODO link the filetypes (buffer[2].&ft = buffer[1].&ft)
    " TODO set the statusbar for the original to say '---ORIGINAL---'
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
    map <Leader>rpa :%s//<Left>
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

"---- PEP 8 SETTING ----
    "TODO Get PEP8.vim and or pylint
    set tabstop=4
    set shiftwidth=4
    set expandtab

"---- FOLDING AND INDENT OPTION ----
    "Enable indent folding
    set foldenable
    set fdm=indent

    "Set space to toggle a fold
    "TODO can we get recursive toggle?

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

"---- PLUGIN OPTION ----
    " BUNDLE: git://github.com/scrooloose/nerdcommenter.git
        nmap <F2> <plug>NERDCommenterToggle
        let NERDSpaceDelims=1 " Add a space before the comment
    " BUNDLE: git://github.com/scrooloose/nerdtree.git
        nnoremap <LEADER>nt :NERDTreeToggle<CR>
        " TODO remap the keys for better motion in the tree (space=>fold)

    " BUNDLE: git://github.com/msanders/snipmate.vim.git
        let g:snips_author = 'Chris Toomey'

    " BUNDLE: git://github.com/vim-scripts/TaskList.vim.git

    " BUNDLE: git://github.com/vim-scripts/ScrollColors.git

    " BUNDLE: git://github.com/edsono/vim-bufexplorer.git
        nmap <LEADER>be :BufExplorer<CR>
        let g:bufExplorerDefaultHelp=1       " Show default help.
        let g:bufExplorerDetailedHelp=0      " Don't show detailed help.
        let g:bufExplorerShowRelativePath=1  " Show relative paths.
        let g:bufExplorerSortBy='name'       " Sort by the buffer's name.
        let g:bufExplorerSplitOutPathName=1  " Split the path and file

    " BUNDLE: git://github.com/bronson/vim-trailing-whitespace.git
        nmap <LEADER>wht :FixWhitespace<CR>

    " BUNDLE: git://github.com/vim-scripts/kwbdi.vim.git
        " Use <LEADER>bd to dump buffer w/o closing window

    " BUNDLE: git://github.com/vim-scripts/YankRing.vim.git
        " Use <Ctrl-p> to cycle back to earlier yanks after a paste

    " BUNDLE: git://github.com/mileszs/ack.vim.git
        "TODO stop vim from jumping to the first match
        nmap <LEADER>a :Ack<SPACE>

    " BUNDLE: git://github.com/ervandew/supertab.git
        let g:SuperTabDefaultCompletionType = 'context'

    " BUNDLE: git://github.com/tpope/vim-surround.git

    " BUNDLE: git://github.com/vim-scripts/IndexedSearch.git

    " BUNDLE: git://github.com/scrooloose/vim-space.git
        " TODO Need to reclaim the ; normal map from this
        " Unfortunately I have to unhook ';' in the plugin itself
        " to avoid conflict with how I `nnoremap ; :`


    " Can't figure out the issue here. Had to load the old
    " fahioned way by putting the script into a plugin dir
    " #BUNDLE: git://github.com/vim-scripts/MRU.git
        " nmap <LEADER>mru :MRU<SPACE>


    " For me to try
    "--------------
    " For speedy HTMLing (using zen coding)
    " #BUNDLE: git://github.com/rstacruz/sparkup.git
    "
    " TODO LustyExplorer or CommandT for quick file open
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

    " Create a table of contents with tags for vimrc
    " TODO create this regex based function, then extract to script
    " Ref: http://www.vim.org/scripts/script.php?script_id=2014

    " Custom statusline format, and always show it
    " TODO setup a truly snazy statusline per scroolose (NERDTree)
    " http://got-ravings.blogspot.com/2008/08/vim-pr0n-making-statuslines-that-own.html
    set stl=\ %{HasPaste()}%t%m%r%h\ %w\ \ CWD:\ %r%{CurDir()}%h\ \ \ Line:\ %l/%L\ \ \ Col:%c
    set laststatus=2

    " TODO Create a plugin to launch http version of git bundle <cword>

    function! CurDir()
        let curdir = substitute(getcwd(), '/Users/nation/', "~/", "g")
        return curdir
    endfunction

    function! HasPaste()
        if &paste
            return 'PASTE MODE  '
        else
            return ''
        endif
    endfunction

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

"---- REFERENCE ----
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

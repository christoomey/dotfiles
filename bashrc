alias py='python'
alias ip='ipython'
alias ls='ls -FG' #Colorize and show folders, symlinks, etc
alias so='source' # Like vim
alias nose='nosetests'
alias rmpyc='rm *.pyc'

#Configure multi-line prompt
PS1='
$PWD
$=>'

function mkcd
{
    dir=$1;
    mkdir -p $dir && cd $dir;
}

# Vim split diff view function for modified hg file
# Additional settings applied in the vimrc
function hgd {
    vim -d $1 <(hg cat $1);
    cols 80
}

# MacVim Diff ftw
function mhgd {
    mvim -d $1 <(hg cat $1);
    return 0
}

# Positive integer test (including zero)
function positive_int() {
    return $(test "$@" -eq "$@" > /dev/null 2>&1 && test "$@" -ge 0 > /dev/null 2>&1);
}

# Set the terminal to the specified number of columns
function cols() {
    if [[ $# -eq 1 ]] && $(positive_int "$1"); then
        printf "\e[8;999;${1};t"
        return 0
    fi
    return 1
}

# Less severe than rm
function trash() { mv "$@" ~/.Trash; }

#TODO Switch to vi style keybindings. use "set -o vi"

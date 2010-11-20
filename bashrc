alias py='python'
alias ip='ipython'
alias ls='ls -FG' #Colorize and show folders, symlinks, etc
alias so='source' # Like vim
alias nose='nosetests'
alias rmpyc='rm *.pyc'
alias ..='cd ..'
alias pu='pushd'
alias po='popd'

# Django related aliases
alias djsh='python manage.py shell'
alias djru='python manage.py runserver'
alias djte='python manage.py test'

# Git bash completion
source ~/.git-completion.bash

# TODO Figure out how to highlight from within function. Specifically for
# git PS1 with staged, modified, and untracked as green, red, yellow
PS1='\n\e[33;1m\]$PWD'                # {YELLOW} working dir
PS1+='\e[31;1m\]$(jobcount)'          # {RED} Current background jobs
PS1+='\e[35;1m\]$(dirstack_count)'    # {VIOLET} Dirs in the stack
PS1+='\e[34;1m\]$(git_branch)'        # {BLUE} My fancy git PS1
PS1+='\e[32;1m\]$(git_staged)'        # {GREEN} My fancy git PS1
PS1+='\e[31;1m\]$(git_modified)'      # {RED} My fancy git PS1
PS1+='\e[33;1m\]$(git_untracked)'     # {YELLOW} My fancy git PS1
PS1+='\e[34;1m\]$(git_close_branch)'  # {VIOLET} My fancy git PS1
PS1+='\n\[\e[0m\]$=>'                 # Restore normal color

# List the number of background jobs. See next section for details
function jobcount {
    count=`jobs | wc -l | awk '{print $1}'`
    if [ $count -eq "0" ]; then
        echo ""
    else
        echo " j(+$count)"
    fi
}
function dirstack_count {
    count=`dirs | wc -w | awk '{print $1}'`
    let "count-=1" # Returns current dir in list, so less one
    if [ $count -eq "0" ]; then
        echo ""
    else
        echo " d(+$count)"
    fi
}
function git_branch {
    branch=`__git_ps1 "%s"`
    if [ -z "$branch" ]; then
        echo ""
    else
        echo " ($branch"
    fi
}
function git_staged {
    branch=`__git_ps1 "%s"`
    if [ -z "$branch" ]; then
        echo ""
    else
        staged=`git diff --cached --name-status | \
            wc -l | awk '{print $1}'`
        if [ $staged -gt "0" ]; then
            echo " $staged"
        else
            echo ""
        fi
    fi
}
function git_modified {
    branch=`__git_ps1 "%s"`
    if [ -z "$branch" ]; then
        echo ""
    else
        modified=`git diff --name-status | wc -l | awk '{print $1}'`
        if [ $modified -gt "0" ]; then
            echo " $modified"
        else
            echo ""
        fi
    fi
}
function git_untracked {
    branch=`__git_ps1 "%s"`
    if [ -z "$branch" ]; then
        echo ""
    else
        untracked=`git ls-files --other --exclude-standard | \
            wc -l | awk '{print $1}'`
        if [ $untracked -gt "0" ]; then
            echo " $untracked"
        else
            echo ""
        fi
    fi
}
function git_close_branch {
    branch=`__git_ps1 "%s"`
    if [ -z "$branch" ]; then
        echo ""
    else
        echo ")"
    fi
}

# Some fun with switching the foreground terminal process
# Use Ctrl-z to freeze current fg process (ie vim), do something else
# in the shell, ie manpage lookup, then use Ctrl-g to get back to vim
# ref: http://oinopa.com/2010/10/24/laptop-driven-development.html
# TODO Get Ctrl-f set as the freeze command
export HISTIGNORE="fg*"
bind '"\C-g": "fg %-\n"'

function mkcd {
    dir=$1;
    mkdir -p $dir && cd $dir;
}

function gitd {
    vim -d $1 <(git show HEAD:$1);
    cols 80
}

function mgitd {
    mvim -d $1 <(git show HEAD:$1);
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
    if [ $# -eq 1 ] && $(positive_int "$1"); then
        printf "\e[8;999;${1};t"
        return 0
    fi
    if [ $# -eq "0" ]; then
        printf "\e[8;999;80;t"
    fi
    return 1
}

# Less severe than rm
function trash() { mv "$@" ~/.Trash; }

#TODO Switch to vi style keybindings. use "set -o vi"

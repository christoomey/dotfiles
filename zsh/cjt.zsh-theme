# ------------------------------------------------------------------------
# Chris Toomey oh-my-zsh theme
# based on them by Juan G. Hurtado
# (Needs Git plugin for current_branch method)
# ------------------------------------------------------------------------

# Color shortcuts
RED=$fg[red]
YELLOW=$fg[yellow]
GREEN=$fg[green]
WHITE=$fg[white]
BLUE=$fg[blue]
RED_BOLD=$fg_bold[red]
YELLOW_BOLD=$fg_bold[yellow]
GREEN_BOLD=$fg_bold[green]
WHITE_BOLD=$fg_bold[white]
BLUE_BOLD=$fg_bold[blue]
RESET_COLOR=$reset_color

# Format for git_prompt_info()
ZSH_THEME_GIT_PROMPT_PREFIX=""
ZSH_THEME_GIT_PROMPT_SUFFIX=""

# Format for parse_git_dirty()
ZSH_THEME_GIT_PROMPT_DIRTY=" %{$RED%}(*)"
ZSH_THEME_GIT_PROMPT_CLEAN=""

# Format for git_prompt_ahead()
ZSH_THEME_GIT_PROMPT_AHEAD=" %{$RED%}(!)"

# Format for git_prompt_long_sha() and git_prompt_short_sha()
ZSH_THEME_GIT_PROMPT_SHA_BEFORE=" %{$WHITE%}[%{$YELLOW%}"
ZSH_THEME_GIT_PROMPT_SHA_AFTER="%{$WHITE%}]"


# Show job count for backgrounded jobs
function jobbies() {
    job_count=$#jobstates
    if [ $job_count -gt "0" ]; then
        echo ${job_count}
    else
        echo ""
    fi
}

# Only display host if this is via SSH
if [[ -n $SSH_CONNECTION ]]; then
    sshing="%n@%m"
else
    sshing=""
fi

PROMPT='
%{$GREEN_BOLD%}$sshing%{$WHITE%} %{$YELLOW%}${PWD/#$HOME/~}%{$RESET_COLOR%}$(jobbies) \
%{$BLUE%}$(current_branch)$(git_prompt_short_sha)$(parse_git_dirty)%{$RESET_COLOR%}
%{$BLUE%}>%{$RESET_COLOR%} '
RPROMPT='[%{$GREEN%}$(rvm-prompt v g)%{$WHITE%}]'

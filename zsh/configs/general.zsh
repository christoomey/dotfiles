# General ZSH configurations

export EDITOR="nvim"

alias vim=nvim
alias cl=claude

sz() { source ~/.zshrc }


first() { awk '{print $1}' }
second() { awk '{print $2}' }
sum() { paste -sd+ - | bc }

igrep() { grep -i $@ }

restart-postgres() {
  rm /usr/local/var/postgres/postmaster.pid && ( \
    cd ~/Library/LaunchAgents && \
      launchctl unload homebrew.mxcl.postgresql.plist && \
      launchctl load -w homebrew.mxcl.postgresql.plist \
  )
}

alias pods="KUBECONFIG=~/code/work/august/ops/k8s/kubeconfigs/branch k9s"

Dotfiles
========

My dotfiles, a constantly evolving set of configurations which I arguably spend
too much time tweaking, but they make the command line feel like home, so here
we are.

Vim
---

I do love me some Vim, that's for sure. I run Vim with [vim-plug][] to manage
plugins. I have _many_ plugins and customizations (stored in `vim/rcfiles` and
`vim/rcplugins` respectively) which may not be everyone's cup of tea, but I sure
do love a sharp tool.

[vim-plug]: https://github.com/junegunn/vim-plug


Zsh
---

I run zsh as my shell, finding it to be a great middle ground between adding
additional niceties and features, while remaining a largely compatible shell
scripting target. I use [zplug][] to manage zsh plugins (like the amazing
[zsh-syntax-highlighting][] plugin), and I use [pure][] as my prompt.

[zplug]: https://github.com/zplug/zplug
[zsh-syntax-highlighting]: https://github.com/zsh-users/zsh-syntax-highlighting
[pure]: https://github.com/sindresorhus/pure


Tmux
----

Tmux allows me to combine processes, shells, and Vim in any way I need for the
project at hand. I'm able to build my own IDE-like experience at the command
line while still using the best tool for any given job. I'm a big fan.

Core to my tmux work is the combination of two plugins that bring Vim & tmux
together, [vim-tmux-navigator][] for navigation, and [vim-tmux-runner][] for
sending commands from vim to tmux.

[vim-tmux-navigator]: https://github.com/christoomey/vim-tmux-navigator
[vim-tmux-runner]: https://github.com/christoomey/vim-tmux-runner


fzf
---

Lastly have [fzf][], "a command-line fuzzy finder". In the end this is a much
smaller component being just a shell command, but I find I use it across each of
Vim, zsh, and tmux, and it has become absolutely core to many of my workflows,
thus it gets top billing.

[fzf]: https://github.com/junegunn/fzf

Inspiration
-----------

- [thoughtbot](https://github.com/thoughtbot/dotfiles)
- [Ryanb](https://github.com/ryanb/dotfiles)
- [Gary Bernhardt](https://github.com/garybernhardt/dotfiles)
- [Rtomayko](https://github.com/rtomayko/dotfiles)

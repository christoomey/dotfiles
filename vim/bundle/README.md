# Plugin Management

All plugins are stored as single line bundle directives in my vimrc file, eg:

    " BUNDLE: scroloose/nerdtree

The bundles are loaded onto the vim runtime path at startup by
[Pathogen](https://github.com/tpope/vim-pathogen). I use a modified version of
the [vim-update-bundles](https://github.com/bronson/vim-update-bundles) to
install and remove plugin bundles. Since the vim-update-bundles plugin outputs a
bundles.txt file listing the current version of all the bundles, I treat the bundles
themselves as build output, and thus I .gitignore it. In the past I incldued the
plugins as git submodules, but I found the submodule functionality to be one thing in
git that I just didn't trust and thus I stopped using it.

**TL;DR** - Look into the PLUGIN OPTIONS section of my vimrc for the plugins I am
using. This bundle folder is intended to be empty as far as git is concerned.


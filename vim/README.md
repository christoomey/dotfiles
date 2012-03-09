## Installation

First get a hold of Vim. MacVim (via [Homebrew](http://mxcl.github.com/homebrew/)) on OS X is a great choice. [gVim](http://www.vim.org/download.php#pc) is good for windows.

### Unixy

``` bash
cd ~
ln -s ~/code/dotfiles/vim .vim
git clone https://github.com/gmarik/vundle.git .vim/bundles/vundle
```

### Windows

Ensure that the supporting prerequisites for Vundle (git & curl) are in place
before continuing. Use [this guide](https://github.com/gmarik/vundle/wiki/Vundle-for-Windows) for details.

``` cmd
REM Note: this all must be run as Administrator (specifically the mklink)
cd /d %USERPROFILE%
mklink /d .vim code\dotfiles\vim
mkdir code\dotfiles\vim\bundles
git clone https://github.com/gmarik/vundle.git code\dotfiles\vim\bundles\vundle
mklink _vimrc code\dotfiles\vim\vimrc
cd /d "\Program Files\Vim"
mklink /h _vimrc %USERPROFILE%\code\dotfiles\vim\windows_vimrc
REM Now open gVim and run :BundleInstall<cr>
```

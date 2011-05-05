#!/bin/sh
#######################################################
#  UNIX TREE                                          #
#  Version: 2.3                                       #
#  File: ~/apps/tree/tree.sh                          #
#                                                     #
#  Displays Structure of Directory Hierarchy          #
#  -------------------------------------------------  #
#  This tiny script uses "ls", "grep", and "sed"      #
#  in a single command to show the nesting of         #
#  sub-directories.  The setup command for PATH       #
#  works with the Bash shell (the Mac OS X default).  #
#                                                     #
#  Setup:                                             #
#     $ cd ~/apps/tree                                #
#     $ chmod u+x tree.sh                             #
#     $ ln -s ~/apps/tree/tree.sh ~/bin/tree          #
#     $ echo "PATH=~/bin:\${PATH}" >> ~/.profile      #
#                                                     #
#  Usage:                                             #
#     $ tree [directory]                              #
#                                                     #
#  Examples:                                          #
#     $ tree                                          #
#     $ tree /etc/opt                                 #
#     $ tree ..                                       #
#                                                     #
#  Public Domain Software -- Free to Use as You Like  #
#  http://www.centerkey.com/tree  -  By Dem Pilafian  #
#######################################################

echo
if [ "$1" != "" ]  #if parameter exists, use as base folder
   then cd "$1"
   fi
pwd
ls -R | grep ":$" |   \
   sed -e 's/:$//' -e 's/[^-][^\/]*\//--/g' -e 's/^/   /' -e 's/-/|/'
# 1st sed: remove colons
# 2nd sed: replace higher level folder names with dashes
# 3rd sed: indent graph three spaces
# 4th sed: replace first dash with a vertical bar
if [ `ls -F -1 | grep "/" | wc -l` = 0 ]   # check if no folders
   then echo "   -> no sub-directories"
   fi
echo
exit

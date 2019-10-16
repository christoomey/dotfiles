wvim() {
  command_name="$1"
  case "$(_definition_type "$command_name")" in
    "alias") _edit_shell_alias "$command_name";;
    "function") _edit_shell_function "$command_name";;
    "script") _edit_script "$command_name";;
    "git-alias") _edit_git_alias "$command_name";;
    "not found") _handle_unknown_command "$command_name"
  esac
}
compdef wvim=which

_handle_unknown_command() {
  echo "\"$1\" is not recognized as a script, function, or alias" >&2
  return 1
}

_definition_type() {
  arg_type_string=$(type -a "$1")
  if [[ "$arg_type_string" == *"is an alias"* ]]; then
    echo "alias"
  elif [[ "$arg_type_string" == *"is a shell function"* ]]; then
    echo "function"
  elif which "$1" > /dev/null 2>&1; then
    echo "script"
  elif _is_git_alias "$1"; then
    echo "git-alias"
  else
    echo "not found"
  fi
}

_is_git_alias() {
  echo "$1" | grep -q 'git-.*' && _git_aliases | grep -qF $(_git_alias_name "$1")
}

_git_alias_name() {
  echo "$1" | sed 's/^git-//'
}

_git_aliases() {
  git config --get-regexp ^alias\. |\
    sed -e s/^alias\.// -e s/\ /\ =\ / |\
    awk '{print $1}'
}

_edit_git_alias() {
  line=$(grep -En "^\s+$(_git_alias_name "$1") =" ~/.gitconfig | cut -d ":" -f 1)
  vim ~/.gitconfig "+$line"
}

_edit_script() {
  vim "$(which "$1")"
}

_definition_field_for_pattern() {
  definition=$(grep -En "$2" ~/.zshrc ~/.zshenv ~/.zsh/**/*.zsh 2>/dev/null)
  echo "$definition" | cut -d ":" -f "$1"
}

_edit_based_on_pattern() {
  file=$(_definition_field_for_pattern 1 "$1")
  line=$(_definition_field_for_pattern 2 "$1")
  vim "$file" "+$line"
}

_edit_shell_function() {
  _edit_based_on_pattern "(^function $1)|(^$1\(\))"
}

_edit_shell_alias() {
  _edit_based_on_pattern "alias $1="
}

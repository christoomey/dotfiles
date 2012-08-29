#!/usr/bin/env ruby

class String
  def unindent
    gsub(/^#{scan(/^\s*/).min_by{|l|l.length}}/, "")
  end
end

def die_unhappily(msg)
  puts msg
  exit 1
end

def usage_and_exit_happy
  usage = <<-eos
  newbie -- Minimal rails app creation script

  USAGE: newbie <app_name>
    app_name => Name of the rails app to be created, and associated rvm gemset
  eos
  puts usage
  exit 0
end

def create_powrc app_name
  powrc_str = <<-eos.unindent
  if [ -f "$rvm_path/scripts/rvm" ] && [ -f ".rvmrc" ]; then
    source "$rvm_path/scripts/rvm"
    source ".rvmrc"
  fi
  eos
  write_rcfile "#{app_name}/.powrc", powrc_str
end

def create_rvmrc app_name
  rvmrc_string = "rvm use 1.9.3@#{app_name} --create"
  write_rcfile "#{app_name}/.rvmrc", rvmrc_string
end

def write_rcfile path, text
  File.open(path, "w") {|f| f.write text}
end

def print_next_steps app_name
  next_steps = <<-eos
  App directory [#{app_name}] created. Next steps:
    - cd into new app
    - gem install rails
    - rails new -m <template_file>
  eos
  puts next_steps
end

def create_dir app_name
  die_unhappily("Directory \"#{app_name}\" already exists") if Dir.exists? app_name
  Dir.mkdir app_name
end

def main
  usage_and_exit_happy if ARGV.length == 0
  die_unhappily("Too many arguments. Exiting") if ARGV.length > 1
  app_name = ARGV[0]
  create_dir app_name
  create_rvmrc app_name
  create_powrc app_name
  print_next_steps app_name
end

main

#!/usr/bin/env ruby

require 'base64'
require 'nokogiri'
require 'open3'
require 'optparse'
require 'pathname'
require 'time'

# Helpers
def consider_control_center
  check_file = Pathname.new('/tmp/calm-notifications-control-center-check.txt')
  return if check_file.exist?

  check_file.write(
    'The first time after boot that DND is programatically turned off, ' \
    'the interface doesn’t update until Control Center is restarted. ' \
    'This file exists to confirm that has been done.'
  )
  system('/usr/bin/killall', 'ControlCenter')
end

def grab_dnd_xml
  prefs_file = Pathname.new(ENV['HOME']).join('Library/Preferences/com.apple.ncprefs.plist')

  outer_xml = Nokogiri::XML(Open3.capture2(
    '/usr/bin/plutil', '-extract', 'dnd_prefs', 'xml1', prefs_file.to_path, '-o', '-'
  ).first)

  Nokogiri::XML(Open3.capture2(
    '/usr/bin/plutil', '-convert', 'xml1', '-', '-o', '-',
    stdin_data: Base64.decode64(outer_xml.at('data').text)
  ).first)
end

def commit_status(xml)
  binary_plist = Open3.capture2(
    '/usr/bin/plutil', '-convert', 'binary1', '-', '-o', '-',
    stdin_data: xml.to_s
  ).first

  hex_data = binary_plist.unpack1('H*')

  system('/usr/bin/defaults', 'write', 'com.apple.ncprefs.plist', 'dnd_prefs', '-data', hex_data)
  system('/usr/bin/killall', 'usernoted')
end

def dnd_on?(xml)
  xml.xpath('boolean(//key[text()="userPref"]/following-sibling::dict/key[text()="enabled"])')
end

def status_dnd
  xml = grab_dnd_xml
  puts(dnd_on?(xml) ? 'on' : 'off')
end

def enable_dnd
  xml = grab_dnd_xml
  return if dnd_on?(xml)

  user_pref = <<~XML.freeze
    <key>userPref</key>
    <dict>
      <key>date</key>
      <date>#{Time.now.utc.iso8601}</date>
      <key>enabled</key>
      <true/>
      <key>reason</key>
      <integer>1</integer>
    </dict>
  XML

  xml.at('dict').add_child(user_pref)

  commit_status(xml)
end

def disable_dnd
  xml = grab_dnd_xml
  return unless dnd_on?(xml)

  xml.xpath('//key[text()="userPref"]/following-sibling::dict').remove
  xml.xpath('//key[text()="userPref"]').remove

  commit_status(xml)
  consider_control_center
end

def toggle_dnd
  xml = grab_dnd_xml
  dnd_on?(xml) ? disable_dnd : enable_dnd
end

# Options
ARGV.push('--help') if ARGV.empty?

OptionParser.new do |parser|
  parser.banner = <<~BANNER
    Enable, disable, toggle, and show status of Do Not Disturb on macOS Big Sur.

    Settings may take a few seconds to visually come into effect.

    Usage:
      #{Pathname.new($PROGRAM_NAME).basename} [-h|--help] <status|on|off|toggle>
  BANNER
end.parse!

# Main
if Open3.capture2('/usr/bin/sw_vers', '-productVersion').first.split('.').first != '11'
  abort 'This script only works on macOS Big Sur. For Monterey and above, use the Shortcuts command line tool.'
end

case ARGV[0]
when 'status'
  status_dnd
when 'on'
  enable_dnd
when 'off'
  disable_dnd
when 'toggle'
  toggle_dnd
else
  warn('Invalid command! Try "--help"')
end

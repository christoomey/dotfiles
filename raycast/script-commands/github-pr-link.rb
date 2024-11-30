#!/usr/bin/env ruby

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title github-pr-link
# @raycast.mode silent

# Optional parameters:
# @raycast.icon icons/github-logo.png

# Documentation:
# @raycast.description Copy GitHub PR Markdown Link
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

class GithubMarkdownUrl
  def self.run
    new.send(:run)
  end

  private

  def run
    system "printf '[#{title}](#{url})' | pbcopy"
    puts "Copied to clipboard -- [#{title}](#{url})"
  end

  def url
    raw_url = `chrome-cli info | grep Url:`.chomp
    raw_url.split(" ").last
  end

  def title
    regex = /Title: (?<title>.*?) by .* · Pull Request (?<pr_number>#[0-9]+) ·/
    raw_title = `chrome-cli info | grep Title:`.chomp

    if match = raw_title.match(regex)
      title = match[:title]
      pr_number = match[:pr_number]

      "#{title} (#{pr_number})"
    else
      puts "No match found"
      exit 1
    end
  end
end

GithubMarkdownUrl.run


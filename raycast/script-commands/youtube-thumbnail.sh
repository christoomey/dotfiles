#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Youtube Thumbnail
# @raycast.mode silent

# Optional parameters:
# @raycast.icon icons/youtube.webp
# @raycast.argument1 { "type": "text", "placeholder": "Placeholder" }

# Documentation:
# @raycast.description Copy the default thumbnail for the provided youtube video
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

url="$1"
video_id=$(echo "$url" | sed -n 's/.*v=\([^&]*\).*/\1/p')
thumbnail_url="https://img.youtube.com/vi/$video_id/maxresdefault.jpg"

curl -s "$thumbnail_url" -o /tmp/image.jpg \
  && osascript -e 'set the clipboard to (read (POSIX file "/tmp/image.jpg") as JPEG picture)'

echo "Thumbnail copied to the clipboard"

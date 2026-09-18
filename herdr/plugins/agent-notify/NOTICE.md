# icon

`icon.svg` is the source for the Herdr Agents notification icon (a shell caret
and cursor). `icon.png` is rendered from it:

    rsvg-convert -w 1024 -h 1024 icon.svg -o icon.png

then `herdr-notify install` rebuilds the sender bundle's AppIcon.icns.

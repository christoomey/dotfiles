#!/bin/bash

if pgrep mbsync; then
  echo "Another instance of mbsync is running. Killing it now."
  pkill mbsync
fi

/usr/local/bin/mbsync -q -a

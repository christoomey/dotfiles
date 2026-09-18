# agent-notify

Desktop notifications for herdr agent panes. Replaces
`yankewei/herdr-focus-notify`.

```
✓ august / order-times-data-model     <- workspace, worktrees nested under their repo
Effective period data model upgrade   <- the session name
Finished · workspace 5                <- status, and the number to jump to
```

`×` for blocked, `✓` for done, matching `status_indicators = "symbols"` in
`config.toml`. Clicking raises Ghostty and focuses the pane and its tab.

The logic lives in `bin/herdr-notify` (`herdr-notify --help`); the symlink here
is so herdr can resolve the command whether it looks in the plugin root or on
PATH.

## Why the old one was useless

- every card said "claude finished" with a fixed body, and the event's
  `workspace_id` and pane title were thrown away
- `alerter` defaults to `--sender com.apple.Terminal`, so notifications were
  delivered *as Terminal.app*, which ships `NSUserNotificationAlertStyle =
  alert` -> they stuck on screen until dismissed. Ghostty's notification
  settings were never involved.
- it passed `--app-icon`, which macOS 14 ignores; the icon was always
  Terminal's

So this one delivers as its own bundle, `com.ctoomey.herdr-agents`, whose
Info.plist declares `banner`. That bundle is also what gives the notification
its icon and its own row in System Settings > Notifications.

## Activation

```sh
herdr-notify install                    # build + register the sender bundle
herdr plugin uninstall herdr-focus-notify # if still installed; it would duplicate
herdr plugin link ~/code/dotfiles/herdr/plugins/agent-notify
herdr-notify doctor                     # expect: alert style Banners
herdr-notify test                       # one real card, 10s timeout
```

`herdr-notify preview [PANE]` prints what a pane's card would say without
sending anything.

## Requirements

`alerter` (`brew install vjeantet/tap/alerter`) and `jq`, both already
installed. `[ui.toast] delivery = "off"` in `config.toml` stays off, otherwise
herdr's own toast duplicates this.

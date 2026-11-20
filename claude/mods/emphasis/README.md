# emphasis

A Claude Code mod that draws `[[Done]]`, `[[Failed]]`, `[[Needs you]]` and
friends in Claude's replies as colored pills with Nerd Font icons, adds icons
to lead labels like `**Fix:**` and `## Plan`, and colors `✓`, `✗` and `->` on
pill lines.

Ported from [pi-emphasis](https://github.com/johnlindquist/pi-emphasis) by
John Lindquist (MIT), a Pi extension. The pill vocabulary, colors, synonyms and
reply-style rules follow that project; the rendering is rewritten for Claude
Code's `ui.render` hook.

## How it works

- `ui.render` on `AssistantMessage`: lines with a pill are drawn piece by
  piece (pills, bold, italic, strike, code, clickable links); every other
  line, plus tables and code fences, goes through the normal markdown
  renderer. Stored messages are untouched.
- `prompt.compose`: adds a short reply-style section (terminal only) telling
  Claude which pill words to use.
- `/emphasis on|off`: toggle for the current session.

Needs a terminal with Nerd Font symbols (Ghostty bundles them).

## Install / opt out

```sh
claude plugin marketplace add ~/code/dotfiles/claude/mods
claude plugin install emphasis@dotfiles-mods
claude plugin disable emphasis   # opt out
claude plugin enable emphasis
```

The marketplace is a local folder, so edits here take effect after
`/reload-plugins`.

## Limits

- Pills inside tables stay as typed.
- `✓`, `✗`, `->` are colored only on lines that also hold a pill.

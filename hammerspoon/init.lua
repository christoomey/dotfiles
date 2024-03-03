local fn = require("fn")
local hyper = require("hyper")

hyper.bindAll({
  ["\\"] = fn.openRaycastExtension("raycast/clipboard-history/clipboard-history"),
  ["]"] = fn.openRaycastExtension("raycast/snippets/search-snippets"),
  ["1"] = { base = fn.openApp("1Password"), withShift = fn.sendKeys({ "ctrl", "alt", "cmd" }, "1") },
  ["space"] = fn.openApp("Finder"),
  f = fn.openApp("Slack"),
  h = {
    fn.openApp("Hammerspoon"),
    { ["shift"] = reloadHammerspoonConfig },
  },
  i = fn.openApp("Google Chrome"),
  j = fn.openApp("IntelliJ IDEA"),
  l = {
    fn.openApp("Linear"),
    { ["shift"] = fn.openRaycastExtension("thomaslombart/linear/assigned-issues") },
  },
  n = {
    fn.openApp("Notes"),
    { ["shift"] = fn.openRaycastExtension("tumtum/apple-notes/index") },
  },
  o = {
    fn.openApp("Obsidian"),
    { ["shift"] = fn.openRaycastExtension("KevinBatdorf/obsidian/searchNoteCommand") },
  },
  p = fn.openRaycastExtension("raycast/raycast-ai/ai-chat"),
  s = fn.openApp("iTerm"),
  r = {
    fn.openApp("Spotify"),
    { ["shift"] = fn.openRaycastExtension("mattisssa/spotify-player/yourLibrary") },
    { ["shift,ctrl"] = fn.runShortcut("Spotify Random") },
  },
  t = {
    fn.openRaycastScriptCommand("open-trello"),
    { ["shift"] = fn.openRaycastExtension("ChrisChinchilla/trello/searchBoards") }
  },
  u = {
    fn.openApp("Things3"),
    { ["shift"] = fn.sendKeys({ "cmd", "shift", "ctrl" }, "u") },
  },
})

-- TODO things, 1password, etc even when the app is closed
-- https://evantravers.com/articles/2020/06/08/hammerspoon-a-better-better-hyper-key/#even-when-the-app-is-closed

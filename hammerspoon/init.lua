local fn = require("fn")
local hyper = require("hyper")

hyper.bindAll({
  ["\\"] = fn.openRaycastExtension("raycast/clipboard-history/clipboard-history"),
  ["]"] = fn.openRaycastExtension("raycast/snippets/search-snippets"),
  ["1"] = { base = fn.openApp("1Password"), withShift = fn.sendKeys({ "ctrl", "alt", "cmd" }, "1") },
  ["space"] = fn.openApp("Finder"),
  f = fn.openApp("Slack"),
  h = { base = fn.openApp("Hammerspoon"), withShift = reloadHammerspoonConfig },
  i = fn.openApp("Google Chrome"),
  j = fn.openApp("IntelliJ IDEA"),
  l = {
    base = fn.openApp("Linear"),
    withShift = fn.openRaycastExtension("thomaslombart/linear/assigned-issues"),
  },
  n = {
    base = fn.openApp("Notes"),
    withShift = fn.openRaycastExtension("tumtum/apple-notes/index"),
  },
  o = {
    base = fn.openApp("Obsidian"),
    withShift = fn.openRaycastExtension("KevinBatdorf/obsidian/searchNoteCommand"),
  },
  p = fn.openRaycastExtension("raycast/raycast-ai/ai-chat"),
  s = fn.openApp("iTerm"),
  r = {
    base = fn.openApp("Spotify"),
    withShift = fn.openRaycastExtension("mattisssa/spotify-player/yourLibrary"),
  },
  t = {
    base = fn.openRaycastScriptCommand("open-trello"),
    withShift = fn.openRaycastExtension("ChrisChinchilla/trello/searchBoards")
  },
  u = {
    base = fn.openApp("Things3"),
    withShift = fn.sendKeys({ "cmd", "shift", "ctrl" }, "u"),
  },
})

-- TODO things, 1password, etc even when the app is closed
-- https://evantravers.com/articles/2020/06/08/hammerspoon-a-better-better-hyper-key/#even-when-the-app-is-closed

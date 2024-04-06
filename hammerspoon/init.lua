local fn = require("fn")
local hyper = require("hyper")

local windowModal = hs.hotkey.modal.new({}, nil)

hyper.bindAll({
  ["\\"] = fn.openRaycastExtension("raycast/clipboard-history/clipboard-history"),
  ["]"] = fn.openRaycastExtension("raycast/snippets/search-snippets"),
  ["0"] = function() windowModal:enter() end,
  ["1"] = {
    fn.openApp("1Password"),
    { ["shift"] = fn.sendKeys({ "shift", "ctrl", "cmd" }, "1") },
  },
  ["."] = {
    fn.runShortcut("Start Session"),
    { ["shift"] = fn.openRaycastExtension("jameslyons/session/session-finish") },
  },
  ["space"] = fn.openApp("Finder"),
  a = fn.openRaycastExtension("raycast/floating-notes/toggle-floating-notes-window"),
  b = fn.sendKeys({ "shift", "ctrl", "cmd" }, "y"), -- Bartender quick search
  c = {
    fn.openTab("calendar.google.com/calendar/u/0/r"),
    { ["shift"] = fn.openTab("calendar.google.com/calendar/u/1/r") },
  },
  e = {
    fn.sendKeys({ "shift", "ctrl", "cmd" }, "e"), -- Expose
    { ["shift"] = fn.sendKeys({ "shift", "ctrl", "cmd" }, "d") }, -- Desktop
  },
  f = fn.openApp("Slack"),
  g = {
    fn.openTab("github.com"),
    { ["shift"] = fn.openRaycastExtension("raycast/github/search-repositories") },
  },
  h = {
    fn.openApp("Hammerspoon"),
    { ["shift"] = fn.reloadHammerspoonConfig },
  },
  i = {
    fn.openApp("Google Chrome"),
    { ["shift"] = fn.openRaycastExtension("Codely/google-chrome/search-bookmarks") },
  },
  j = fn.openApp("IntelliJ IDEA"),
  k = fn.openApp("Slack"),
  l = {
    fn.openApp("Linear"),
    { ["shift"] = fn.openRaycastExtension("thomaslombart/linear/assigned-issues") },
  },
  m = {
    fn.openTab("mail.google.com/mail/u/0"),
    { ["shift"] = fn.openTab("mail.google.com/mail/u/1") },
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
  },
  t = {
    fn.openTab("trello.com"),
    { ["shift"] = fn.openRaycastExtension("ChrisChinchilla/trello/searchBoards") }
  },
  u = {
    fn.openApp("Things3"),
    { ["shift"] = fn.sendKeys({ "cmd", "shift", "ctrl" }, "u") },
  },
  w = { -- Notion ("w" for "wiki")
    fn.openTab("notion.so"),
    { ["shift"] = fn.openRaycastExtension("notion/notion/search-page") }
  },
  z = fn.openApp("zoom.us")
})

local windowModalTimer = nil
local function resetTimer()
  if type(windowModalTimer) == 'table' then windowModalTimer:stop() end
  windowModalTimer = hs.timer.doAfter(5, function() windowModal:exit() end)
  print("5 more seconds on the clock")
end

local function bind(key, action)
  windowModal:bind({}, key, nil, function()
    action()
    resetTimer()
  end)
end

bind("0", function() windowModal:exit() end)
bind("escape", function() windowModal:exit() end)

bind("tab", fn.windowCommand("next-display"))

bind("a", fn.windowCommand("almost-maximize"))

bind("q", fn.windowCommand("first-two-thirds"))
bind("p", fn.windowCommand("last-third"))

bind("h", fn.windowCommand("maximize-height"))
bind("w", fn.windowCommand("maximize-width"))

bind("f", fn.windowCommand("maximize"))
bind("m", fn.windowCommand("center-three-fourths"))

bind("h", fn.windowCommand("move-left"))
bind("k", fn.windowCommand("move-up"))
bind("j", fn.windowCommand("move-down"))
bind("l", fn.windowCommand("move-right"))

bind("r", fn.windowCommand("restore"))

bind("-", fn.windowCommand("make-smaller"))
bind("=", fn.windowCommand("make-larger"))


local menuItem = hs.menubar.new(true, 'window')
menuItem:setTitle("Window")

function windowModal:entered()
  hyper.exit()
  menuItem:setTitle("WINDOW")
end

function windowModal:exited()
  menuItem:setTitle("Window")
end


-- TODO things, 1password, etc even when the app is closed
-- https://evantravers.com/articles/2020/06/08/hammerspoon-a-better-better-hyper-key/#even-when-the-app-is-closed

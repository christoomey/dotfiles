local fn = require("fn")
local hyper = require("hyper")

local windowModal = hs.hotkey.modal.new({}, nil)

hyper.bindAll({
  ["\\"] = fn.openRaycastExtension("raycast/clipboard-history/clipboard-history"),
  ["]"] = fn.openRaycastExtension("raycast/snippets/search-snippets"),
  ["0"] = function()
    windowModal:enter()
  end,
  ["1"] = {
    fn.openApp("1Password"),
    { ["shift"] = fn.sendKeys({ "ctrl", "alt", "cmd" }, "1") },
  },
  ["."] = {
    fn.runShortcut("Start Session"),
    { ["shift"] = fn.openRaycastExtension("jameslyons/session/session-finish") },
  },
  ["space"] = fn.openApp("Finder"),
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
    { ["shift"] = reloadHammerspoonConfig },
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
    { ["shift,ctrl"] = fn.runShortcut("Spotify Random") },
  },
  t = {
    fn.openTab("trello.com"),
    { ["shift"] = fn.openRaycastExtension("ChrisChinchilla/trello/searchBoards") }
  },
  u = {
    fn.openApp("Things3"),
    { ["shift"] = fn.sendKeys({ "cmd", "shift", "ctrl" }, "u") },
    { ["shift,ctrl"] = fn.openRaycastExtension("loris/things/show-today-list") },
  },
  w = { -- Notion ("w" for "wiki")
    fn.openTab("notion.so"),
    { ["shift"] = fn.openRaycastExtension("notion/notion/search-page") }
  },
})

windowModal:bind({}, "0", nil, function() windowModal:exit() end)
windowModal:bind({}, "escape", nil, function() windowModal:exit() end)

windowModal:bind({}, "a", nil, fn.openRaycastExtension("raycast/window-management/almost-maximize"))

windowModal:bind({}, "q", nil, fn.openRaycastExtension("raycast/window-management/first-two-thirds"))
windowModal:bind({}, "p", nil, fn.openRaycastExtension("raycast/window-management/last-third"))

windowModal:bind({}, "h", nil, fn.openRaycastExtension("raycast/window-management/maximize-height"))
windowModal:bind({}, "w", nil, fn.openRaycastExtension("raycast/window-management/maximize-width"))

windowModal:bind({}, "f", nil, fn.openRaycastExtension("raycast/window-management/maximize"))
windowModal:bind({}, "m", nil, fn.openRaycastExtension("raycast/window-management/center-three-fourths"))

windowModal:bind({}, "h", nil, fn.openRaycastExtension("raycast/window-management/move-left"))
windowModal:bind({}, "k", nil, fn.openRaycastExtension("raycast/window-management/move-up"))
windowModal:bind({}, "j", nil, fn.openRaycastExtension("raycast/window-management/move-down"))
windowModal:bind({}, "l", nil, fn.openRaycastExtension("raycast/window-management/move-right"))

windowModal:bind({}, "r", nil, fn.openRaycastExtension("raycast/window-management/restore"))

windowModal:bind({}, "-", nil, fn.openRaycastExtension("raycast/window-management/make-smaller"))
windowModal:bind({}, "=", nil, fn.openRaycastExtension("raycast/window-management/make-larger"))


local menuItem = hs.menubar.new(true, 'window')
menuItem:setTitle("Window")

function windowModal:entered()
  hyper.exit()
  menuItem:setTitle("WINDOW")
  hs.timer.doAfter(10, function() windowModal:exit() end) -- make this count from last action
end

function windowModal:exited()
  menuItem:setTitle("Window")
end

-- TODO things, 1password, etc even when the app is closed
-- https://evantravers.com/articles/2020/06/08/hammerspoon-a-better-better-hyper-key/#even-when-the-app-is-closed

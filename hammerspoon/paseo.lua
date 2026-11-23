-- Opens Paseo's New thread screen from Leader Key via
-- hammerspoon://paseo-new-thread. Paseo has no deep link to plugin screens, so
-- this focuses the app and presses the plugin's Cmd+Option+N shortcut once any
-- held modifiers are released.

local M = {}

local POLL_SECONDS = 0.05
local MAX_POLLS = 60

local function paseoIsFront()
  local app = hs.application.frontmostApplication()
  return app ~= nil and app:name() == "Paseo"
end

local function modifiersReleased()
  return next(hs.eventtap.checkKeyboardModifiers()) == nil
end

local function ready()
  return paseoIsFront() and modifiersReleased()
end

function M.newThread()
  hs.application.launchOrFocus("Paseo")
  local polls = 0
  M.timer = hs.timer.waitUntil(function()
    polls = polls + 1
    return ready() or polls >= MAX_POLLS
  end, function()
    if ready() then
      hs.eventtap.keyStroke({ "cmd", "alt" }, "n", 0)
    end
  end, POLL_SECONDS)
end

hs.urlevent.bind("paseo-new-thread", M.newThread)

return M

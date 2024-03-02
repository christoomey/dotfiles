hs.hotkey.bindSpec({ { "ctrl", "cmd", "shift" }, "h" },
  function()
    hs.notify.show("Hello World!", "Welcome to Hammerspoon", "")
  end
)

-- A variable for the Hyper Mode
local k = hs.hotkey.modal.new({}, 'F19')

-- All of the keys, from here:
-- https://github.com/Hammerspoon/hammerspoon/blob/f3446073f3e58bba0539ff8b2017a65b446954f7/extensions/keycodes/internal.m
-- except with ' instead of " (not sure why but it didn't work otherwise)
-- and the function keys greater than F12 removed.
local keys = {
  "a",
  "b",
  "c",
  "d",
  "e",
  "f",
  "g",
  "h",
  "i",
  "j",
  "k",
  "l",
  "m",
  "n",
  "o",
  "p",
  "q",
  "r",
  "s",
  "t",
  "u",
  "v",
  "w",
  "x",
  "y",
  "z",
  "0",
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  "`",
  "=",
  "-",
  "]",
  "[",
  "\'",
  ";",
  "\\",
  ",",
  "/",
  ".",
  "§",
  "f1",
  "f2",
  "f3",
  "f4",
  "f5",
  "f6",
  "f7",
  "f8",
  "f9",
  "f10",
  "f11",
  "f12",
  "pad.",
  "pad*",
  "pad+",
  "pad/",
  "pad-",
  "pad=",
  "pad0",
  "pad1",
  "pad2",
  "pad3",
  "pad4",
  "pad5",
  "pad6",
  "pad7",
  "pad8",
  "pad9",
  "padclear",
  "padenter",
  "return",
  "tab",
  "space",
  "delete",
  "help",
  "home",
  "pageup",
  "forwarddelete",
  "end",
  "pagedown",
  "left",
  "right",
  "down",
  "up"
}

-- local printIsdown = function(b) return b and 'down' or 'up' end

-- sends a key event with all modifiers
-- bool -> string -> void -> side effect
local hyper = function(isdown)
  return function(key)
    return function()
      k.triggered = true
      local event = hs.eventtap.event.newKeyEvent(
        {'cmd', 'alt', 'ctrl'},
        key,
        isdown
      )
      event:post()
      k:exit()
    end
  end
end

local hyperDown = hyper(true)
local hyperUp = hyper(false)

-- actually bind a key
local hyperBind = function(key)
  k:bind('', key, msg, hyperDown(key), hyperUp(key), nil)
end

-- bind all the keys in the huge keys table
for index, key in pairs(keys) do hyperBind(key) end

-- Enter Hyper Mode when F19 (Hyper/Capslock) is pressed
local pressedCmd = function()
  k.triggered = false
  k:enter()
end

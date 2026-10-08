-- Multi-display window placement, triggered from Leader Key via
-- hammerspoon://desk-middle and hammerspoon://laptop-full.

local M = {}

local function builtinScreen()
  for _, s in ipairs(hs.screen.allScreens()) do
    if s:name():find("Built%-in") then return s end
  end
end

-- The "big" screen: the largest non-built-in display, falling back to primary.
local function bigScreen()
  local builtin = builtinScreen()
  local best
  for _, s in ipairs(hs.screen.allScreens()) do
    if s ~= builtin then
      local f = s:fullFrame()
      if not best or f.w * f.h > best:fullFrame().w * best:fullFrame().h then best = s end
    end
  end
  return best or hs.screen.primary()
end

local function withWindow(fn)
  local win = hs.window.focusedWindow()
  if win then fn(win) end
end

-- Center three-fourths of the big screen's width, full height (same geometry
-- as Raycast's center-three-fourths). Moves the window there first if needed.
function M.deskMiddle()
  withWindow(function(win)
    local target = bigScreen()
    if win:screen() ~= target then win:moveToScreen(target, false, true, 0) end
    local f = target:frame()
    win:setFrame({ x = f.x + f.w / 8, y = f.y, w = f.w * 3 / 4, h = f.h }, 0)
  end)
end

-- Maximize on the laptop screen, moving there first if needed.
function M.laptopFull()
  withWindow(function(win)
    local target = builtinScreen() or hs.screen.primary()
    if win:screen() ~= target then win:moveToScreen(target, false, true, 0) end
    win:setFrame(target:frame(), 0)
  end)
end

hs.urlevent.bind("desk-middle", M.deskMiddle)
hs.urlevent.bind("laptop-full", M.laptopFull)

return M

-- A variable for the Hyper Mode
local k = hs.hotkey.modal.new({}, 'F19')

local menuItem = hs.menubar.new(true, 'hyperKey')
menuItem:setTitle("Hyper")

function k:entered()
  menuItem:setTitle("HYPER")
  hs.timer.doAfter(1, function() k:exit() end)
end

function k:exited()
  menuItem:setTitle("Hyper")
end

function runAndExit(action)
  return function()
    k:exit()
    action()
  end
end

local bind = function(mods, key, action)
  k:bind(mods, key, runAndExit(action))
end

return {
  bind = bind,
}

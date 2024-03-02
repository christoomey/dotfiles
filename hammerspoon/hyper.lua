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

local function runAndExit(action)
  return function()
    k:exit()
    action()
  end
end

local function bind(mods, key, action)
  k:bind(mods, key, runAndExit(action))
end

local function bindAll(keyMap)
  for key, action in pairs(keyMap) do
    if type(action) == 'table' then
      print('var is a table')
    elseif type(action) == 'function' then
      bind({}, key, action)
    else
      print('config value for ' .. key .. ' is neither a table nor a function')
    end

  end
end

return {
  bind = bind,
  bindAll = bindAll,
}

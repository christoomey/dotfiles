local function openApp(name)
  return function()
    hs.application.launchOrFocus(name)
    if name == 'Finder' then
      hs.appfinder.appFromName(name):activate()
    end
  end
end

local function execute(command)
  print("Debug: running command `" .. command .. "`")
  local _, exitedSuccessfully, _, _ = hs.execute(command)
  if not exitedSuccessfully then
    hs.notify.show("Command exited non-zero", "Command: `" .. command .."`")
  end
end

-- https://manual.raycast.com/deeplinks#e460b1f1c034468db3fbf9f028d8f01c
-- Get the deep link from Cmd-k in the menu
local function openRaycastExtension(extensionDeepLinkPath)
  return function()
    execute("open -g " .. "raycast://extensions/" .. extensionDeepLinkPath)
  end
end

local function windowCommand(commandName)
  return function()
    print("Debug: running function 'windowCommand' with: `" .. commandName .. "`")
    openRaycastExtension("raycast/window-management/" .. commandName)()
  end
end

local function openRaycastScriptCommand(scriptCommand)
  return function()
    execute("open " .. "raycast://script-commands/" .. scriptCommand)
  end
end

local function reloadHammerspoonConfig()
  hs.reload()
  hs.notify.show("Hammerspoon", "Config reloaded", "")
end

local function sendKeys(mods, key)
  return function()
    hs.eventtap.keyStroke(mods, key)
  end
end

local function runShortcut(shortcutName)
  return function()
    execute('shortcuts run "' .. shortcutName .. '"')
  end
end

local function openTab(tabUrl)
  return function()
    execute("~/bin/open-tab '" .. "https://" .. tabUrl .. "'")
  end
end

return {
  openApp = openApp,
  openTab = openTab,
  openRaycastExtension = openRaycastExtension,
  openRaycastScriptCommand = openRaycastScriptCommand,
  sendKeys = sendKeys,
  reloadHammerspoonConfig = reloadHammerspoonConfig,
  runShortcut = runShortcut,
  windowCommand = windowCommand,
}

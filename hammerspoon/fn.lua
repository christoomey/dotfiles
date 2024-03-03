function openApp(name)
  return function()
    hs.application.launchOrFocus(name)
    if name == 'Finder' then
      hs.appfinder.appFromName(name):activate()
    end
  end
end

-- https://manual.raycast.com/deeplinks#e460b1f1c034468db3fbf9f028d8f01c
-- Get the deep link from Cmd-k in the menu
function openRaycastExtension(extensionDeepLinkPath)
  return function()
    hs.execute("open " .. "raycast://extensions/" .. extensionDeepLinkPath)
  end
end

function openRaycastScriptCommand(scriptCommand)
  return function()
    hs.execute("open " .. "raycast://script-commands/" .. scriptCommand)
  end
end

function reloadHammerspoonConfig()
  hs.reload()
  hs.notify.show("Hammerspoon", "Config reloaded", "")
end

function sendKeys(mods, key)
  return function()
    hs.eventtap.keyStroke(mods, key)
  end
end

function runShortcut(shortcutName)
  return function()
    hs.execute('shortcuts run "' .. shortcutName .. '"')
  end
end

return {
  openApp = openApp,
  openRaycastExtension = openRaycastExtension,
  openRaycastScriptCommand = openRaycastScriptCommand,
  sendKeys = sendKeys,
  reloadHammerspoonConfig = reloadHammerspoonConfig,
  runShortcut = runShortcut
}

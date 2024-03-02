local hyper = require("hyper")

function openAppFn(name)
  return function()
    hs.application.launchOrFocus(name)
    if name == 'Finder' then
      hs.appfinder.appFromName(name):activate()
    end
  end
end

-- https://manual.raycast.com/deeplinks#e460b1f1c034468db3fbf9f028d8f01c
-- Get the deep link from Cmd-k in the menu
function openRaycastCommandFn(author, extenstionName, commandName)
  return function()
    local commandUrl = "raycast://extensions/" .. author .. "/" .. extenstionName .. "/" .. commandName
    print(commandUrl)
    hs.execute("open " .. commandUrl)
  end
end

function reloadHammerspoonConfig()
  hs.reload()
  hs.notify.show("Hammerspoon", "Config reloaded", "")
end


hyper.bind({}, "\\", openRaycastCommandFn("raycast", "clipboard-history", "clipboard-history"))

hyper.bind({}, "h", openAppFn("Hammerspoon"))
hyper.bind({"shift"}, "h", reloadHammerspoonConfig)

hyper.bind({}, "i", openAppFn("Google Chrome"))

hyper.bind({}, "j", openAppFn("IntelliJ IDEA"))

hyper.bind({}, "l", openAppFn("Linear"))
hyper.bind({"shift"}, "l", openRaycastCommandFn("thomaslombart", "linear", "assigned-issues"))

hyper.bind({}, "n", openAppFn("Notes"))
hyper.bind({"shift"}, "n", openRaycastCommandFn("tumtum", "apple-notes", "index"))

hyper.bind({}, "s", openAppFn("iTerm"))

hyper.bind({}, "r", openAppFn("Spotify"))

-- TODO -- Use a table (key value style) for config
-- TODO -- any non-mapped key should exit and send through

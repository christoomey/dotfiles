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
function openRaycastCommandFn(extensionDeepLinkPath)
  return function()
    hs.execute("open " .. "raycast://extensions/" .. extensionDeepLinkPath)
  end
end

function reloadHammerspoonConfig()
  hs.reload()
  hs.notify.show("Hammerspoon", "Config reloaded", "")
end

hyper.bindAll({
  ["\\"] = openRaycastCommandFn("raycast/clipboard-history/clipboard-history"),
  i = openAppFn("Google Chrome"),
  j = openAppFn("IntelliJ IDEA"),
  p = openRaycastCommandFn("raycast/raycast-ai/ai-chat"),
  s = openAppFn("iTerm"),
  r = openAppFn("Spotify"),
})


hyper.bind({}, "h", openAppFn("Hammerspoon"))
hyper.bind({"shift"}, "h", reloadHammerspoonConfig)

hyper.bind({}, "l", openAppFn("Linear"))
hyper.bind({"shift"}, "l", openRaycastCommandFn("thomaslombart/linear/assigned-issues"))

hyper.bind({}, "n", openAppFn("Notes"))
hyper.bind({"shift"}, "n", openRaycastCommandFn("tumtum/apple-notes/index"))


-- TODO -- Use a table (key value style) for config
-- TODO -- any non-mapped key should exit and send through

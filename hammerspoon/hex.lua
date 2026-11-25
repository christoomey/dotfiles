-- Filterable list of recent Hex dictations via hammerspoon://hex-history.
-- Choosing one pastes it into the window that was focused before the list.

local M = {}

local HEX = "/Applications/Hex.app/Contents/MacOS/hex"
local SINCE = "2h"
local LIMIT = 50

local function relativeTime(timestampMs)
  local seconds = os.time() - math.floor(timestampMs / 1000)
  if seconds < 3600 then
    return math.max(1, math.floor(seconds / 60)) .. "m ago"
  elseif seconds < 86400 then
    return math.floor(seconds / 3600) .. "h ago"
  else
    return math.floor(seconds / 86400) .. "d ago"
  end
end

local function loadEntries()
  local output, ok = hs.execute(string.format("%q history --since %s --limit %d --json", HEX, SINCE, LIMIT))
  local decoded = ok and hs.json.decode(output)
  if not decoded then
    hs.alert.show("Hex history unavailable (is Hex running?)")
    return nil
  end
  local choices = {}
  for _, entry in ipairs(decoded.entries or {}) do
    if entry.final_text and entry.final_text ~= "" then
      table.insert(choices, {
        text = entry.final_text,
        subText = (entry.application or "?") .. " · " .. relativeTime(entry.timestamp_ms),
      })
    end
  end
  return choices
end

local function paste(text, window)
  local previous = hs.pasteboard.getContents()
  hs.pasteboard.setContents(text)
  if window then
    window:focus()
  end
  hs.timer.doAfter(0.15, function()
    hs.eventtap.keyStroke({ "cmd" }, "v", 0)
    hs.timer.doAfter(0.5, function()
      if previous then
        hs.pasteboard.setContents(previous)
      end
    end)
  end)
end

function M.show()
  local choices = loadEntries()
  if not choices then
    return
  end
  local window = hs.window.focusedWindow()
  M.chooser = hs.chooser.new(function(choice)
    if choice then
      paste(choice.text, window)
    end
  end)
  M.chooser:placeholderText("Recent Hex dictations (last " .. SINCE .. ")")
  M.chooser:searchSubText(true)
  M.chooser:choices(choices)
  M.chooser:show()
end

hs.urlevent.bind("hex-history", M.show)

return M

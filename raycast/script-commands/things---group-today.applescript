#!/usr/bin/osascript

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Things - Group Today
# @raycast.mode silent

# Optional parameters:
# @raycast.icon icons/things.webp

# Documentation:
# @raycast.description Toggle grouping by project in the Today view
# @raycast.author Chris Toomey
# @raycast.authorURL https://ctoomey.com

tell application "System Events"
	-- Activate Things3 application
	tell application "Things3" to activate
	
	-- Delay to ensure the application is in focus
	delay 0.2
	
	-- Open Preferences window
	keystroke "," using {command down}
	
	-- Delay to allow Preferences window to open
	delay 0.2
	
	-- Interact with the General tab in Preferences
	tell process "Things3"
		-- Ensure we are in the right tab
		click button "General" of toolbar 1 of window 1
		
		-- Delay to ensure the General tab is active
		delay 0.2
		
		-- Find and toggle the checkbox
		set groupCheckbox to checkbox "Group to-dos in the Today list by project or area" of window 1
		
		-- Check if the checkbox is unchecked and then check it
		if value of groupCheckbox is 0 then
			click groupCheckbox
			-- If it is already checked, click to uncheck it (toggle behavior)
		else
			click groupCheckbox
		end if
	end tell
	
	-- Close Preferences window
	keystroke "w" using {command down}
end tell


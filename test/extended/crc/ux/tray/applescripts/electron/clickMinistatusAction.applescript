use framework "Quartz"

global KEYSTROKE_ENTER
global KEYSTROKE_TAB
global MINISTATUS_WINDOW_ID

on run {CFBundleIdentifier, windowName, action}

set KEYSTROKE_ENTER to 76
set KEYSTROKE_TAB to 48
-- set MINISTATUS_WINDOW_ID to "CodeReady Containers"
set BUTTON_START_ORDER to 1
set BUTTON_STOP_ORDER to 2
set BUTTON_DELETE_ORDER to 3

if (action = "start") then
	executeAction(CFBundleIdentifier, windowName, BUTTON_START_ORDER)
else if (action = "stop") then
	executeAction(CFBundleIdentifier, windowName, BUTTON_STOP_ORDER)
else if (action = "delete") then
	executeAction(CFBundleIdentifier, windowName, BUTTON_DELETE_ORDER)
end if

end run

on getTrayPosition(CFBundleIdentifier) 
tell application "System Events"
    tell (first application process whose bundle identifier is CFBundleIdentifier)
        tell menu bar item of menu bar 2
            set trayPosition to position
            set x to item 1 of item 1 of trayPosition as integer
            set y to item 2 of item 1 of trayPosition as integer
            return {x, y}
        end tell
    end tell
end tell
end getTrayPosition

on mouseLeftClick(tPoint)
    current application's CGPostMouseEvent(tPoint, 1, 1, 1)
    delay 0.5
    current application's CGPostMouseEvent(tPoint, 1, 1, 0)
end mouseLeftClick

on mouseRightClick(tPoint)
    current application's CGPostMouseEvent(tPoint, 1, 2, 0, 1)
    delay 0.5
    current application's CGPostMouseEvent(tPoint, 1, 2, 0, 0)
end mouseRightClick

on getPosition(position) 
    set x to item 1 of item 1 of position as integer
    set y to item 2 of item 1 of position as integer
    return {x, y}
end getPosition

on executeAction(CFBundleIdentifier, windowName, buttonOrder)
   set tPoint to getTrayPosition(CFBundleIdentifier)
   mouseLeftClick(tPoint)
   delay 2
   focusMinistatusWindow(CFBundleIdentifier, windowName)
   hitKey(KEYSTROKE_TAB, buttonOrder)
   hitKey(KEYSTROKE_ENTER, 1)
   delay 2
   closeMinistatusWindow(CFBundleIdentifier, windowName)
end executeAction

on closeMinistatusWindow(CFBundleIdentifier, windowName)
tell application "System Events" 
	tell (first application process whose bundle identifier is CFBundleIdentifier) 
		tell window windowName
			click button 1 #Default mac window button 1 is close
        end tell
	end tell
end tell
end closeMinistatusWindow

on focusMinistatusWindow(CFBundleIdentifier, windowName)
tell application "System Events"
	tell (first application process whose bundle identifier is CFBundleIdentifier)
        tell window windowName
            set windowPosition to position
        end tell
    end tell
end tell
mouseLeftClick(windowPosition)
end focusMinistatusWindow

on hitKey(keyCode, pressCount)
tell application "System Events"
    repeat pressCount times
	delay 1
        key code keyCode
    end repeat
end tell
end hitKey


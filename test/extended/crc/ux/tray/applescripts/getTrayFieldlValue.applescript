on run {bundleIdentifer, elementAXIdentifier}
set systemMenuBarId to 2
set trayMainMenuId to 1
set statusAndLogsMenuItem to "Status and Logs"
set statusAndLogsWindowId to "Status and Logs"
set statusAndLogsWindowCloseButtonId to 1
set menuClickDelay to 0.5
try
    tell application "System Events"
        tell (first application process whose bundle identifier is bundleIdentifer)
            tell menu bar item of menu bar systemMenuBarId
                click
                delay menuClickDelay
            end tell
        end tell
end tell
on error errMsg
end try
tell application "System Events"
    tell (first application process whose bundle identifier is bundleIdentifer)
        set elementValue to name of (first menu item of menu 1 of menu bar 2 whose value of attribute "AXIdentifier" is elementAXIdentifier)
        tell menu item statusAndLogsMenuItem of menu trayMainMenuId of menu bar item of menu bar systemMenuBarId
            click
            delay menuClickDelay
        end tell
        click button statusAndLogsWindowCloseButtonId of window statusAndLogsWindowId
        return elementValue
    end tell
end tell
end run

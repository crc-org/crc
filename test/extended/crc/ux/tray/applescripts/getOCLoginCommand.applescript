on run {bundleIdentifer, copyLoginUserAXIdentifier}
set systemMenuBarId to 2
set trayMainMenuId to 1
set copyOCLoginMenuItem to "Copy OC Login Command"
set copyOCLoginMenuAXIdentifier to "copy_login"
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
          tell menu item copyOCLoginMenuItem of menu trayMainMenuId of menu bar item of menu bar systemMenuBarId
            click
            delay menuClickDelay
            tell (first menu whose value of attribute "AXIdentifier" is copyOCLoginMenuAXIdentifier)
                tell (first menu item whose value of attribute "AXIdentifier" is copyLoginUserAXIdentifier)
                    click
                end tell
            end tell 
        end tell
    end tell
end tell
end run

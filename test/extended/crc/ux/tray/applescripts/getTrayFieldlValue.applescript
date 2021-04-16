on run {bundleIdentifer, elementAXIdentifier}
    tell application "System Events"
        tell (first application process whose bundle identifier is bundleIdentifer)
            click menu bar item of menu bar 2
            delay 0.5
            return name of (first menu item of menu 1 of menu bar 2 whose value of attribute "AXIdentifier" is elementAXIdentifier)
        end tell
    end tell
end run
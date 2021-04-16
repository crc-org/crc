on run {bundleIdentifer, buttonAXIdentifier}
    tell application "System Events"
        tell (first application process whose bundle identifier is bundleIdentifer)
            tell menu bar item of menu bar 2
                click
                delay 0.5
                tell (first menu item of menu 1 whose value of attribute "AXIdentifier" is buttonAXIdentifier)
                    click
                end tell
            end tell
        end tell
    end tell
end run

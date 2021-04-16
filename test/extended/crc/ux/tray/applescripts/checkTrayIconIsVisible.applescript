on run {bundleIdentifer}
    tell application "System Events" 
        tell (first application process whose bundle identifier is bundleIdentifer) 
            tell menu bar item of menu bar 2 
                try
                    name
                    return true
                on error
                    return false
                end try
            end tell 
        end tell 
    end tell
end run
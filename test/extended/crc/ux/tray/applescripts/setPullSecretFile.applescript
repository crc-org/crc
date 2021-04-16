on run {bundleIdentifer, pullSecretFile}
    tell application "System Events" 
        tell (first application process whose bundle identifier is bundleIdentifer) 
            tell window "Select Pull Secret"
                set value of text field 1 to pullSecretFile    
                tell button "OK" 
                    click
                end tell
            end tell 
        end tell 
    end tell
end run

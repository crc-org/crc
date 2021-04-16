on run {installerPath, adminPassword}
    set installer to installerPath as POSIX file
    set installerWindow to "Install CodeReady Containers"
    tell application "Finder" to open installer
    delay 1
    tell application "System Events"
        tell (first application process whose bundle identifier is "com.apple.installer") 
            # Introduction
            click button "Continue" of window installerWindow 
            delay 2
            # License
            click button "Continue" of window installerWindow 
            delay 2
        end tell
    end tell
    tell application "System Events"
        tell (first application process whose bundle identifier is "com.apple.installer")
            # Agree the License
            click button "Agree" of sheet 1 of window installerWindow
            delay 2
        end tell
    end tell
    tell application "System Events"
        tell (first application process whose bundle identifier is "com.apple.installer")
            # Select Destination
            try
            delay 2
            click button "Continue" of window installerWindow
            on error errMsg
                log errMsg
            end try
            # Installation
            click button "Install" of window installerWindow
            delay 2
            keystroke adminPassword & return
            delay 15
            # Close the installer
            click button "Close" of window installerWindow
        end tell
    end tell
end run

global KEYSTROKE_ENTER
global KEYSTROKE_TAB

on run {pullSecretFilePath, applicationPath, preset}

set KEYSTROKE_TAB to 48
set KEYSTROKE_ENTER to 76
set OPENSHIFT_STARTING_TIME to 200
set PODMAN_STARTING_TIME to 70

if (preset = "openshift") then
        set STARTING_TIME to OPENSHIFT_STARTING_TIME
        copyPullSecretContentToClipboard(pullSecretFilePath)
else
        set STARTING_TIME to PODMAN_STARTING_TIME
end if
openApp(applicationPath)
startOnBoarding()
welcomeOnBoarding()
if (preset = "openshift") then
        presetOpenshift()
        setPullSecret()
else 
        presetPodman()
end if
runSetup()
delay STARTING_TIME
startUsingCRC()

end run


on openApp(appPath)
        set newApp to appPath as POSIX file
        tell application "Finder" to open newApp
end openApp

on startOnBoarding()
        delay 2
        hitKey(KEYSTROKE_TAB, 1)
        hitKey(KEYSTROKE_ENTER, 1)
end startOnBoarding

on welcomeOnBoarding()
        delay 2
        hitKey(KEYSTROKE_TAB, 4)
        hitKey(KEYSTROKE_ENTER, 1)
end welcomeOnBoarding

on presetOpenshift()
        delay 2
        hitKey(KEYSTROKE_ENTER, 1)
end presetOpenshift

on setPullSecret()
        delay 2
        hitKey(KEYSTROKE_TAB, 5)
        delay 2
        tell application "System Events" to keystroke "v" using command down
        delay 2
        hitKey(KEYSTROKE_TAB, 3)
	hitKey(KEYSTROKE_ENTER, 1)        
end setPullSecret

on presetPodman()
        delay 2
        hitKey(KEYSTROKE_TAB, 6)
        delay 2
        hitKey(KEYSTROKE_ENTER, 1)
        delay 2
        hitKey(KEYSTROKE_TAB, 2)
        delay 2
        hitKey(KEYSTROKE_ENTER, 1)
end presetPodman

on runSetup()
	delay 2
	hitKey(KEYSTROKE_ENTER, 1)
end runSetup

on startUsingCRC()
	delay 2
        hitKey(KEYSTROKE_TAB, 2)
        hitKey(KEYSTROKE_ENTER, 1)
end startUsingCRC

on hitKey(keyCode, pressCount)
tell application "System Events"
    repeat pressCount times
	delay 1
        key code keyCode
    end repeat
end tell
end hitKey

on copyPullSecretContentToClipboard(pullSecretFile)
set terminal to getTerminal()
# Ensure clipoard is empty
executeCommand("pbcopy < /dev/null", terminal)
# Copy pull secret content to clipboard
executeCommand("cat " & pullSecretFile & " | pbcopy", terminal)
# close terminal
closeTerminal(terminal)
end copyPullSecretContentToClipboard

on getTerminal()
tell application "Terminal"
    activate
    set terminal to do script ("")
    return terminal
end tell
end getTerminal

on executeCommand(command, terminal)
tell application "Terminal"
   do script (command) in terminal
   delay 2
end tell
end executeCommand

on closeTerminal(terminal)
tell application "Terminal"
    do script ("") in terminal
    close window 1
    delay 1
    tell application "System Events" to keystroke return
end tell
end closeTerminal

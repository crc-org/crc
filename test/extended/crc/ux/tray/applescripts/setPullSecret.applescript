on run{bundleIdentifer, pullSecretFile} 
    set terminal to get_terminal()
    # Ensure clipoard is empty
    execute_command("pbcopy < /dev/null", terminal)
    # Copy pull secret content to clipboard
    execute_command("cat " & pullSecretFile & " | pbcopy", terminal)
    # close terminal
    close_terminal(terminal)
    # Set the content
    paste_pullsecret_content(bundleIdentifer, the clipboard as text)  
end run

on paste_pullsecret_content(bundleIdentifer, pullsecretContent)
tell application "System Events"
	tell (first application process whose bundle identifier is bundleIdentifer)
		tell window "Paste the Pull Secret"	
			set value of text area 1 of scroll area 1 to pullsecretContent
            click button "OK"
        end tell
	end tell
end tell
end paste_pullsecret_content

on get_terminal()
tell application "Terminal"
        set crcTerminal to do script ("")
        return crcTerminal
end tell
end get_terminal

on execute_command(command, crcTerminal)
tell application "Terminal"
   do script (command) in crcTerminal
   delay 2
end tell
end execute_command

on close_terminal(crcTerminal)
tell application "Terminal"
	do script ("") in crcTerminal
        close window 1
        delay 1
        tell application "System Events" to keystroke return
end tell
end close_terminal 

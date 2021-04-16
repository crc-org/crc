on run
try
        set whoAmICommandOutput to 2
        set terminal to get_terminal()
        execute_command("eval $(crc oc-env)", terminal)
        execute_command(the clipboard as text, terminal)
        accept_insecure_connection(terminal)
        execute_command("clear", terminal)
        execute_command("oc whoami", terminal)
        set result to command_output(whoAmICommandOutput)
        return result
on error errMsg
        return errMsg
end try
end run

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

on command_output(commandOutputLine)
tell application "Terminal"
        tell front window to set theText to contents of selected tab as text
        return (paragraph commandOutputLine of theText)
end tell
end command_output

on accept_insecure_connection(crcTerminal)
set insecureConnectionOutput to 6
set insecureConnectionQuestion to "Use insecure connections? (y/n)"
set insecure to command_output(insecureConnectionOutput)
if insecure contains insecureConnectionQuestion
        execute_command("y", crcTerminal)
end if
end accept_insecure_connection
on run {trayAppPath}
    set trayApp to trayAppPath as POSIX file
    tell application "Finder" to open trayApp
end run
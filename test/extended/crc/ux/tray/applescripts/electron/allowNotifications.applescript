use framework "Quartz"

on run 

set notificationPermissionPosition to getNotificationPermissionPosition()
moveMouse(notificationPermissionPosition)
allowNotifications()

end run


on getNotificationPermissionPosition()
tell application "System Events"
                tell group 1 of UI element 1 of scroll area 1 of window 1 of (first application process whose bundle identifier is "com.apple.notificationcenterui")
                tell static text 1
                        set notificationPosition to Position
                        return notificationPosition
                end tell
        end tell
end tell
end getNotificationPermissionPosition

on moveMouse(position)
    current application's CGPostMouseEvent(position, 1, 1, 0)
    delay 0.2
end moveMouse

on allowNotifications()
tell application "System Events"
    tell group 1 of UI element 1 of scroll area 1 of window 1 of (first application process whose bundle identifier is "com.apple.notificationcenterui")
        tell pop up button "Options"
        click
        delay 0.2
        click menu item "Allow" of menu 1
        end tell
    end tell
end tell
end allowNotifications

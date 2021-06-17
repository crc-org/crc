on run{action}
-- valid actions on notification
-- get latests notification
set action_get to "get"
-- close all notifications
set action_clear to "clear"

-- get major version for os
set os_version to system attribute "sys1"

-- notifications should be visible to perform actions 
if not is_notificationmenu_visible() then
	click_notifications(os_version)
end if

-- handle action
if action = action_get then
	set notification to get_notification(os_version)
    click_notifications(os_version)
    return notification
else if action = action_clear then
	clear_notifications(os_version)
	click_notifications(os_version)
else 
	error "not supported action"
end if

end run

-- get the menua bar controller 
-- 11.X uses controlcenter
-- 10.X systemuiserver
on get_systemmenu_controller(os_version)
if os_version = 11 then
	return "com.apple.controlcenter"
else if os_version = 10 then
	return "com.apple.systemuiserver"
else
	error "not supported os version"
end if 
end get_systemmenu_controller

on click_notifications(os_major_version)
set systemmenu_controler to get_systemmenu_controller(os_major_version)
tell application "System Events"
	tell (first application process whose bundle identifier is systemmenu_controler)
        	click menu bar item 1 of menu bar 1
                delay 0.25
    	end tell
end tell  
end click_notifications

on is_notificationmenu_visible() 
try
	tell application "System Events"
    		tell window 1 of (first application process whose bundle identifier is "com.apple.notificationcenterui")
                        name
    		end tell
	end tell
    return true
on error errMsg
	return false
end try
end is_notificationmenu_visible

on get_notification(os_version)
if os_version = 11 then
	return get_notification_11()
else if os_version = 10 then
	return get_notification_10()
else
	error "not supported os version"
end if
end get_notification

on get_notification_11()
try
	tell application "System Events"
		tell group 1 of UI element 1 of scroll area 1 of window 1 of (first application process whose bundle identifier is "com.apple.notificationcenterui")
			return value of static text 4        
			end tell
	end tell
on error
	return ""
end try 
end get_notification_11

on get_notification_10()
try 
	tell application "System Events"
		tell (first application process whose bundle identifier is "com.apple.notificationcenterui")
			tell UI Element 1 of row 3 of table 1 of scroll area 1 of window 1
				if (value of static text 4) as string is not equal to "" then
                    return value of static text 4
                else
                    return value of static text 5
                end if
			end tell
		end tell
	end tell
on error
	return ""
end try 
end get_notification_10

-- TODO Ensure only one vindow on notificationcenterui process
on clear_notifications(os_version)
if os_version = 11 then
        clear_notifications_11()
else if os_version = 10 then
        clear_notifications_10()
else
        error "not supported os version"
end if
end clear_notifications

on clear_notifications_11()
tell application "System Events"
	set allClosed to false
	repeat until allClosed
	try
        tell group 1 of UI element 1 of scroll area 1 of window 1 of (first application process whose bundle identifier is "com.apple.notificationcenterui")
                click
        end tell
	on error
 		set allClosed to true
	end try
	end repeat
end tell
end clear_notifications_11

on clear_notifications_10()
tell application "System Events"
    try 
    tell (first application process whose bundle identifier is "com.apple.notificationcenterui")
		tell (first button of UI Element 1 of row 2 of table 1 of scroll area 1 of window 1 whose value of attribute "AXDescription" is "clear")
			click
        end tell
    end tell
	on error  
	-- no notifications to clear
	end try
end tell
end clear_notifications_10


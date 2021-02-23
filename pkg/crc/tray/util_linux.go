package tray

func DisableEnableTrayAutostart(key string, value interface{}) string {
	// noop
	return ""
}

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return false, "Not supported on Linux"
}

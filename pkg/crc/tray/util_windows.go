package tray

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return false, "Not supported on Windows"
}

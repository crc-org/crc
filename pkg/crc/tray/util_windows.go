package tray

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(_ interface{}) (bool, string) {
	return false, "Not supported on Windows"
}

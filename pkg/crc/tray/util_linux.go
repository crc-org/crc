package tray

func DisableTrayAutostart() error {
	// noop
	return nil
}

func DisableDaemonAutostart() error {
	// noop
	return nil
}

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return false, "Not supported on Linux"
}

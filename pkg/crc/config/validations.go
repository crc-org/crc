package config

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) bool {
	if value.(string) == "true" || value.(string) == "false" {
		return true
	}
	return false
}

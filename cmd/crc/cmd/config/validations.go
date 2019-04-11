package config

func ValidateBool(key string, value interface{}) bool {
	if value.(string) == "true" || value.(string) == "false" {
		return true
	}
	return false
}

package os

import (
	"fmt"
	"strings"
)

// ReplaceEnv changes the value of an environment variable
// It drops the existing value and appends the new value in-place
func ReplaceEnv(variables []string, varName string, value string) []string {
	var result []string
	for _, e := range variables {
		pair := strings.Split(e, "=")
		if pair[0] != varName {
			result = append(result, e)
		} else {
			result = append(result, fmt.Sprintf("%s=%s", varName, value))
		}
	}

	return result
}

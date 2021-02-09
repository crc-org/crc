package shell

import (
	"fmt"
	"strings"
)

type Config struct {
	Prefix     string
	Delimiter  string
	Suffix     string
	PathSuffix string
}

func GetShell(userShell string) (string, error) {
	if userShell != "" {
		if !isSupportedShell(userShell) {
			return "", fmt.Errorf("'%s' is not a supported shell.\nSupported shells are %s", userShell, strings.Join(supportedShell, ", "))
		}
		return userShell, nil
	}
	return detect()
}

func isSupportedShell(userShell string) bool {
	for _, shell := range supportedShell {
		if userShell == shell {
			return true
		}
	}
	return false
}

func GenerateUsageHintWithComment(userShell, cmdLine string) string {
	return fmt.Sprintf("%s Run this command to configure your shell:\n%s %s",
		comment(userShell),
		comment(userShell),
		GenerateUsageHint(userShell, cmdLine))
}

func comment(userShell string) string {
	if userShell == "cmd" {
		return "REM"
	}
	return "#"
}

func GenerateUsageHint(userShell, cmdLine string) string {
	switch userShell {
	case "fish":
		return fmt.Sprintf("eval (%s)", cmdLine)
	case "powershell":
		return fmt.Sprintf("& %s | Invoke-Expression", cmdLine)
	case "cmd":
		return fmt.Sprintf("@FOR /f \"tokens=*\" %%i IN ('%s') DO @call %%i", cmdLine)
	default:
		return fmt.Sprintf("eval $(%s)", cmdLine)
	}
}

func GetEnvString(userShell string, envName string, envValue string) string {
	switch userShell {
	case "powershell":
		return fmt.Sprintf("$Env:%s = \"%s\"", envName, envValue)
	case "cmd":
		return fmt.Sprintf("SET %s=%s", envName, envValue)
	case "fish":
		return fmt.Sprintf("contains %s $fish_user_paths; or set -U fish_user_paths %s $fish_user_paths", envValue, envValue)
	default:
		return fmt.Sprintf("export %s=\"%s\"", envName, envValue)
	}
}

func GetPathEnvString(userShell string, prependedPath string) string {
	var pathStr string
	switch userShell {
	case "fish":
		pathStr = prependedPath
	case "powershell":
		pathStr = fmt.Sprintf("%s;$Env:PATH", prependedPath)
	case "cmd":
		pathStr = fmt.Sprintf("%s;%%PATH%%", prependedPath)
	default:
		pathStr = fmt.Sprintf("%s:$PATH", prependedPath)
	}

	return GetEnvString(userShell, "PATH", pathStr)
}

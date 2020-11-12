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

func GenerateUsageHint(userShell, cmdLine string) string {
	cmd := ""
	comment := "#"

	switch userShell {
	case "fish":
		cmd = fmt.Sprintf("eval (%s)", cmdLine)
	case "powershell":
		cmd = fmt.Sprintf("& %s | Invoke-Expression", cmdLine)
	case "cmd":
		cmd = fmt.Sprintf("\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @call %%i", cmdLine)
		comment = "REM"
	default:
		cmd = fmt.Sprintf("eval $(%s)", cmdLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell:\n%s %s", comment, comment, cmd)
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

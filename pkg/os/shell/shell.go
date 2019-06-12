package shell

import (
	"errors"
	"fmt"
	"strings"

	"github.com/code-ready/machine/libmachine/shell"
)

type ShellConfig struct {
	Prefix     string
	Delimiter  string
	Suffix     string
	PathSuffix string
}

func GetShell(userShell string) (string, error) {
	if userShell != "" {
		if !isSupportedShell(userShell) {
			return "", errors.New(fmt.Sprintf("'%s' is not a supported shell.\nSupported shells are %s.", userShell, strings.Join(supportedShell, ", ")))
		}
		return userShell, nil
	}
	return shell.Detect()
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
	case "powershell":
		cmd = fmt.Sprintf("& %s | Invoke-Expression", cmdLine)
	case "cmd":
		cmd = fmt.Sprintf("\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @call %%i", cmdLine)
		comment = "REM"
	default:
		cmd = fmt.Sprintf("eval $(%s)", cmdLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell:\n%s %s\n", comment, comment, cmd)
}

func GetPrefixSuffixDelimiterForSet(userShell string) (prefix, delimiter, suffix, pathSuffix string) {
	switch userShell {
	case "powershell":
		prefix = "$Env:"
		delimiter = " = \""
		suffix = "\"\n"
		pathSuffix = ";" + prefix + "PATH" + suffix
	case "cmd":
		prefix = "SET "
		delimiter = "="
		suffix = "\n"
		pathSuffix = ";%PATH%" + suffix
	default:
		prefix = "export "
		delimiter = "=\""
		suffix = "\"\n"
		pathSuffix = ":$PATH" + suffix
	}
	return
}

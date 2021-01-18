/*
Copyright (C) 2019 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testsuite

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/clicumber/util"
	"github.com/cucumber/messages-go/v10"
)

var (
	shell = &ShellInstance{}
)

type ShellInstance struct {
	startArgument []string
	name          string
	exitCode      int

	outbuf bytes.Buffer
	errbuf bytes.Buffer
}

func (shell *ShellInstance) GetLastCmdOutput(stdType string) string {
	var returnValue string
	switch stdType {
	case "stdout":
		returnValue = shell.outbuf.String()
	case "stderr":
		returnValue = shell.errbuf.String()
	case "exitcode":
		returnValue = strconv.Itoa(shell.exitCode)
	default:
		fmt.Printf("Field '%s' of shell's output is not supported. Only 'stdout', 'stderr' and 'exitcode' are supported.", stdType)
	}

	returnValue = strings.TrimSuffix(returnValue, "\n")

	return returnValue
}

func (shell *ShellInstance) ConfigureTypeOfShell(shellName string) {
	switch shellName {
	case "bash":
		shell.name = shellName
	case "tcsh":
		shell.name = shellName
	case "zsh":
		shell.name = shellName
	case "cmd":
		shell.name = shellName
	case "powershell":
		shell.name = shellName
		shell.startArgument = []string{"-Command", "-"}
	case "fish":
		fmt.Println("Fish shell is currently not supported by integration tests. Default shell for the OS will be used.")
		fallthrough
	default:
		if shell.name != "" {
			fmt.Printf("Shell %v is not supported, will set the default shell for the OS to be used.\n", shell.name)
		}
		switch runtime.GOOS {
		case "darwin", "linux":
			shell.name = "bash"
		case "windows":
			shell.name = "powershell"
			shell.startArgument = []string{"-Command", "-"}
		}
	}
	return
}

func SetUpHostShellInstance(shellName string) error {
	return shell.Setup(shellName)
}

func (shell *ShellInstance) Setup(shellName string) error {
	if shell.name == "" {
		shell.ConfigureTypeOfShell(shellName)
	}

	fmt.Printf("The %v shell will be used for testing.\n", shell.name)
	return nil
}

func ExecuteCommand(command string) error {
	if shell.name == "" {
		return errors.New("shell instance is not initialized")
	}
	var (
		exitError *exec.ExitError
		pathError *os.PathError
	)
	shell.exitCode = 0
	shell.outbuf.Reset()
	shell.errbuf.Reset()

	util.LogMessage(shell.name, command)

	cmd := exec.Command(shell.name, shell.startArgument...)
	cmd.Stderr = &shell.errbuf
	cmd.Stdout = &shell.outbuf
	cmd.Stdin = strings.NewReader(command + "\n")

	if err := cmd.Run(); err != nil {
		switch {
		case errors.As(err, &exitError):
			shell.exitCode = exitError.ExitCode()
		case errors.As(err, &pathError):
			return fmt.Errorf("os.PathError: %w", err)
		default:
			return fmt.Errorf("something went wrong with %s command: %w", command, err)
		}
	}
	return nil
}

func ExecuteCommandSucceedsOrFails(command string, expectedResult string) error {
	err := ExecuteCommand(command)
	if err != nil {
		return err
	}

	exitCode := strconv.Itoa(shell.exitCode)

	if expectedResult == "succeeds" && exitCode != "0" {
		err = fmt.Errorf("command '%s', expected to succeed, exited with exit code: %s\nCommand stdout: %s\nCommand stderr: %s", command, exitCode, shell.outbuf.String(), shell.errbuf.String())
	}
	if expectedResult == "fails" && exitCode == "0" {
		err = fmt.Errorf("command '%s', expected to fail, exited with exit code: %s\nCommand stdout: %s\nCommand stderr: %s", command, exitCode, shell.outbuf.String(), shell.errbuf.String())
	}

	return err
}

func ExecuteCommandWithRetry(retryCount int, retryTime string, command string, containsOrNot string, expected string) error {
	var exitCode, stdout string
	retryDuration, err := time.ParseDuration(retryTime)
	if err != nil {
		return err
	}

	for i := 0; i < retryCount; i++ {
		err := ExecuteCommand(command)
		exitCode, stdout := strconv.Itoa(shell.exitCode), shell.outbuf.String()
		if strings.Contains(containsOrNot, " not ") {
			if err == nil && exitCode == "0" && !strings.Contains(stdout, expected) {
				return nil
			}
		} else {
			if err == nil && exitCode == "0" && strings.Contains(stdout, expected) {
				return nil
			}
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("command '%s', Expected: exitCode 0, stdout %s, Actual: exitCode %s, stdout %s", command, expected, exitCode, stdout)
}

func ExecuteStdoutLineByLine() error {
	var err error
	stdout := shell.GetLastCmdOutput("stdout")
	commandArray := strings.Split(stdout, "\n")
	for index := range commandArray {
		err = ExecuteCommand(commandArray[index])
	}

	return err
}

func CommandReturnShouldContain(commandField string, expected string) error {
	return CompareExpectedWithActualContains(expected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldContainContent(commandField string, expected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualContains(expected.Content, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotContain(commandField string, notexpected string) error {
	return CompareExpectedWithActualNotContains(notexpected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotContainContent(commandField string, notexpected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualNotContains(notexpected.Content, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldBeEmpty(commandField string) error {
	return CompareExpectedWithActualEquals("", shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotBeEmpty(commandField string) error {
	return CompareExpectedWithActualNotEquals("", shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldEqual(commandField string, expected string) error {
	return CompareExpectedWithActualEquals(expected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldEqualContent(commandField string, expected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualEquals(expected.Content, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotEqual(commandField string, expected string) error {
	return CompareExpectedWithActualNotEquals(expected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotEqualContent(commandField string, expected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualNotEquals(expected.Content, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldMatch(commandField string, expected string) error {
	return CompareExpectedWithActualMatchesRegex(expected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldMatchContent(commandField string, expected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualMatchesRegex(expected.Content, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotMatch(commandField string, expected string) error {
	return CompareExpectedWithActualNotMatchesRegex(expected, shell.GetLastCmdOutput(commandField))
}

func CommandReturnShouldNotMatchContent(commandField string, expected *messages.PickleStepArgument_PickleDocString) error {
	return CompareExpectedWithActualNotMatchesRegex(expected.Content, shell.GetLastCmdOutput(commandField))
}

func ShouldBeInValidFormat(commandField string, format string) error {
	return CheckFormat(format, shell.GetLastCmdOutput(commandField))
}

func SetScenarioVariableExecutingCommand(variableName string, command string) error {
	err := ExecuteCommand(command)
	if err != nil {
		return err
	}

	commandFailed := (shell.GetLastCmdOutput("exitcode") != "0" || len(shell.GetLastCmdOutput("stderr")) != 0)
	if commandFailed {
		return fmt.Errorf("command '%v' did not execute successfully. cmdExit: %v, cmdErr: %v",
			command,
			shell.GetLastCmdOutput("exitcode"),
			shell.GetLastCmdOutput("stderr"))
	}

	stdout := shell.GetLastCmdOutput("stdout")
	util.SetScenarioVariable(variableName, strings.TrimSpace(stdout))

	return nil
}

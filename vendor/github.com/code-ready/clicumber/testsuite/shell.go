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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cucumber/messages-go/v10"
	"github.com/code-ready/clicumber/util"
)

const (
	exitCodeIdentifier = "exitCodeOfLastCommandInShell="

	bashExitCodeCheck       = "echo %v$?"
	fishExitCodeCheck       = "echo %v$status"
	tcshExitCodeCheck       = "echo %v$?"
	zshExitCodeCheck        = "echo %v$?"
	cmdExitCodeCheck        = "echo %v%%errorlevel%%"
	powershellExitCodeCheck = "echo %v$lastexitcode"
)

var (
	shell ShellInstance
)

type ShellInstance struct {
	startArgument    []string
	name             string
	checkExitCodeCmd string

	instance *exec.Cmd
	outbuf   bytes.Buffer
	errbuf   bytes.Buffer
	excbuf   bytes.Buffer

	outPipe io.ReadCloser
	errPipe io.ReadCloser
	inPipe  io.WriteCloser

	outScanner *bufio.Scanner
	errScanner *bufio.Scanner

	stdoutChannel   chan string
	stderrChannel   chan string
	exitCodeChannel chan string
}

func (shell ShellInstance) GetLastCmdOutput(stdType string) string {
	var returnValue string
	switch stdType {
	case "stdout":
		returnValue = shell.outbuf.String()
	case "stderr":
		returnValue = shell.errbuf.String()
	case "exitcode":
		returnValue = shell.excbuf.String()
	default:
		fmt.Printf("Field '%s' of shell's output is not supported. Only 'stdout', 'stderr' and 'exitcode' are supported.", stdType)
	}

	returnValue = strings.TrimSuffix(returnValue, "\n")

	return returnValue
}

func (shell *ShellInstance) ScanPipe(scanner *bufio.Scanner, buffer *bytes.Buffer, stdType string, channel chan string) {
	for scanner.Scan() {
		str := scanner.Text()
		util.LogMessage(stdType, str)

		if strings.Contains(str, exitCodeIdentifier) && !strings.Contains(str, shell.checkExitCodeCmd) {
			exitCode := strings.Split(str, "=")[1]
			shell.exitCodeChannel <- exitCode
		} else {
			buffer.WriteString(str + "\n")
		}
	}

	return
}

func (shell *ShellInstance) ConfigureTypeOfShell(shellName string) {
	switch shellName {
	case "bash":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
	case "tcsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(tcshExitCodeCheck, exitCodeIdentifier)
	case "zsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(zshExitCodeCheck, exitCodeIdentifier)
	case "cmd":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(cmdExitCodeCheck, exitCodeIdentifier)
	case "powershell":
		shell.name = shellName
		shell.startArgument = []string{"-Command", "-"}
		shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
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
			shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
		case "windows":
			shell.name = "powershell"
			shell.startArgument = []string{"-Command", "-"}
			shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
		}
	}

	return
}

func StartHostShellInstance(shellName string) error {
	return shell.Start(shellName)
}

func (shell *ShellInstance) Start(shellName string) error {
	var err error

	if shell.name == "" {
		shell.ConfigureTypeOfShell(shellName)
	}
	shell.stdoutChannel = make(chan string)
	shell.stderrChannel = make(chan string)
	shell.exitCodeChannel = make(chan string)

	shell.instance = exec.Command(shell.name, shell.startArgument...)

	shell.outPipe, err = shell.instance.StdoutPipe()
	if err != nil {
		return err
	}

	shell.errPipe, err = shell.instance.StderrPipe()
	if err != nil {
		return err
	}

	shell.inPipe, err = shell.instance.StdinPipe()
	if err != nil {
		return err
	}

	shell.outScanner = bufio.NewScanner(shell.outPipe)
	shell.errScanner = bufio.NewScanner(shell.errPipe)

	go shell.ScanPipe(shell.outScanner, &shell.outbuf, "stdout", shell.stdoutChannel)
	go shell.ScanPipe(shell.errScanner, &shell.errbuf, "stderr", shell.stderrChannel)

	err = shell.instance.Start()
	if err != nil {
		return err
	}

	fmt.Printf("The %v instance has been started and will be used for testing.\n", shell.name)
	return err
}

func CloseHostShellInstance() error {
	return shell.Close()
}

func (shell *ShellInstance) Close() error {
	closingCmd := "exit\n"
	io.WriteString(shell.inPipe, closingCmd)
	err := shell.instance.Wait()
	if err != nil {
		fmt.Println("error closing shell instance:", err)
	}

	shell.instance = nil

	return err
}

func ExecuteCommand(command string) error {
	if shell.instance == nil {
		return errors.New("shell instance is not started")
	}

	shell.outbuf.Reset()
	shell.errbuf.Reset()
	shell.excbuf.Reset()

	util.LogMessage(shell.name, command)

	_, err := io.WriteString(shell.inPipe, command+"\n")
	if err != nil {
		return err
	}

	_, err = shell.inPipe.Write([]byte(shell.checkExitCodeCmd + "\n"))
	if err != nil {
		return err
	}

	exitCode := <-shell.exitCodeChannel
	shell.excbuf.WriteString(exitCode)

	return err
}

func ExecuteCommandSucceedsOrFails(command string, expectedResult string) error {
	err := ExecuteCommand(command)
	if err != nil {
		return err
	}

	exitCode := shell.excbuf.String()

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
		exitCode, stdout := shell.excbuf.String(), shell.outbuf.String()
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
		if !strings.Contains(commandArray[index], exitCodeIdentifier) {
			err = ExecuteCommand(commandArray[index])
		}
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

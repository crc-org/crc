// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyring

import (
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
)

const (
	execPathKeychain = "/usr/bin/security"

	// encodingPrefix is a well-known prefix added to strings encoded by Set.
	encodingPrefix = "go-keyring-encoded:"
)

type macOSXKeychain struct{}

// func (*MacOSXKeychain) IsAvailable() bool {
// 	return exec.Command(execPathKeychain).Run() != exec.ErrNotFound
// }

// Set stores stores user and pass in the keyring under the defined service
// name.
func (k macOSXKeychain) Get(service, username string) (string, error) {
	out, err := exec.Command(
		execPathKeychain,
		"find-generic-password",
		"-s", service,
		"-wa", username).CombinedOutput()
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
			err = ErrNotFound
		}
		return "", err
	}

	trimStr := strings.TrimSpace(string(out[:]))
	// if the string has the well-known prefix, assume it's encoded
	if strings.HasPrefix(trimStr, encodingPrefix) {
		dec, err := hex.DecodeString(trimStr[len(encodingPrefix):])
		return string(dec), err
	}

	return trimStr, nil
}

// Set stores a secret in the keyring given a service name and a user.
func (k macOSXKeychain) Set(service, username, password string) error {
	// if the added secret has multiple lines or some non ascii,
	// osx will hex encode it on return. To avoid getting garbage, we
	// encode all passwords
	password = encodingPrefix + hex.EncodeToString([]byte(password))

	cmd := exec.Command(execPathKeychain, "-i")
	stdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	command := fmt.Sprintf("add-generic-password -U -s %s -a %s -w %s\n", shellescape.Quote(service), shellescape.Quote(username), shellescape.Quote(password))
	if _, err := io.WriteString(stdIn, command); err != nil {
		return err
	}

	if err = stdIn.Close(); err != nil {
		return err
	}

	err = cmd.Wait()
	return err
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k macOSXKeychain) Delete(service, username string) error {
	out, err := exec.Command(
		execPathKeychain,
		"delete-generic-password",
		"-s", service,
		"-a", username).CombinedOutput()
	if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
		err = ErrNotFound
	}
	return err
}

func init() {
	provider = macOSXKeychain{}
}

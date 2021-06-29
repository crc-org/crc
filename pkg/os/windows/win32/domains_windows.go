// Copyright 2021 Google LLC
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

// copied from https://github.com/google/glazier/blob/a00964a46d35de3f6193f0d3f1b4e490e9630e19/go/identity/identity.go

package win32

import (
	"errors"

	"github.com/StackExchange/wmi"
)

//nolint
type Win32_ComputerSystem struct {
	Partofdomain bool
}

// DomainJoined attempts to determine whether the machine is actively domain joined.
func DomainJoined() (bool, error) {
	var c []Win32_ComputerSystem
	q := wmi.CreateQuery(&c, "")

	if err := wmi.Query(q, &c); err != nil {
		return false, err
	}
	if len(c) < 1 {
		return false, errors.New("no result from wmi query")
	}
	return c[0].Partofdomain, nil
}

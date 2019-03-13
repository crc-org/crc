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

package util

import (
	"fmt"
	"strings"
)

var scenarioVariables []scenarioVariable

type scenarioVariable struct {
	Name  string
	Value string
}

func ProcessScenarioVariables(command string) string {
	for _, variable := range scenarioVariables {
		command = strings.Replace(command, fmt.Sprintf("$(%s)", variable.Name), variable.Value, -1)
	}

	return command
}

func ClearScenarioVariables() {
	scenarioVariables = nil
}

func SetScenarioVariable(name string, value string) {
	newVariable := scenarioVariable{
		name,
		value,
	}

	scenarioVariables = append([]scenarioVariable{newVariable}, scenarioVariables...)
}

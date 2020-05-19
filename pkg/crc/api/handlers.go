package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/crc/version"
)

func statusHandler(client machine.Client, _ json.RawMessage) string {
	statusConfig := machine.ClusterStatusConfig{Name: constants.DefaultName}
	clusterStatus, _ := client.Status(statusConfig)
	return encodeStructToJSON(clusterStatus)
}

func stopHandler(client machine.Client, _ json.RawMessage) string {
	stopConfig := machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: true,
	}
	commandResult, _ := client.Stop(stopConfig)
	return encodeStructToJSON(commandResult)
}

func startHandler(client machine.Client, _ json.RawMessage) string {
	startConfig := machine.StartConfig{
		Name:          constants.DefaultName,
		BundlePath:    crcConfig.GetString(config.Bundle.Name),
		Memory:        crcConfig.GetInt(config.Memory.Name),
		CPUs:          crcConfig.GetInt(config.CPUs.Name),
		NameServer:    crcConfig.GetString(config.NameServer.Name),
		GetPullSecret: getPullSecretFileContent,
		Debug:         true,
	}
	status, _ := client.Start(startConfig)
	return encodeStructToJSON(status)
}

func versionHandler(client machine.Client, _ json.RawMessage) string {
	v := &machine.VersionResult{
		CrcVersion:       version.GetCRCVersion(),
		CommitSha:        version.GetCommitSha(),
		OpenshiftVersion: version.GetBundleVersion(),
		Success:          true,
	}
	return encodeStructToJSON(v)
}

func getPullSecretFileContent() (string, error) {
	data, err := ioutil.ReadFile(crcConfig.GetString(config.PullSecretFile.Name))
	if err != nil {
		return "", err
	}
	pullsecret := string(data)
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", err
	}
	return pullsecret, nil
}

func deleteHandler(client machine.Client, _ json.RawMessage) string {
	delConfig := machine.DeleteConfig{Name: constants.DefaultName}
	r, _ := client.Delete(delConfig)
	return encodeStructToJSON(r)
}

func webconsoleURLHandler(client machine.Client, _ json.RawMessage) string {
	consoleConfig := machine.ConsoleConfig{Name: constants.DefaultName}
	r, _ := client.GetConsoleURL(consoleConfig)
	return encodeStructToJSON(r)
}

func setConfigHandler(_ machine.Client, args json.RawMessage) string {
	setConfigResult := setOrUnsetConfigResult{}
	if args == nil {
		setConfigResult.Error = "No config keys provided"
		return encodeStructToJSON(setConfigResult)
	}

	var multiError = errors.MultiError{}
	var a = make(map[string]interface{})

	err := json.Unmarshal(args, &a)
	if err != nil {
		setConfigResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(setConfigResult)
	}

	configs := a["properties"].(map[string]interface{})

	// successProps slice contains the properties that were successfully set
	var successProps []string

	for k, v := range configs {
		_, err := crcConfig.Set(k, v)
		if err != nil {
			multiError.Collect(err)
			continue
		}
		err = crcConfig.WriteConfig()
		if err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, k)
	}

	if len(multiError.Errors) != 0 {
		setConfigResult.Error = fmt.Sprintf("%v", multiError.ToError())
	}

	setConfigResult.Properties = successProps
	return encodeStructToJSON(setConfigResult)
}

func unsetConfigHandler(_ machine.Client, args json.RawMessage) string {
	unsetConfigResult := setOrUnsetConfigResult{}
	if args == nil {
		unsetConfigResult.Error = "No config keys provided"
		return encodeStructToJSON(unsetConfigResult)
	}

	var multiError = errors.MultiError{}
	var keys = make(map[string][]string)

	err := json.Unmarshal(args, &keys)
	if err != nil {
		unsetConfigResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(unsetConfigResult)
	}

	// successProps slice contains the properties that were successfully unset
	var successProps []string

	keysToUnset := keys["properties"]
	for _, key := range keysToUnset {
		_, err := crcConfig.Unset(key)
		if err != nil {
			multiError.Collect(err)
			continue
		}
		err = crcConfig.WriteConfig()
		if err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, key)
	}
	if len(multiError.Errors) != 0 {
		unsetConfigResult.Error = fmt.Sprintf("%v", multiError.ToError())
	}
	unsetConfigResult.Properties = successProps
	return encodeStructToJSON(unsetConfigResult)
}

func getConfigHandler(_ machine.Client, args json.RawMessage) string {
	configResult := getConfigResult{}
	if args == nil {
		allConfigs := crcConfig.AllConfigs()
		configResult.Error = ""
		configResult.Configs = allConfigs
		return encodeStructToJSON(configResult)
	}

	var a = make(map[string][]string)

	err := json.Unmarshal(args, &a)
	if err != nil {
		configResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(configResult)
	}

	keys := a["properties"]

	var configs = make(map[string]interface{})

	for _, key := range keys {
		v, err := crcConfig.Get(key)
		if err != nil {
			continue
		}
		configs[key] = v
	}
	if len(configs) == 0 {
		configResult.Error = "Unable to get configs"
		configResult.Configs = nil
	} else {
		configResult.Error = ""
		configResult.Configs = configs
	}
	return encodeStructToJSON(configResult)
}

func encodeStructToJSON(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		logging.Error(err.Error())
		err := commandError{
			Err: "Failed while encoding JSON to string",
		}
		s, _ := json.Marshal(err)
		return string(s)
	}
	return string(s)
}

func encodeErrorToJSON(errMsg string) string {
	err := commandError{
		Err: errMsg,
	}
	return encodeStructToJSON(err)
}

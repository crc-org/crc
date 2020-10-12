package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/crc/version"
)

type Handler struct {
	MachineClient AdaptedClient
}

func newHandler() *Handler {
	return &Handler{
		MachineClient: &Adapter{Underlying: machine.NewClient()},
	}
}

func (h *Handler) Status() string {
	statusConfig := machine.ClusterStatusConfig{Name: constants.DefaultName}
	clusterStatus, _ := h.MachineClient.Status(statusConfig)
	return encodeStructToJSON(clusterStatus)
}

func (h *Handler) Stop() string {
	stopConfig := machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: true,
	}
	commandResult, _ := h.MachineClient.Stop(stopConfig)
	return encodeStructToJSON(commandResult)
}

func (h *Handler) Start(crcConfig crcConfig.Storage, args json.RawMessage) string {
	var parsedArgs startArgs
	var err error
	if args != nil {
		parsedArgs, err = parseStartArgs(args)
		if err != nil {
			startErr := &machine.StartResult{
				Name:  constants.DefaultName,
				Error: fmt.Sprintf("Incorrect arguments given: %s", err.Error()),
			}
			return encodeStructToJSON(startErr)
		}
	}
	if err := preflight.StartPreflightChecks(crcConfig); err != nil {
		startErr := &machine.StartResult{
			Name:  constants.DefaultName,
			Error: err.Error(),
		}
		return encodeStructToJSON(startErr)
	}

	startConfig := getStartConfig(crcConfig, parsedArgs)
	status, _ := h.MachineClient.Start(startConfig)
	return encodeStructToJSON(status)
}

func parseStartArgs(args json.RawMessage) (startArgs, error) {
	var parsedArgs startArgs
	dec := json.NewDecoder(bytes.NewReader(args))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&parsedArgs); err != nil {
		return startArgs{}, err
	}
	return parsedArgs, nil
}

func getStartConfig(cfg crcConfig.Storage, args startArgs) machine.StartConfig {
	startConfig := machine.StartConfig{
		Name:       constants.DefaultName,
		BundlePath: cfg.Get(config.Bundle).AsString(),
		Memory:     cfg.Get(config.Memory).AsInt(),
		CPUs:       cfg.Get(config.CPUs).AsInt(),
		NameServer: cfg.Get(config.NameServer).AsString(),
		Debug:      true,
	}
	pullSecretFile := args.PullSecretFile
	if pullSecretFile == "" {
		pullSecretFile = cfg.Get(config.PullSecretFile).AsString()
	}
	startConfig.PullSecret = &cluster.PullSecret{
		Getter: getPullSecretFileContent(pullSecretFile),
	}
	return startConfig
}

func (h *Handler) GetVersion() string {
	v := &machine.VersionResult{
		CrcVersion:       version.GetCRCVersion(),
		CommitSha:        version.GetCommitSha(),
		OpenshiftVersion: version.GetBundleVersion(),
		Success:          true,
	}
	return encodeStructToJSON(v)
}

func getPullSecretFileContent(path string) func() (string, error) {
	return func() (string, error) {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		pullsecret := string(data)
		if err := validation.ImagePullSecret(pullsecret); err != nil {
			return "", err
		}
		return pullsecret, nil
	}
}

func (h *Handler) Delete() string {
	delConfig := machine.DeleteConfig{Name: constants.DefaultName}
	r := h.MachineClient.Delete(delConfig)
	return encodeStructToJSON(r)
}

func (h *Handler) GetWebconsoleInfo() string {
	consoleConfig := machine.ConsoleConfig{Name: constants.DefaultName}
	r, _ := h.MachineClient.GetConsoleURL(consoleConfig)
	return encodeStructToJSON(r)
}

func (h *Handler) SetConfig(crcConfig crcConfig.Storage, args json.RawMessage) string {
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
		successProps = append(successProps, k)
	}

	if len(multiError.Errors) != 0 {
		setConfigResult.Error = fmt.Sprintf("%v", multiError)
	}

	setConfigResult.Properties = successProps
	return encodeStructToJSON(setConfigResult)
}

func (h *Handler) UnsetConfig(crcConfig crcConfig.Storage, args json.RawMessage) string {
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
		successProps = append(successProps, key)
	}
	if len(multiError.Errors) != 0 {
		unsetConfigResult.Error = fmt.Sprintf("%v", multiError)
	}
	unsetConfigResult.Properties = successProps
	return encodeStructToJSON(unsetConfigResult)
}

func (h *Handler) GetConfig(crcConfig crcConfig.Storage, args json.RawMessage) string {
	configResult := getConfigResult{}
	if args == nil {
		allConfigs := crcConfig.AllConfigs()
		configResult.Error = ""
		configResult.Configs = make(map[string]interface{})
		for k, v := range allConfigs {
			configResult.Configs[k] = v.Value
		}
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
		v := crcConfig.Get(key)
		if v.Invalid {
			continue
		}
		configs[key] = v.Value
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

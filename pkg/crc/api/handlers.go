package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/api/client"
	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/version"
)

type Handler struct {
	Logger        Logger
	MachineClient AdaptedClient
	Config        crcConfig.Storage
	Telemetry     Telemetry
}

type Logger interface {
	Messages() []string
}

type Telemetry interface {
	UploadAction(action, source, status string) error
}

func (h *Handler) Logs() string {
	return encodeStructToJSON(&loggerResult{
		Success:  true,
		Messages: h.Logger.Messages(),
	})
}

func NewHandler(config crcConfig.Storage, machine machine.Client, logger Logger, telemetry Telemetry) *Handler {
	return &Handler{
		MachineClient: &Adapter{
			Underlying: machine,
		},
		Config:    config,
		Logger:    logger,
		Telemetry: telemetry,
	}
}

func (h *Handler) Status() string {
	clusterStatus := h.MachineClient.Status()
	return encodeStructToJSON(clusterStatus)
}

func (h *Handler) Stop() string {
	commandResult := h.MachineClient.Stop()
	return encodeStructToJSON(commandResult)
}

func (h *Handler) PowerOff() string {
	commandResult := h.MachineClient.PowerOff()
	return encodeStructToJSON(commandResult)
}

func (h *Handler) Start(args json.RawMessage) string {
	var parsedArgs startArgs
	var err error
	if args != nil {
		parsedArgs, err = parseStartArgs(args)
		if err != nil {
			startErr := &client.StartResult{
				Success: false,
				Error:   fmt.Sprintf("Incorrect arguments given: %s", err.Error()),
			}
			return encodeStructToJSON(startErr)
		}
	}
	if err := preflight.StartPreflightChecks(h.Config); err != nil {
		startErr := &client.StartResult{
			Success: false,
			Error:   err.Error(),
		}
		return encodeStructToJSON(startErr)
	}

	startConfig := getStartConfig(h.Config, parsedArgs)
	status := h.MachineClient.Start(context.Background(), startConfig)
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

func getStartConfig(cfg crcConfig.Storage, args startArgs) types.StartConfig {
	return types.StartConfig{
		BundlePath: cfg.Get(crcConfig.Bundle).AsString(),
		Memory:     cfg.Get(crcConfig.Memory).AsInt(),
		CPUs:       cfg.Get(crcConfig.CPUs).AsInt(),
		NameServer: cfg.Get(crcConfig.NameServer).AsString(),
		PullSecret: cluster.NewNonInteractivePullSecretLoader(cfg, args.PullSecretFile),
	}
}

func (h *Handler) GetVersion() string {
	v := &client.VersionResult{
		CrcVersion:       version.GetCRCVersion(),
		CommitSha:        version.GetCommitSha(),
		OpenshiftVersion: version.GetBundleVersion(),
		Success:          true,
	}
	return encodeStructToJSON(v)
}

func (h *Handler) Delete() string {
	r := h.MachineClient.Delete()
	return encodeStructToJSON(r)
}

func (h *Handler) GetWebconsoleInfo() string {
	r := h.MachineClient.GetConsoleURL()
	return encodeStructToJSON(r)
}

func (h *Handler) SetConfig(args json.RawMessage) string {
	setConfigResult := client.SetOrUnsetConfigResult{
		Success: true,
	}
	if args == nil {
		setConfigResult.Success = false
		setConfigResult.Error = "No config keys provided"
		return encodeStructToJSON(setConfigResult)
	}

	var multiError = errors.MultiError{}
	var a = make(map[string]interface{})

	err := json.Unmarshal(args, &a)
	if err != nil {
		setConfigResult.Success = false
		setConfigResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(setConfigResult)
	}

	configs, ok := a["properties"].(map[string]interface{})
	if !ok {
		setConfigResult.Success = false
		setConfigResult.Error = "No config keys provided"
		return encodeStructToJSON(setConfigResult)
	}

	// successProps slice contains the properties that were successfully set
	var successProps []string

	for k, v := range configs {
		_, err := h.Config.Set(k, v)
		if err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, k)
	}

	if len(multiError.Errors) != 0 {
		setConfigResult.Success = false
		setConfigResult.Error = fmt.Sprintf("%v", multiError)
	}

	setConfigResult.Properties = successProps
	return encodeStructToJSON(setConfigResult)
}

func (h *Handler) UnsetConfig(args json.RawMessage) string {
	unsetConfigResult := client.SetOrUnsetConfigResult{
		Success: true,
	}
	if args == nil {
		unsetConfigResult.Success = false
		unsetConfigResult.Error = "No config keys provided"
		return encodeStructToJSON(unsetConfigResult)
	}

	var multiError = errors.MultiError{}
	var keys = make(map[string][]string)

	err := json.Unmarshal(args, &keys)
	if err != nil {
		unsetConfigResult.Success = false
		unsetConfigResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(unsetConfigResult)
	}

	// successProps slice contains the properties that were successfully unset
	var successProps []string

	keysToUnset := keys["properties"]
	for _, key := range keysToUnset {
		_, err := h.Config.Unset(key)
		if err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, key)
	}
	if len(multiError.Errors) != 0 {
		unsetConfigResult.Success = false
		unsetConfigResult.Error = fmt.Sprintf("%v", multiError)
	}
	unsetConfigResult.Properties = successProps
	return encodeStructToJSON(unsetConfigResult)
}

func (h *Handler) GetConfig(args json.RawMessage) string {
	configResult := client.GetConfigResult{
		Success: true,
	}
	if args == nil {
		allConfigs := h.Config.AllConfigs()
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
		configResult.Success = false
		configResult.Error = fmt.Sprintf("%v", err)
		return encodeStructToJSON(configResult)
	}

	keys := a["properties"]

	var configs = make(map[string]interface{})

	for _, key := range keys {
		v := h.Config.Get(key)
		if v.Invalid {
			continue
		}
		configs[key] = v.Value
	}
	if len(configs) == 0 {
		configResult.Success = false
		configResult.Error = "Unable to get configs"
		configResult.Configs = nil
	} else {
		configResult.Error = ""
		configResult.Configs = configs
	}
	return encodeStructToJSON(configResult)
}

func (h *Handler) UploadTelemetry(args json.RawMessage) string {
	var req client.TelemetryRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return encodeErrorToJSON(err.Error())
	}
	if err := h.Telemetry.UploadAction(req.Action, req.Source, req.Status); err != nil {
		return encodeErrorToJSON(err.Error())
	}
	return encodeStructToJSON(client.Result{
		Success: true,
	})
}

func encodeStructToJSON(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		logging.Error(err.Error())
		err := client.Result{
			Success: false,
			Error:   "Failed while encoding JSON to string",
		}
		s, _ := json.Marshal(err)
		return string(s)
	}
	return string(s)
}

func encodeErrorToJSON(errMsg string) string {
	err := client.Result{
		Success: false,
		Error:   errMsg,
	}
	return encodeStructToJSON(err)
}

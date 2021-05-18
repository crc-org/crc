package api

import (
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
	var parsedArgs client.StartConfig
	if args != nil {
		if err := json.Unmarshal(args, &parsedArgs); err != nil {
			return encodeStructToJSON(&client.StartResult{
				Success: false,
				Error:   fmt.Sprintf("Incorrect arguments given: %s", err.Error()),
			})
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

func getStartConfig(cfg crcConfig.Storage, args client.StartConfig) types.StartConfig {
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
	if args == nil {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   "No config keys provided",
		})
	}

	var req client.SetConfigRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   err.Error(),
		})
	}

	// successProps slice contains the properties that were successfully set
	var successProps []string
	var multiError = errors.MultiError{}
	for k, v := range req.Properties {
		_, err := h.Config.Set(k, v)
		if err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, k)
	}
	if len(multiError.Errors) != 0 {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   multiError.Error(),
		})
	}
	return encodeStructToJSON(client.SetOrUnsetConfigResult{
		Success:    true,
		Properties: successProps,
	})
}

func (h *Handler) UnsetConfig(args json.RawMessage) string {
	if args == nil {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   "No config keys provided",
		})
	}

	var req client.GetOrUnsetConfigRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   err.Error(),
		})
	}

	// successProps slice contains the properties that were successfully unset
	var successProps []string
	var multiError = errors.MultiError{}
	for _, key := range req.Properties {
		if _, err := h.Config.Unset(key); err != nil {
			multiError.Collect(err)
			continue
		}
		successProps = append(successProps, key)
	}
	if len(multiError.Errors) != 0 {
		return encodeStructToJSON(client.SetOrUnsetConfigResult{
			Success: false,
			Error:   multiError.Error(),
		})
	}
	return encodeStructToJSON(client.SetOrUnsetConfigResult{
		Success:    true,
		Properties: successProps,
	})
}

func (h *Handler) GetConfig(args json.RawMessage) string {
	if args == nil {
		allConfigs := h.Config.AllConfigs()
		configs := make(map[string]interface{})
		for k, v := range allConfigs {
			configs[k] = v.Value
		}
		return encodeStructToJSON(client.GetConfigResult{
			Success: true,
			Configs: configs,
		})
	}

	var req client.GetOrUnsetConfigRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return encodeStructToJSON(client.GetConfigResult{
			Success: false,
			Error:   err.Error(),
		})
	}

	configs := make(map[string]interface{})
	for _, key := range req.Properties {
		v := h.Config.Get(key)
		if v.Invalid {
			continue
		}
		configs[key] = v.Value
	}

	if len(configs) == 0 {
		return encodeStructToJSON(client.GetConfigResult{
			Success: false,
			Error:   "Unable to get configs",
		})
	}
	return encodeStructToJSON(client.GetConfigResult{
		Success: true,
		Configs: configs,
	})
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

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
	Logger    Logger
	Client    machine.Client
	Config    crcConfig.Storage
	Telemetry Telemetry
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
		Client:    machine,
		Config:    config,
		Logger:    logger,
		Telemetry: telemetry,
	}
}

func (h *Handler) Status() string {
	res, err := h.Client.Status()
	if err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.ClusterStatusResult{
			Error:   err.Error(),
			Success: false,
		})
	}
	return encodeStructToJSON(client.ClusterStatusResult{
		CrcStatus:        string(res.CrcStatus),
		OpenshiftStatus:  string(res.OpenshiftStatus),
		OpenshiftVersion: res.OpenshiftVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		Success:          true,
	})
}

func (h *Handler) Stop() string {
	_, err := h.Client.Stop()
	if err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.Result{
			Success: false,
			Error:   err.Error(),
		})
	}
	return encodeStructToJSON(client.Result{
		Success: true,
	})
}

func (h *Handler) PowerOff() string {
	if err := h.Client.PowerOff(); err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.Result{
			Success: false,
			Error:   err.Error(),
		})
	}
	return encodeStructToJSON(client.Result{
		Success: true,
	})
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
	res, err := h.Client.Start(context.Background(), startConfig)
	if err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.StartResult{
			Success: false,
			Error:   err.Error(),
		})
	}
	return encodeStructToJSON(client.StartResult{
		Success:        true,
		Status:         string(res.Status),
		ClusterConfig:  res.ClusterConfig,
		KubeletStarted: res.KubeletStarted,
	})
}

func getStartConfig(cfg crcConfig.Storage, args client.StartConfig) types.StartConfig {
	return types.StartConfig{
		BundlePath:        cfg.Get(crcConfig.Bundle).AsString(),
		Memory:            cfg.Get(crcConfig.Memory).AsInt(),
		DiskSize:          cfg.Get(crcConfig.DiskSize).AsInt(),
		CPUs:              cfg.Get(crcConfig.CPUs).AsInt(),
		NameServer:        cfg.Get(crcConfig.NameServer).AsString(),
		PullSecret:        cluster.NewNonInteractivePullSecretLoader(cfg, args.PullSecretFile),
		KubeAdminPassword: cfg.Get(crcConfig.KubeAdminPassword).AsString(),
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
	err := h.Client.Delete()
	if err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.Result{
			Success: false,
			Error:   err.Error(),
		})
	}
	return encodeStructToJSON(client.Result{
		Success: true,
	})
}

func (h *Handler) GetWebconsoleInfo() string {
	res, err := h.Client.GetConsoleURL()
	if err != nil {
		logging.Error(err)
		return encodeStructToJSON(client.ConsoleResult{
			Success: false,
			Error:   err.Error(),
		})
	}
	return encodeStructToJSON(client.ConsoleResult{
		ClusterConfig: res.ClusterConfig,
		Success:       true,
	})
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

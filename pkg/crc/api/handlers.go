package api

import (
	gocontext "context"
	"net/http"

	"github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preflight"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"
)

type Handler struct {
	Logger    Logger
	Client    machine.Client
	Config    *crcConfig.Config
	Telemetry Telemetry
}

type Logger interface {
	Messages() []string
}

type Telemetry interface {
	UploadAction(action, source, status string) error
}

type loggerResult struct {
	Messages []string
}

func (h *Handler) Logs(c *context) error {
	return c.JSON(http.StatusOK, &loggerResult{
		Messages: h.Logger.Messages(),
	})
}

func NewHandler(config *crcConfig.Config, machine machine.Client, logger Logger, telemetry Telemetry) *Handler {
	return &Handler{
		Client:    machine,
		Config:    config,
		Logger:    logger,
		Telemetry: telemetry,
	}
}

func (h *Handler) Status(c *context) error {
	exists, err := h.Client.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return c.String(http.StatusInternalServerError, string(errors.VMNotExist))
	}

	res, err := h.Client.Status()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.ClusterStatusResult{
		CrcStatus:        string(res.CrcStatus),
		OpenshiftStatus:  string(res.OpenshiftStatus),
		OpenshiftVersion: res.OpenshiftVersion,
		PodmanVersion:    res.PodmanVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		RAMSize:          res.RAMSize,
		RAMUse:           res.RAMUse,
		Preset:           res.Preset,
	})
}

func (h *Handler) Stop(c *context) error {
	_, err := h.Client.Stop()
	if err != nil {
		return err
	}
	return c.Code(http.StatusOK)
}

func (h *Handler) PowerOff(c *context) error {
	if c.method != http.MethodPost {
		return c.String(http.StatusMethodNotAllowed, "Only POST is allowed")
	}
	if err := h.Client.PowerOff(); err != nil {
		return err
	}
	return c.Code(http.StatusOK)
}

func (h *Handler) Start(c *context) error {
	crcConfig.UpdateDefaults(h.Config)
	var parsedArgs client.StartConfig
	if len(c.requestBody) > 0 {
		if err := c.Bind(&parsedArgs); err != nil {
			return err
		}
	}
	if err := preflight.StartPreflightChecks(h.Config); err != nil {
		return err
	}

	startConfig := getStartConfig(h.Config, parsedArgs)
	res, err := h.Client.Start(gocontext.Background(), startConfig)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.StartResult{
		Status:         string(res.Status),
		ClusterConfig:  res.ClusterConfig,
		KubeletStarted: res.KubeletStarted,
	})
}

func getStartConfig(cfg crcConfig.Storage, args client.StartConfig) types.StartConfig {
	return types.StartConfig{
		BundlePath:               cfg.Get(crcConfig.Bundle).AsString(),
		Memory:                   cfg.Get(crcConfig.Memory).AsInt(),
		DiskSize:                 cfg.Get(crcConfig.DiskSize).AsInt(),
		CPUs:                     cfg.Get(crcConfig.CPUs).AsInt(),
		NameServer:               cfg.Get(crcConfig.NameServer).AsString(),
		PullSecret:               cluster.NewNonInteractivePullSecretLoader(cfg, args.PullSecretFile),
		KubeAdminPassword:        cfg.Get(crcConfig.KubeAdminPassword).AsString(),
		IngressHTTPPort:          cfg.Get(crcConfig.IngressHTTPPort).AsUInt(),
		IngressHTTPSPort:         cfg.Get(crcConfig.IngressHTTPSPort).AsUInt(),
		Preset:                   crcConfig.GetPreset(cfg),
		EnableSharedDirs:         cfg.Get(crcConfig.EnableSharedDirs).AsBool(),
		EmergencyLogin:           cfg.Get(crcConfig.EmergencyLogin).AsBool(),
		EnableBundleQuayFallback: cfg.Get(crcConfig.EnableBundleQuayFallback).AsBool(),
	}
}

func (h *Handler) GetVersion(c *context) error {
	return c.JSON(http.StatusOK, &client.VersionResult{
		CrcVersion:       version.GetCRCVersion(),
		CommitSha:        version.GetCommitSha(),
		OpenshiftVersion: version.GetBundleVersion(preset.OpenShift),
		PodmanVersion:    version.GetBundleVersion(preset.Podman),
	})
}

func (h *Handler) Delete(c *context) error {
	err := h.Client.Delete()
	if err != nil {
		return err
	}
	return c.Code(http.StatusOK)
}

func (h *Handler) GetWebconsoleInfo(c *context) error {
	if err := machine.CheckIfMachineMissing(h.Client); err != nil {
		// In case of machine doesn't exist then consoleResult error
		// should be updated so that when rendering the result it have
		// error details also.
		return err
	}
	res, err := h.Client.GetConsoleURL()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.ConsoleResult{
		ClusterConfig: res.ClusterConfig,
		State:         res.State,
	})
}

func (h *Handler) SetConfig(c *context) error {
	var req client.SetConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if len(req.Properties) == 0 {
		return c.JSON(http.StatusBadRequest, client.SetOrUnsetConfigResult{})
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
		return multiError
	}
	return c.JSON(http.StatusOK, client.SetOrUnsetConfigResult{
		Properties: successProps,
	})
}

func (h *Handler) UnsetConfig(c *context) error {
	var req client.GetOrUnsetConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if len(req.Properties) == 0 {
		return c.JSON(http.StatusBadRequest, client.SetOrUnsetConfigResult{})
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
		return multiError
	}
	return c.JSON(http.StatusOK, client.SetOrUnsetConfigResult{
		Properties: successProps,
	})
}

func (h *Handler) GetConfig(c *context) error {
	crcConfig.UpdateDefaults(h.Config)
	queries := c.url.Query()
	var req client.GetOrUnsetConfigRequest
	for key := range queries {
		req.Properties = append(req.Properties, key)
	}

	if len(req.Properties) == 0 {
		allConfigs := h.Config.AllConfigs()
		configs := make(map[string]interface{})
		for k, v := range allConfigs {
			if v.IsSecret {
				continue
			}
			configs[k] = v.Value
		}
		return c.JSON(http.StatusOK, client.GetConfigResult{
			Configs: configs,
		})
	}

	configs := make(map[string]interface{})
	for _, key := range req.Properties {
		v := h.Config.Get(key)
		if v.Invalid {
			continue
		}
		if v.IsSecret {
			continue
		}
		configs[key] = v.Value
	}

	if len(configs) == 0 {
		return c.String(http.StatusInternalServerError, "Unable to get configs")
	}
	return c.JSON(http.StatusOK, client.GetConfigResult{
		Configs: configs,
	})
}

func (h *Handler) UploadTelemetry(c *context) error {
	var req client.TelemetryRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.Telemetry.UploadAction(req.Action, req.Source, req.Status); err != nil {
		return err
	}
	return c.Code(http.StatusOK)
}

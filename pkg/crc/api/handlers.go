package api

import (
	gocontext "context"
	"net/http"

	"github.com/code-ready/crc/pkg/crc/api/client"
	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
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

func (h *Handler) Logs(c *context) error {
	return c.JSON(http.StatusOK, &loggerResult{
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

func (h *Handler) Status(c *context) error {
	res, err := h.Client.Status()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.ClusterStatusResult{
		CrcStatus:        string(res.CrcStatus),
		OpenshiftStatus:  string(res.OpenshiftStatus),
		OpenshiftVersion: res.OpenshiftVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		Success:          true,
	})
}

func (h *Handler) Stop(c *context) error {
	_, err := h.Client.Stop()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.Result{
		Success: true,
	})
}

func (h *Handler) PowerOff(c *context) error {
	if c.method != http.MethodPost {
		return c.String(http.StatusMethodNotAllowed, "Only POST is allowed")
	}
	if err := h.Client.PowerOff(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.Result{
		Success: true,
	})
}

func (h *Handler) Start(c *context) error {
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

func (h *Handler) GetVersion(c *context) error {
	return c.JSON(http.StatusOK, &client.VersionResult{
		CrcVersion:       version.GetCRCVersion(),
		CommitSha:        version.GetCommitSha(),
		OpenshiftVersion: version.GetBundleVersion(),
		Success:          true,
	})
}

func (h *Handler) Delete(c *context) error {
	err := h.Client.Delete()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.Result{
		Success: true,
	})
}

func (h *Handler) GetWebconsoleInfo(c *context) error {
	res, err := h.Client.GetConsoleURL()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, client.ConsoleResult{
		ClusterConfig: res.ClusterConfig,
		Success:       true,
	})
}

func (h *Handler) SetConfig(c *context) error {
	var req client.SetConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if len(req.Properties) == 0 {
		return c.JSON(http.StatusBadRequest, client.SetOrUnsetConfigResult{
			Success: false,
			Error:   "No config keys provided",
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
		return multiError
	}
	return c.JSON(http.StatusOK, client.SetOrUnsetConfigResult{
		Success:    true,
		Properties: successProps,
	})
}

func (h *Handler) UnsetConfig(c *context) error {
	var req client.GetOrUnsetConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if len(req.Properties) == 0 {
		return c.JSON(http.StatusBadRequest, client.SetOrUnsetConfigResult{
			Success: false,
			Error:   "No config keys provided",
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
		return multiError
	}
	return c.JSON(http.StatusOK, client.SetOrUnsetConfigResult{
		Success:    true,
		Properties: successProps,
	})
}

func (h *Handler) GetConfig(c *context) error {
	var req client.GetOrUnsetConfigRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if len(req.Properties) == 0 {
		allConfigs := h.Config.AllConfigs()
		configs := make(map[string]interface{})
		for k, v := range allConfigs {
			configs[k] = v.Value
		}
		return c.JSON(http.StatusOK, client.GetConfigResult{
			Success: true,
			Configs: configs,
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
		return c.String(http.StatusInternalServerError, "Unable to get configs")
	}
	return c.JSON(http.StatusOK, client.GetConfigResult{
		Success: true,
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
	return c.JSON(http.StatusOK, client.Result{
		Success: true,
	})
}

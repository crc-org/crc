package api

import (
	"net/http"

	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func NewMux(config crcConfig.Storage, machine machine.Client, logger Logger, telemetry Telemetry) http.Handler {
	handler := NewHandler(config, machine, logger, telemetry)

	server := newServer()

	server.POST("/start", handler.Start)
	server.GET("/start", handler.Start)

	server.POST("/stop", handler.Stop)
	server.GET("/stop", handler.Stop)

	server.POST("/poweroff", handler.PowerOff)
	server.GET("/status", handler.Status)

	server.DELETE("/delete", handler.Delete)
	server.GET("/delete", handler.Delete)

	server.GET("/version", handler.GetVersion)

	server.GET("/webconsoleurl", handler.GetWebconsoleInfo)

	server.GET("/config/set", handler.SetConfig)
	server.GET("/config/get", handler.GetConfig)
	server.GET("/config/unset", handler.UnsetConfig)
	server.POST("/config/set", handler.SetConfig)
	server.POST("/config/get", handler.GetConfig)
	server.POST("/config/unset", handler.UnsetConfig)

	server.GET("/logs", handler.Logs)

	server.GET("/telemetry", handler.UploadTelemetry)
	server.POST("/telemetry", handler.UploadTelemetry)

	server.GET("/pull-secret", getPullSecret(config))
	server.POST("/pull-secret", setPullSecret())

	return server.Handler()
}

func setPullSecret() func(c *context) error {
	return func(c *context) error {
		if err := cluster.StoreInKeyring(string(c.requestBody)); err != nil {
			return err
		}
		return c.String(http.StatusCreated, "pull secret added")
	}
}

func getPullSecret(config crcConfig.Storage) func(c *context) error {
	return func(c *context) error {
		if _, err := cluster.NewNonInteractivePullSecretLoader(config, "").Value(); err == nil {
			return c.String(http.StatusOK, "pull secret exists")
		}
		return c.String(http.StatusNotFound, "pull secret not found")
	}
}

package api

import (
	"net/http"

	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func NewMux(config crcConfig.Storage, machine machine.Client, logger Logger, telemetry Telemetry) http.Handler {
	handler := NewHandler(config, machine, logger, telemetry)

	server := newServerWithRoutes(handler)

	return server.Handler()
}

func newServerWithRoutes(handler *Handler) *server {
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

	server.GET("/config", handler.GetConfig)
	server.POST("/config", handler.SetConfig)
	server.DELETE("/config", handler.UnsetConfig)

	server.GET("/logs", handler.Logs)

	server.GET("/telemetry", handler.UploadTelemetry)
	server.POST("/telemetry", handler.UploadTelemetry)

	server.GET("/pull-secret", getPullSecret(handler.Config))
	server.POST("/pull-secret", setPullSecret())

	return server
}

func setPullSecret() func(c *context) error {
	return func(c *context) error {
		if err := cluster.StoreInKeyring(string(c.requestBody)); err != nil {
			return err
		}
		return c.String(http.StatusCreated, "")
	}
}

func getPullSecret(config crcConfig.Storage) func(c *context) error {
	return func(c *context) error {
		if _, err := cluster.NewNonInteractivePullSecretLoader(config, "").Value(); err == nil {
			return c.String(http.StatusOK, "")
		}
		return c.String(http.StatusNotFound, "")
	}
}

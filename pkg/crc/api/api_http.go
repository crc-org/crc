package api

import (
	"net/http"

	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func NewMux(config *crcConfig.Config, machine machine.Client, logger Logger, telemetry Telemetry) http.Handler {
	handler := NewHandler(config, machine, logger, telemetry)

	server := newServerWithRoutes(handler)

	return server.Handler()
}

func newServerWithRoutes(handler *Handler) *server {
	server := newServer()

	server.POST("/start", handler.Start, apiV1)
	server.GET("/start", handler.Start, apiV1)

	server.POST("/stop", handler.Stop, apiV1)
	server.GET("/stop", handler.Stop, apiV1)

	server.POST("/poweroff", handler.PowerOff, apiV1)

	server.GET("/status", handler.Status, apiV1)

	server.DELETE("/delete", handler.Delete, apiV1)
	server.GET("/delete", handler.Delete, apiV1)

	server.GET("/version", handler.GetVersion, apiV1)

	server.GET("/webconsoleurl", handler.GetWebconsoleInfo, apiV1)

	server.GET("/config", handler.GetConfig, apiV1)
	server.POST("/config", handler.SetConfig, apiV1)
	server.DELETE("/config", handler.UnsetConfig, apiV1)

	server.GET("/logs", handler.Logs, apiV1)

	server.GET("/telemetry", handler.UploadTelemetry, apiV1)
	server.POST("/telemetry", handler.UploadTelemetry, apiV1)

	server.GET("/pull-secret", getPullSecret(handler.Config), apiV1)
	server.POST("/pull-secret", setPullSecret(), apiV1)

	return server
}

func setPullSecret() func(c *context) error {
	return func(c *context) error {
		if err := cluster.StoreInKeyring(string(c.requestBody)); err != nil {
			return err
		}
		return c.Code(http.StatusCreated)
	}
}

func getPullSecret(config crcConfig.Storage) func(c *context) error {
	return func(c *context) error {
		if _, err := cluster.NewNonInteractivePullSecretLoader(config, "").Value(); err == nil {
			return c.Code(http.StatusOK)
		}
		return c.Code(http.StatusNotFound)
	}
}

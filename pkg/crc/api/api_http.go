package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/code-ready/crc/pkg/crc/api/client"
	"github.com/code-ready/crc/pkg/crc/cluster"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func NewMux(config crcConfig.Storage, machine machine.Client, logger Logger, telemetry Telemetry) http.Handler {
	handler := NewHandler(config, machine, logger, telemetry)

	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		var startResult string
		data, err := verifyRequestAndReadBody(w, r, http.MethodGet, http.MethodPost)
		if err != nil {
			return
		}
		if len(data) > 0 {
			startResult = handler.Start(data)
		} else {
			startResult = handler.Start(nil)
		}
		sendResponse(w, startResult)
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodGet, http.MethodPost) {
			return
		}
		stopResult := handler.Stop()
		sendResponse(w, stopResult)
	})

	mux.HandleFunc("/poweroff", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodPost) {
			return
		}
		stopResult := handler.PowerOff()
		sendResponse(w, stopResult)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodGet) {
			return
		}
		status := handler.Status()
		sendResponse(w, status)
	})

	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodGet, http.MethodDelete) {
			return
		}
		deleteResult := handler.Delete()
		sendResponse(w, deleteResult)
	})

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodGet) {
			return
		}
		versionResult := handler.GetVersion()
		sendResponse(w, versionResult)
	})

	mux.HandleFunc("/webconsoleurl", func(w http.ResponseWriter, r *http.Request) {
		if wrongHTTPMethodUsed(r, w, http.MethodGet) {
			return
		}
		webconsoleInfo := handler.GetWebconsoleInfo()
		sendResponse(w, webconsoleInfo)
	})

	mux.HandleFunc("/config", configHandler(handler))

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		sendResponse(w, handler.Logs())
	})

	mux.HandleFunc("/telemetry", func(w http.ResponseWriter, r *http.Request) {
		data, err := verifyRequestAndReadBody(w, r, http.MethodGet, http.MethodPost)
		if err != nil {
			return
		}
		sendResponse(w, handler.UploadTelemetry(data))
	})

	mux.HandleFunc("/pull-secret", pullSecretHandler(config))

	return mux
}

func pullSecretHandler(config crcConfig.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			bin, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := cluster.StoreInKeyring(string(bin)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		if _, err := cluster.NewNonInteractivePullSecretLoader(config, "").Value(); err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func configHandler(handler *Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var configGetResult string
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if len(data) > 0 {
				configGetResult = handler.GetConfig(data)
			} else {
				configGetResult = handler.GetConfig(nil)
			}
			sendResponse(w, configGetResult)
			return
		}
		if r.Method == http.MethodPost {
			handleConfigSetUnset(w, r, func(data []byte) string { return handler.SetConfig(data) })
			return
		}
		if r.Method == http.MethodDelete {
			handleConfigSetUnset(w, r, func(data []byte) string { return handler.UnsetConfig(data) })
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func wrongHTTPMethodUsed(r *http.Request, w http.ResponseWriter, allowedMethods ...string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return false
		}
	}
	logging.Debugf("Wrong method: %s used for request to: %s", r.Method, r.RequestURI)
	http.Error(w, fmt.Sprintf("Only %s is allowed.", strings.Join(allowedMethods, ",")), http.StatusMethodNotAllowed)
	return true
}

func sendResponse(w http.ResponseWriter, response string) {
	var result client.Result
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, err.Error()); err != nil {
			logging.Error("Failed to send response: ", err)
		}
		return
	}

	if result.Success {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, response); err != nil {
			logging.Error("Failed to send response: ", err)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, result.Error); err != nil {
			logging.Error("Failed to send response: ", err)
		}
	}
}

func handleConfigSetUnset(w http.ResponseWriter, r *http.Request, fn func([]byte) string) {
	data, err := verifyRequestAndReadBody(w, r, http.MethodPost, http.MethodDelete)
	if err != nil {
		return
	}
	if len(data) < 1 {
		http.Error(w, "Not enough arguments for config (un)set", http.StatusBadRequest)
		return
	}

	result := fn(data)
	sendResponse(w, result)
}

func verifyRequestAndReadBody(w http.ResponseWriter, r *http.Request, method ...string) ([]byte, error) {
	if wrongHTTPMethodUsed(r, w, method...) {
		return nil, fmt.Errorf("Wrong HTTP method used")
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Debugf("Unable to read request body for URI: %s", r.RequestURI)
		http.Error(w, "Unable to read request body", http.StatusInternalServerError)
		return nil, err
	}
	return data, nil
}

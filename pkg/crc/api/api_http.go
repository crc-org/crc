package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func NewMux(config crcConfig.Storage, machine machine.Client, logger Logger) http.Handler {
	handler := NewHandler(config, machine, logger)

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
		if wrongHTTPMethodUsed(r, w, http.MethodGet) {
			return
		}
		stopResult := handler.Stop()
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
		if wrongHTTPMethodUsed(r, w, http.MethodGet) {
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

	mux.Handle("/config/", http.StripPrefix("/config", newConfigMux(handler)))

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		sendResponse(w, handler.Logs())
	})

	return mux
}

func newConfigMux(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		var configGetResult string
		data, err := verifyRequestAndReadBody(w, r, http.MethodGet, http.MethodPost)
		if err != nil {
			return
		}
		if len(data) > 0 {
			configGetResult = handler.GetConfig(data)
		} else {
			configGetResult = handler.GetConfig(nil)
		}
		sendResponse(w, configGetResult)
	})

	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		handleConfigSetUnset(w, r, func(data []byte) string { return handler.SetConfig(data) })
	})

	mux.HandleFunc("/unset", func(w http.ResponseWriter, r *http.Request) {
		handleConfigSetUnset(w, r, func(data []byte) string { return handler.UnsetConfig(data) })
	})

	return mux
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
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.WriteString(w, response); err != nil {
		logging.Error("Failed to send response: ", err)
	}
}

func handleConfigSetUnset(w http.ResponseWriter, r *http.Request, fn func([]byte) string) {
	data, err := verifyRequestAndReadBody(w, r, http.MethodPost)
	if err != nil {
		return
	}
	if len(data) < 1 {
		http.Error(w, "Not enough arguments for /config/(un)set", http.StatusBadRequest)
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

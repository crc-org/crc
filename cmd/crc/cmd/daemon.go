package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	"github.com/crc-org/crc/v2/pkg/crc/api"
	"github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/api/events"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/fileserver/fs9p"
	"github.com/gorilla/handlers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	watchdog              bool
	daemonVersionSupplier func() (client.VersionResult, error)
)

func init() {
	daemonVersionSupplier = func() (client.VersionResult, error) {
		return daemonclient.New().APIClient.Version()
	}
	daemonCmd.Flags().BoolVar(&watchdog, "watchdog", false, "Monitor stdin and shutdown the daemon if stdin is closed")
	rootCmd.AddCommand(daemonCmd)
}

const ErrDaemonAlreadyRunning = "daemon has been started in the background"

func checkDaemonVersion() (bool, error) {
	if _, err := daemonVersionSupplier(); err == nil {
		return true, errors.New(ErrDaemonAlreadyRunning)
	}
	return false, nil
}

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		if running, _ := checkIfDaemonIsRunning(); running {
			return errors.New(ErrDaemonAlreadyRunning)
		}

		return run(config)
	},
}

func run(cfg *crcConfig.Config) error {
	errCh := make(chan error)

	// Main HTTP listener for /api and /events endpoints
	listener, err := httpListener()
	if err != nil {
		return err
	}

	go func() {
		if listener == nil {
			return
		}
		mux := http.NewServeMux()
		machineClient := newMachine()
		mux.Handle("/api/", interceptResponseBodyMiddleware(http.StripPrefix("/api", api.NewMux(config, machineClient, logging.Memory, segmentClient)), logResponseBodyConditionally))
		mux.Handle("/events", interceptResponseBodyMiddleware(http.StripPrefix("/events", events.NewEventServer(machineClient)), logResponseBodyConditionally))
		s := &http.Server{
			Handler:           handlers.LoggingHandler(os.Stderr, mux),
			ReadHeaderTimeout: 10 * time.Second,
		}
		if err := s.Serve(listener); err != nil {
			errCh <- errors.Wrap(err, "api http.Serve failed")
		}
	}()

	// Admin helper listener for /hosts endpoints (separate socket)
	adminHelperLn, err := adminHelperListener()
	if err != nil {
		return err
	}

	go func() {
		if adminHelperLn == nil {
			return
		}
		mux := gatewayAPIMux(cfg, adminHelperHostsFileEditor{})
		s := &http.Server{
			Handler:           handlers.LoggingHandler(os.Stderr, mux),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
		}
		if err := s.Serve(adminHelperLn); err != nil {
			errCh <- errors.Wrap(err, "admin helper http.Serve failed")
		}
	}()

	// 9p home directory sharing (Windows only)
	if runtime.GOOS == "windows" && cfg.Get(crcConfig.EnableSharedDirs).AsBool() {
		// 9p over hvsock
		listener9pHvsock, err := fs9p.GetHvsockListener(constants.Plan9HvsockGUID)
		if err != nil {
			return err
		}
		server9pHvsock, err := fs9p.New9pServer(listener9pHvsock, constants.GetHomeDir())
		if err != nil {
			return err
		}
		if err := server9pHvsock.Start(); err != nil {
			return err
		}
		defer func() {
			if err := server9pHvsock.Stop(); err != nil {
				logging.Warnf("error stopping 9p server (hvsock): %v", err)
			}
		}()
		go func() {
			if err := server9pHvsock.WaitForError(); err != nil {
				logging.Errorf("9p server (hvsock) error: %v", err)
			}
		}()

		// 9p over TCP (as a backup)
		listener9pTCP, err := net.Listen("tcp", fmt.Sprintf("%s:%d", constants.VSockGateway, constants.Plan9TcpPort))
		if err != nil {
			return err
		}
		server9pTCP, err := fs9p.New9pServer(listener9pTCP, constants.GetHomeDir())
		if err != nil {
			return err
		}
		if err := server9pTCP.Start(); err != nil {
			return err
		}
		defer func() {
			if err := server9pTCP.Stop(); err != nil {
				logging.Warnf("error stopping 9p server (tcp): %v", err)
			}
		}()
		go func() {
			if err := server9pTCP.WaitForError(); err != nil {
				logging.Errorf("9p server (tcp) error: %v", err)
			}
		}()
	}

	startupDone()

	c := make(chan os.Signal, 1)

	if watchdog {
		go func() {
			if _, err := io.ReadAll(os.Stdin); err != nil {
				logging.Errorf("unexpected error while reading stdin: %v", err)
			}
			logging.Error("stdin is closed, shutdown...")
			c <- syscall.SIGTERM
		}()
	}

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		return nil
	case err := <-errCh:
		return err
	}
}

type HostsFileEditor interface {
	Add(ip string, hostnames ...string) error
	Remove(hostnames ...string) error
}

type adminHelperHostsFileEditor struct{}

func (adminHelperHostsFileEditor) Add(ip string, hostnames ...string) error {
	return adminhelper.AddToHostsFile(ip, hostnames...)
}

func (adminHelperHostsFileEditor) Remove(hostnames ...string) error {
	return adminhelper.RemoveFromHostsFile(hostnames...)
}

// gatewayAPIMux creates the HTTP mux for the admin helper hosts file API.
// This API allows adding and removing entries from the hosts file.
func gatewayAPIMux(cfg *crcConfig.Config, hostsEditor HostsFileEditor) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hosts/add", func(w http.ResponseWriter, r *http.Request) {
		acceptJSONStringArray(w, r, func(hostnames []string) error {
			if !cfg.Get(crcConfig.ModifyHostsFile).AsBool() {
				logging.Infof("Skipping hosts file modification because 'modify-hosts-file' is set to false")

				return nil
			}
			return hostsEditor.Add("127.0.0.1", hostnames...)
		})
	})
	mux.HandleFunc("/hosts/remove", func(w http.ResponseWriter, r *http.Request) {
		acceptJSONStringArray(w, r, func(hostnames []string) error {
			if !cfg.Get(crcConfig.ModifyHostsFile).AsBool() {
				logging.Infof("Skipping hosts file modification because 'modify-hosts-file' is set to false")

				return nil
			}
			return hostsEditor.Remove(hostnames...)
		})
	})
	return mux
}

func acceptJSONStringArray(w http.ResponseWriter, r *http.Request, fun func(hostnames []string) error) {
	if r.Method != http.MethodPost {
		http.Error(w, "post only", http.StatusBadRequest)
		return
	}
	var req []string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := fun(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func logResponseBodyConditionally(statusCode int, buffer *bytes.Buffer, r *http.Request) {
	responseBody := buffer.String()
	if statusCode != http.StatusOK && responseBody != "" {
		log.Errorf("[%s] \"%s %s\" Response Body: %s\n", time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method, r.URL.Path, buffer.String())
	}
}

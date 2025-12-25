package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	"github.com/crc-org/crc/v2/pkg/crc/api"
	"github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/api/events"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/fileserver/fs9p"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/docker/go-units"
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

const (
	hostVirtualIP           = "192.168.127.254"
	ErrDaemonAlreadyRunning = "daemon has been started in the background"
)

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

		virtualNetworkConfig := createNewVirtualNetworkConfig(config)
		err := run(&virtualNetworkConfig)
		return err
	},
}

func createNewVirtualNetworkConfig(providedConfig *crcConfig.Config) types.Configuration {
	virtualNetworkConfig := types.Configuration{
		Debug:             false, // never log packets
		CaptureFile:       os.Getenv("CRC_DAEMON_PCAP_FILE"),
		MTU:               4000, // Large packets slightly improve the performance. Less small packets.
		Subnet:            "192.168.127.0/24",
		GatewayIP:         constants.VSockGateway,
		GatewayMacAddress: "5a:94:ef:e4:0c:dd",
		DHCPStaticLeases: map[string]string{
			"192.168.127.2": constants.VsockMacAddress,
		},
		DNS: []types.Zone{
			{
				Name:      "apps-crc.testing.",
				DefaultIP: net.ParseIP("192.168.127.2"),
			},
			{
				Name: "crc.testing.",
				Records: []types.Record{
					{
						Name: "host",
						IP:   net.ParseIP(hostVirtualIP),
					},
					{
						Name: "gateway",
						IP:   net.ParseIP("192.168.127.1"),
					},
					{
						Name: "api",
						IP:   net.ParseIP("192.168.127.2"),
					},
					{
						Name: "api-int",
						IP:   net.ParseIP("192.168.127.2"),
					},
					{
						Regexp: regexp.MustCompile("crc-(.*?)-master-0"),
						IP:     net.ParseIP("192.168.126.11"),
					},
				},
			},
			{
				Name: "containers.internal.",
				Records: []types.Record{
					{
						Name: "gateway",
						IP:   net.ParseIP(hostVirtualIP),
					},
				},
			},
			{
				Name: "docker.internal.",
				Records: []types.Record{
					{
						Name: "gateway",
						IP:   net.ParseIP(hostVirtualIP),
					},
				},
			},
		},
		Protocol:          types.HyperKitProtocol,
		GatewayVirtualIPs: []string{hostVirtualIP},
	}
	if providedConfig.Get(crcConfig.HostNetworkAccess).AsBool() {
		log.Debugf("Enabling host network access")
		if virtualNetworkConfig.NAT == nil {
			virtualNetworkConfig.NAT = make(map[string]string)
		}
		virtualNetworkConfig.NAT[hostVirtualIP] = "127.0.0.1"
	}
	return virtualNetworkConfig
}

func run(configuration *types.Configuration) error {
	vn, err := virtualnetwork.New(configuration)
	if err != nil {
		return err
	}

	errCh := make(chan error)

	listener, err := httpListener()
	if err != nil {
		return err
	}

	go func() {
		if listener == nil {
			return
		}
		mux := http.NewServeMux()
		mux.Handle("/network/", interceptResponseBodyMiddleware(http.StripPrefix("/network", vn.Mux()), logResponseBodyConditionally))
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

	ln, err := vn.Listen("tcp", net.JoinHostPort(configuration.GatewayIP, "80"))
	if err != nil {
		return err
	}
	go func() {
		mux := gatewayAPIMux(config, adminHelperHostsFileEditor{})
		s := &http.Server{
			Handler:      handlers.LoggingHandler(os.Stderr, mux),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := s.Serve(ln); err != nil {
			errCh <- errors.Wrap(err, "gateway http.Serve failed")
		}
	}()

	networkListener, err := vn.Listen("tcp", net.JoinHostPort(hostVirtualIP, "80"))
	if err != nil {
		return err
	}
	go func() {
		mux := networkAPIMux(vn)
		s := &http.Server{
			Handler:      handlers.LoggingHandler(os.Stderr, mux),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := s.Serve(networkListener); err != nil {
			errCh <- errors.Wrap(err, "host virtual IP http.Serve failed")
		}
	}()

	go func() {
		var oldCancel context.CancelFunc
		for {
			ctx, cancel := context.WithCancel(context.Background())
			conn, err := unixgramListener(ctx, vn)
			if err != nil && errors.Is(err, drivers.ErrNotImplemented) {
				cancel()
				break
			}
			if err != nil && !errors.Is(err, net.ErrClosed) {
				logging.Errorf("unixgramListener error: %v", err)
			}

			if oldCancel != nil {
				logging.Warnf("New connection to %s. Closing old connection", conn.LocalAddr().String())
				oldCancel()
			}
			oldCancel = cancel
			time.Sleep(1 * time.Second)
		}
	}()

	vsockListener, err := vsockListener()
	if err != nil {
		return err
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle(types.ConnectPath, vn.Mux())
		s := &http.Server{
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		if err := s.Serve(vsockListener); err != nil {
			errCh <- errors.Wrap(err, "virtualnetwork http.Serve failed")
		}
	}()

	// 9p home directory sharing
	if runtime.GOOS == "windows" && config.Get(crcConfig.EnableSharedDirs).AsBool() {
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
		listener9pTCP, err := vn.Listen("tcp", net.JoinHostPort(configuration.GatewayIP, fmt.Sprintf("%d", constants.Plan9TcpPort)))
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

	if logging.IsDebug() {
		go func() {
			for {
				fmt.Printf("%v sent to the VM, %v received from the VM\n", units.HumanSize(float64(vn.BytesSent())), units.HumanSize(float64(vn.BytesReceived())))
				time.Sleep(5 * time.Second)
			}
		}()
	}

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

// This API is only exposed in the virtual network (only the VM can reach this).
// Any process inside the VM can reach it by connecting to gateway.crc.testing:80.
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

func networkAPIMux(vn *virtualnetwork.VirtualNetwork) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", vn.Mux())
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

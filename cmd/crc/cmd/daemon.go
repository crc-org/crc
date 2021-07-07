package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/api"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/code-ready/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/docker/go-units"
	"github.com/gorilla/handlers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/exec"
)

var watchdog bool

func init() {
	daemonCmd.Flags().BoolVar(&watchdog, "watchdog", false, "Monitor stdin and shutdown the daemon if stdin is closed")
	rootCmd.AddCommand(daemonCmd)
}

const hostVirtualIP = "192.168.127.254"

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := daemonclient.New().APIClient.Version(); err == nil {
			return errors.New("daemon is already running")
		}

		if err := preflight.StartPreflightChecks(config); err != nil {
			return exec.CodeExitError{
				Err:  err,
				Code: preflightFailedExitCode,
			}
		}

		virtualNetworkConfig := types.Configuration{
			Debug:             false, // never log packets
			CaptureFile:       os.Getenv("CRC_DAEMON_PCAP_FILE"),
			MTU:               4000, // Large packets slightly improve the performance. Less small packets.
			Subnet:            "192.168.127.0/24",
			GatewayIP:         constants.VSockGateway,
			GatewayMacAddress: "\x5A\x94\xEF\xE4\x0C\xDD",
			DNS: []types.Zone{
				{
					Name:      "apps-crc.testing.",
					DefaultIP: net.ParseIP("192.168.127.2"),
				},
				{
					Name: "crc.testing.",
					Records: []types.Record{
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
			},
		}
		if config.Get(crcConfig.HostNetworkAccess).AsBool() {
			log.Debugf("Enabling host network access")
			for i := range virtualNetworkConfig.DNS {
				zone := &virtualNetworkConfig.DNS[i]
				if zone.Name != "crc.testing." {
					continue
				}
				log.Debugf("Adding \"host\" -> %s DNS record to crc.testing. zone", hostVirtualIP)

				zone.Records = append(zone.Records, types.Record{Name: "host", IP: net.ParseIP(hostVirtualIP)})
			}

			if virtualNetworkConfig.NAT == nil {
				virtualNetworkConfig.NAT = make(map[string]string)
			}
			virtualNetworkConfig.NAT[hostVirtualIP] = "127.0.0.1"
		}
		err := run(&virtualNetworkConfig)
		return err
	},
}

func run(configuration *types.Configuration) error {
	vsockListener, err := vsockListener()
	if err != nil {
		return err
	}

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
		mux.Handle("/network/", http.StripPrefix("/network", vn.Mux()))
		mux.Handle("/api/", http.StripPrefix("/api", api.NewMux(config, newMachine(), logging.Memory, segmentClient)))
		if err := http.Serve(listener, handlers.LoggingHandler(os.Stderr, mux)); err != nil {
			errCh <- errors.Wrap(err, "api http.Serve failed")
		}
	}()

	ln, err := vn.Listen("tcp", fmt.Sprintf("%s:80", configuration.GatewayIP))
	if err != nil {
		return err
	}
	go func() {
		mux := gatewayAPIMux()
		if err := http.Serve(ln, handlers.LoggingHandler(os.Stderr, mux)); err != nil {
			errCh <- errors.Wrap(err, "gateway http.Serve failed")
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle(types.ConnectPath, vn.Mux())
		if err := http.Serve(vsockListener, mux); err != nil {
			errCh <- errors.Wrap(err, "virtualnetwork http.Serve failed")
		}
	}()

	if isDebugLog() {
		go func() {
			for {
				fmt.Printf("%v sent to the VM, %v received from the VM\n", units.HumanSize(float64(vn.BytesSent())), units.HumanSize(float64(vn.BytesReceived())))
				time.Sleep(5 * time.Second)
			}
		}()
	}

	go func() {
		if err := runBinaryAPI(); err != nil {
			errCh <- err
		}
	}()

	c := make(chan os.Signal, 1)

	if watchdog {
		go func() {
			if _, err := ioutil.ReadAll(os.Stdin); err != nil {
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

// This API is only exposed in the virtual network (only the VM can reach this).
// Any process inside the VM can reach it by connecting to gateway.crc.testing:80.
func gatewayAPIMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hosts/add", func(w http.ResponseWriter, r *http.Request) {
		acceptJSONStringArray(w, r, func(hostnames []string) error {
			return adminhelper.AddToHostsFile("127.0.0.1", hostnames...)
		})
	})
	mux.HandleFunc("/hosts/remove", func(w http.ResponseWriter, r *http.Request) {
		acceptJSONStringArray(w, r, func(hostnames []string) error {
			return adminhelper.RemoveFromHostsFile(hostnames...)
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

func runBinaryAPI() error {
	if runtime.GOOS != "windows" { // only Windows tray requires the binary protocol
		return nil
	}
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	apiServer, err := api.CreateServer(constants.DaemonSocketPath, config, newMachine(), logging.Memory, segmentClient)
	if err != nil {
		return err
	}
	return apiServer.Serve()
}

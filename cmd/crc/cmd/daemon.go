package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/api"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/code-ready/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/docker/go-units"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

const hostVirtualIP = "192.168.127.254"

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		virtualNetworkConfig := types.Configuration{
			Debug:             false, // never log packets
			CaptureFile:       captureFile(),
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
			Forwards: map[string]string{
				fmt.Sprintf(":%d", constants.VsockSSHPort): "192.168.127.2:22",
				":6443": "192.168.127.2:6443",
				":443":  "192.168.127.2:443",
			},
		}
		if config.Get(cmdConfig.HostNetworkAccess).AsBool() {
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

func captureFile() string {
	if !isDebugLog() {
		return ""
	}
	return filepath.Join(constants.CrcBaseDir, "capture.pcap")
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
		if err := http.Serve(listener, mux); err != nil {
			errCh <- err
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle(types.ConnectPath, vn.Mux())
		if err := http.Serve(vsockListener, mux); err != nil {
			errCh <- err
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
		if err := runDaemon(); err != nil {
			errCh <- err
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		return nil
	case err := <-errCh:
		return err
	}
}

func runDaemon() error {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	apiServer, err := api.CreateServer(constants.DaemonSocketPath, config, newMachine())
	if err != nil {
		return err
	}
	return apiServer.Serve()
}

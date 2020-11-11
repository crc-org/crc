package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/goodhosts"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/gvisor-tap-vsock/pkg/transport"
	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/code-ready/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/docker/go-units"
	v1 "github.com/openshift/api/route/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"
	informers "github.com/openshift/client-go/route/informers/externalversions"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// setup separate logging for daemon
		logging.CloseLogging()
		logging.InitLogrus(logging.LogLevel, constants.DaemonLogFilePath)

		go runDaemon()

		go func() {
			client := newMachine()
			var ip string
			for {
				var err error
				ip, err = client.IP()
				if err == nil {
					break
				}
				logging.Info("VM doesn't exist yet, waiting IP 2sec.")
				time.Sleep(2 * time.Second)
			}
			logging.Infof("Found IP: %s", ip)
			kubeconfig := filepath.Join(constants.MachineInstanceDir, constants.DefaultName, "kubeconfig")
			for {
				if _, err := os.Stat(kubeconfig); err == nil {
					break
				}
				logging.Info("VM doesn't exist yet, waiting kubeconfig 2sec.")
				time.Sleep(2 * time.Second)
			}

			stopper := make(chan struct{})
			if err := routesController(stopper, kubeconfig, ip); err != nil {
				logging.Fatal(err)
			}
		}()

		var endpoints []string
		if runtime.GOOS == "windows" {
			endpoints = append(endpoints, transport.DefaultURL)
		} else {
			_ = os.Remove(constants.NetworkSocketPath)
			endpoints = append(endpoints, fmt.Sprintf("unix://%s", constants.NetworkSocketPath))
			if runtime.GOOS == "linux" {
				endpoints = append(endpoints, transport.DefaultURL)
			}
		}

		if err := run(&types.Configuration{
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
		}, endpoints); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func captureFile() string {
	if !isDebugLog() {
		return ""
	}
	return filepath.Join(constants.CrcBaseDir, "capture.pcap")
}

func run(configuration *types.Configuration, endpoints []string) error {
	vn, err := virtualnetwork.New(configuration)
	if err != nil {
		return err
	}
	log.Info("waiting for clients...")
	errCh := make(chan error)

	for _, endpoint := range endpoints {
		log.Infof("listening %s", endpoint)
		ln, err := transport.Listen(endpoint)
		if err != nil {
			return errors.Wrap(err, "cannot listen")
		}

		go func() {
			if err := http.Serve(ln, vn.Mux()); err != nil {
				errCh <- err
			}
		}()
	}
	if isDebugLog() {
		go func() {
			for {
				fmt.Printf("%v sent to the VM, %v received from the VM\n", units.HumanSize(float64(vn.BytesSent())), units.HumanSize(float64(vn.BytesReceived())))
				time.Sleep(5 * time.Second)
			}
		}()
	}
	return <-errCh
}

func newConfig() (crcConfig.Storage, error) {
	config, _, err := newViperConfig()
	return config, err
}

func routesController(stopCh chan struct{}, kubeconfig, ip string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := routeclientset.NewForConfig(config)
	if err != nil {
		return err
	}
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Route().V1().Routes().Informer()
	defer close(stopCh)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			route := obj.(*v1.Route)
			fmt.Printf("added: %s %s\n", route.GetName(), route.Spec.Host)
			if err := goodhosts.AddToHostsFile(ip, route.Spec.Host); err != nil {
				logging.Error(err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*v1.Route)
			route := newObj.(*v1.Route)
			fmt.Printf("updated: %s (%s -> %s)\n", route.GetName(), old.Spec.Host, route.Spec.Host)
			if err := goodhosts.RemoveFromHostsFile(ip, old.Spec.Host); err != nil {
				logging.Error(err)
			}
			if err := goodhosts.AddToHostsFile(ip, route.Spec.Host); err != nil {
				logging.Error(err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			route := obj.(*v1.Route)
			fmt.Printf("deleted: %s %s\n", route.GetName(), route.Spec.Host)
			if err := goodhosts.RemoveFromHostsFile(ip, route.Spec.Host); err != nil {
				logging.Error(err)
			}
		},
	})
	informer.Run(stopCh)
	return nil
}

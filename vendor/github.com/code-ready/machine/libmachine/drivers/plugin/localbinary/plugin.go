package localbinary

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// Timeout where we will bail if we're not able to properly contact the
	// plugin server.
	defaultTimeout = 10 * time.Second
)

const (
	pluginOut           = "(%s) %s"
	pluginErr           = "(%s) DBG | %s"
	PluginEnvKey        = "MACHINE_PLUGIN_TOKEN"
	PluginEnvVal        = "42"
	PluginEnvDriverName = "MACHINE_PLUGIN_DRIVER_NAME"
)

type McnBinaryExecutor interface {
	// Execute the driver plugin.  Returns scanners for plugin binary
	// stdout and stderr.
	Start() (*bufio.Scanner, *bufio.Scanner, error)

	// Stop reading from the plugins in question.
	Close() error
}

type Plugin struct {
	Executor    McnBinaryExecutor
	Addr        string
	MachineName string
	addrCh      chan string
	stopCh      chan struct{}
	timeout     time.Duration
}

type Executor struct {
	pluginStdout, pluginStderr io.ReadCloser
	DriverName                 string
	cmd                        *exec.Cmd
	binaryPath                 string
}

// driverPath finds the path of a driver binary by its name. The separate binary must be in the PATH and its name must be
// `crc-machine-driverName`
func driverPath(driverName string, binaryPath string) string {
	driverName = fmt.Sprintf("crc-driver-%s", driverName)
	if binaryPath != "" {
		return filepath.Join(binaryPath, driverName)
	}
	return driverName
}

func NewPlugin(driverName string, binaryPath string) (*Plugin, error) {
	driverPath := driverPath(driverName, binaryPath)
	binaryPath, err := exec.LookPath(driverPath)
	if err != nil {
		return nil, err
	}

	log.Debugf("Found binary path at %s", binaryPath)

	return &Plugin{
		stopCh: make(chan struct{}),
		addrCh: make(chan string, 1),
		Executor: &Executor{
			DriverName: driverName,
			binaryPath: binaryPath,
		},
	}, nil
}

func (lbe *Executor) Start() (*bufio.Scanner, *bufio.Scanner, error) {
	var err error

	log.Debugf("Launching plugin server for driver %s", lbe.DriverName)

	// #nosec G204
	lbe.cmd = exec.Command(lbe.binaryPath)

	lbe.pluginStdout, err = lbe.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stdout pipe: %s", err)
	}

	lbe.pluginStderr, err = lbe.cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stderr pipe: %s", err)
	}

	outScanner := bufio.NewScanner(lbe.pluginStdout)
	errScanner := bufio.NewScanner(lbe.pluginStderr)

	os.Setenv(PluginEnvKey, PluginEnvVal)
	os.Setenv(PluginEnvDriverName, lbe.DriverName)

	if err := lbe.cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("Error starting plugin binary: %s", err)
	}

	return outScanner, errScanner, nil
}

func (lbe *Executor) Close() error {
	if err := lbe.cmd.Wait(); err != nil {
		return fmt.Errorf("Error waiting for binary close: %s", err)
	}

	return nil
}

func stream(scanner *bufio.Scanner, streamOutCh chan<- string, stopCh <-chan struct{}) {
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Warnf("Scanning stream: %s", err)
		}
		select {
		case streamOutCh <- strings.Trim(line, "\n"):
		case <-stopCh:
			return
		}
	}
}

func (lbp *Plugin) AttachStream(scanner *bufio.Scanner) <-chan string {
	streamOutCh := make(chan string)
	go stream(scanner, streamOutCh, lbp.stopCh)
	return streamOutCh
}

func (lbp *Plugin) execServer() error {
	outScanner, errScanner, err := lbp.Executor.Start()
	if err != nil {
		return err
	}

	// Scan just one line to get the address, then send it to the relevant
	// channel.
	outScanner.Scan()
	addr := outScanner.Text()
	if err := outScanner.Err(); err != nil {
		return fmt.Errorf("Reading plugin address failed: %s", err)
	}

	lbp.addrCh <- strings.TrimSpace(addr)

	stdOutCh := lbp.AttachStream(outScanner)
	stdErrCh := lbp.AttachStream(errScanner)

	for {
		select {
		case out := <-stdOutCh:
			log.Infof(pluginOut, lbp.MachineName, out)
		case err := <-stdErrCh:
			log.Debugf(pluginErr, lbp.MachineName, err)
		case <-lbp.stopCh:
			if err := lbp.Executor.Close(); err != nil {
				return fmt.Errorf("Error closing local plugin binary: %s", err)
			}
			return nil
		}
	}
}

func (lbp *Plugin) Serve() error {
	return lbp.execServer()
}

func (lbp *Plugin) Address() (string, error) {
	if lbp.Addr == "" {
		if lbp.timeout == 0 {
			lbp.timeout = defaultTimeout
		}

		select {
		case lbp.Addr = <-lbp.addrCh:
			log.Debugf("Plugin server listening at address %s", lbp.Addr)
			close(lbp.addrCh)
			return lbp.Addr, nil
		case <-time.After(lbp.timeout):
			return "", fmt.Errorf("Failed to dial the plugin server in %s", lbp.timeout)
		}
	}
	return lbp.Addr, nil
}

func (lbp *Plugin) Close() error {
	close(lbp.stopCh)
	return nil
}

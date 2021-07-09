package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/gvisor-tap-vsock/pkg/transport"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
	"github.com/mdlayher/vsock"
)

const (
	vsockUnitName = "crc-vsock.socket"
	httpUnitName  = "crc-http.socket"
)

var systemdListeners map[string][]net.Listener

func init() {
	// listenerWithNames() cannot be called multiple times
	systemdListeners, _ = listenersWithNames()

	checkIfDaemonIsRunning = func() (bool, error) {
		ln, _ := getSystemdListener(httpUnitName)
		if ln != nil {
			/* detect if the daemon is being started by systemd,
			* and socket activation is in use. In this scenario,
			* trying to send an HTTP version check on the daemon
			* HTTP socket would hang as the socket is listening for
			* connections but is not setup to handle them yet.
			 */
			return false, nil
		}

		return checkDaemonVersion()
	}
}

// listenersWithNames maps a listener name to a set of net.Listener instances.
// This is the same code as https://github.com/coreos/go-systemd/blob/main/activation/listeners.go
// with support for vsock
//
// This function can only be called once, subsequent calls will always return an empty list
func listenersWithNames() (map[string][]net.Listener, error) {
	files := activation.Files(true)
	listeners := map[string][]net.Listener{}

	for _, f := range files {
		pc, err := net.FileListener(f)
		if err != nil {
			logging.Infof("got error %t %v", err, err)
			// net.FileListener does not support vsock, need to fallback to vsock-specific code
			pc, err = vsock.FileListener(f)
			if err != nil {
				logging.Debugf("failed to create listener for %s: %v", f.Name(), err)
				continue
			}
		}
		current, ok := listeners[f.Name()]
		if !ok {
			listeners[f.Name()] = []net.Listener{pc}
		} else {
			listeners[f.Name()] = append(current, pc)
		}
		f.Close()
	}
	return listeners, nil
}

func getSystemdListener(unitName string) (net.Listener, error) {
	listeners, found := systemdListeners[unitName]
	if !found {
		logging.Debugf("no listener for %s: %+v", unitName, systemdListeners)
		// it's not an error if systemd provided no listener
		return nil, nil
	}
	if len(listeners) != 1 {
		logging.Debugf("unexpected number of sockets for %s (%d != 1)", unitName, len(listeners))
		return nil, fmt.Errorf("unexpected number of sockets for %s (%d != 1)", unitName, len(listeners))
	}

	return listeners[0], nil
}

func vsockListener() (net.Listener, error) {
	ln, err := getSystemdListener(vsockUnitName)
	if err != nil {
		return nil, err
	}
	if ln != nil {
		logging.Infof("using socket provided by %s", vsockUnitName)
		return ln, nil
	}

	// no socket activation, we need to create the listener
	ln, err = transport.Listen(transport.DefaultURL)
	logging.Infof("listening %s", transport.DefaultURL)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func httpListener() (net.Listener, error) {
	// check for systemd socket-activation
	ln, err := getSystemdListener(httpUnitName)
	if err != nil {
		return nil, err
	}
	if ln != nil {
		logging.Infof("using socket provided by %s", httpUnitName)
		return ln, nil
	}

	// no socket activation, we need to create the listener
	_ = os.Remove(constants.DaemonHTTPSocketPath)
	ln, err = net.Listen("unix", constants.DaemonHTTPSocketPath)
	logging.Infof("listening %s", constants.DaemonHTTPSocketPath)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func startupDone() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
}

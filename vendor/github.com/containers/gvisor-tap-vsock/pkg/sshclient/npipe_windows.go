package sshclient

import (
	"fmt"
	"net"
	"net/url"
	"os/user"
	"strings"

	winio "github.com/Microsoft/go-winio"
	"github.com/sirupsen/logrus"
)

// https://docs.microsoft.com/en-us/windows-hardware/drivers/kernel/sddl-for-device-objects
// Allow built-in admins and system/kernel components
const SddlDevObjSysAllAdmAll = "D:P(A;;GA;;;SY)(A;;GA;;;BA)"

func ListenNpipe(socketURI *url.URL) (net.Listener, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	// Also allow current user
	sddl := fmt.Sprintf("%s(A;;GA;;;%s)", SddlDevObjSysAllAdmAll, user.Uid)
	config := winio.PipeConfig{
		SecurityDescriptor: sddl,
		MessageMode:        true,
		InputBufferSize:    65536,
		OutputBufferSize:   65536,
	}
	path := strings.ReplaceAll(socketURI.Path, "/", "\\")

	listener, err := winio.ListenPipe(path, &config)
	if err != nil {
		return listener, fmt.Errorf("error listening on socket: %s: %w", socketURI, err)
	}

	logrus.Info("Listening on: " + path)

	return listener, nil
}

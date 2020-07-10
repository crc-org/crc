package api

import (
	"fmt"
	"net"
	"os/user"

	"github.com/Microsoft/go-winio"
	"github.com/code-ready/crc/pkg/crc/logging"
)

//createIPCListener returns a listener of windows named pipes
func createIPCListener(socketPath string) (net.Listener, error) {
	/* the named pipe is  configured that built in users are able
	 * to read,write from the pipe, info about security descriptor
	 * can be found at:
	 *	- https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-string-format
	 *	- https://itconnect.uw.edu/wares/msinf/other-help/understanding-sddl-syntax/
	 */
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	sid, err := winio.LookupSidByName(u.Username)
	if err != nil {
		return nil, err
	}
	pipeConfig := winio.PipeConfig{
		SecurityDescriptor: fmt.Sprintf("D:P(A;;GA;;;%s)", sid),
	}
	listener, err := winio.ListenPipe("\\\\.\\pipe\\crc", &pipeConfig)
	if err != nil {
		logging.Error("Failed to create named pipe 'crc': ", err.Error())
		return nil, err
	}

	return listener, err
}

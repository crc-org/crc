package api

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/config"

	"github.com/code-ready/crc/pkg/crc/constants"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type crcDaemonService struct {
	socketPath string
	config     config.Storage
}

func (m *crcDaemonService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	crcAPIServer, err := CreateAPIServer(m.socketPath, m.config)
	if err != nil {
		elog.Error(1, fmt.Sprintf("Failed to start CRC daemon service: %v", err)) // nolint
		return
	}
	// Windows servivce manager calls OnStart when starting the service
	// and it calls this method, which the service manager expects to handle
	// commands like stop, shutdown. crcAPIServer.Serve is being called as a go routine
	// since we don't want it block
	go crcAPIServer.Serve()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	MainLoop(r, changes)

	// Above MainLoop will return when the daemon service is stopped
	changes <- svc.Status{State: svc.Stopped}
	return
}

func MainLoop(r <-chan svc.ChangeRequest, changes chan<- svc.Status) {
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			_ = elog.Info(1, "Shutting down CRC daemon service")
			return
		default:
			_ = elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
		}
	}
}

func RunCrcDaemonService(name string, isDebug bool, config config.Storage) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name)) // nolint
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &crcDaemonService{socketPath: constants.DaemonSocketPath, config: config})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err)) // nolint
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name)) // nolint
}

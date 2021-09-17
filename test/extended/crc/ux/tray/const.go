package tray

const (
	actionStart  string = "start"
	actionStop   string = "stop"
	actionDelete string = "delete"
	actionQuit   string = "quit"

	fieldState string = "state"

	stateRunning string = "Running"
	stateStopped string = "Stopped"

	userKubeadmin string = "kubeadmin"
	userDeveloper string = "developer"

	trayClusterStateRetries int = 15
	trayClusterStateTimeout int = 90
)

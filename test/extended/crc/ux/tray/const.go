package tray

const (
	actionStart  string = "start"
	actionStop   string = "stop"
	actionDelete string = "delete"
	actionQuit   string = "quit"

	fieldState string = "state"

	InstanceStateRunning string = "Running"
	InstanceStateStopped string = "Stopped"

	userKubeadmin string = "kubeadmin"
	userDeveloper string = "developer"

	trayClusterStateRetries int = 15
	trayClusterStateTimeout int = 90
)

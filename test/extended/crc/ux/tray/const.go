// +build !windows

package tray

const (
	actionStart  string = "start"
	actionStop   string = "stop"
	actionDelete string = "delete"
	actionQuit   string = "quit"

	fieldState string = "state"

	stateInitial string = "Machine doesn't exist"
	stateRunning string = "Running"
	stateStopped string = "Stopped"

	userKubeadmin string = "kubeadmin"
	userDeveloper string = "developer"
)

const (
	uxCheckAccessibilityDuration = "2s"
	uxCheckAccessibilityRetry    = 10
)

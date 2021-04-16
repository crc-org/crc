// +build !windows

package tray

// TODO check if move to tray package
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
)

const (
	uxCheckAccessibilityDuration = "2s"
	uxCheckAccessibilityRetry    = 10
)

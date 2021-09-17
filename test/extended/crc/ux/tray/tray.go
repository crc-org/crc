package tray

import (
	"fmt"
)

type Tray interface {
	Install() error
	IsInstalled() error
	IsAccessible() error
	ClickStart() error
	ClickStop() error
	ClickDelete() error
	ClickQuit() error
	SetPullSecret() error
	IsClusterRunning() error
	IsClusterStopped() error
	CopyOCLoginCommandAsKubeadmin() error
	CopyOCLoginCommandAsDeveloper() error
	// TODO check if make sense create a new ux component
	ConnectClusterAsKubeadmin() error
	ConnectClusterAsDeveloper() error
}

func getElement(name string, elements map[string]string) (string, error) {
	identifier, ok := elements[name]
	if ok {
		return identifier, nil
	}
	return "", fmt.Errorf("element '%s', Can not be accessed from the tray", name)
}

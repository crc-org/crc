package tray

type Tray interface {
	Install() error
	IsInstalled() error
	IsAccessible() error
	ClickStart() error
	ClickStop() error
	ClickDelete() error
	ClickQuit() error
	SetPullSecret() error
	IsInitialStatus() error
	IsClusterRunning() error
	IsClusterStopped() error
	CopyOCLoginCommandAsKubeadmin() error
	CopyOCLoginCommandAsDeveloper() error
	// TODO check if make sense create a new ux component
	ConnectClusterAsKubeadmin() error
	ConnectClusterAsDeveloper() error
}

type Element struct {
	Name         string
	AXIdentifier string
}

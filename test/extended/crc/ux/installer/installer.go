package installer

type Installer interface {
	Install() error
	RebootRequired() error
}

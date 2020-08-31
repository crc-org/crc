package actions

type Action int

const (
	Start Action = iota
	Stop
	Reload
	Restart
	Enable
	Disable
	Status
	DaemonReload
)

func (action Action) String() string {
	actions := [...]string{
		"start",
		"stop",
		"reload",
		"restart",
		"enable",
		"disable",
		"status",
		"daemon-reload",
	}

	if int(action) >= 0 && int(action) < len(actions) {
		return actions[action]
	}

	return ""
}

func (action Action) IsPriviledged() bool {
	switch action {
	case Status:
		return false
	case Start, Stop, Reload, Restart, Enable, Disable, DaemonReload:
		return true
	default:
		/* This should not be reached */
		return false
	}
}

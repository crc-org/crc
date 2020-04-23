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
	IsActive
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
		"is-active",
		"daemon-reload",
	}

	if int(action) >= 0 && int(action) < len(actions) {
		return actions[action]
	}

	return ""
}

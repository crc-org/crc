package api

import "github.com/code-ready/crc/pkg/crc/machine"

type AdaptedClient interface {
	Delete(deleteConfig machine.DeleteConfig) Result
	GetConsoleURL(consoleConfig machine.ConsoleConfig) (machine.ConsoleResult, error)
	Start(startConfig machine.StartConfig) (machine.StartResult, error)
	Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error)
	Stop(stopConfig machine.StopConfig) (machine.StopResult, error)
}

type Result struct {
	Name    string
	Success bool
	Error   string
}

type Adapter struct {
	Underlying machine.Client
}

func (a *Adapter) Delete(deleteConfig machine.DeleteConfig) Result {
	err := a.Underlying.Delete(deleteConfig)
	if err != nil {
		return Result{
			Name:    deleteConfig.Name,
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    deleteConfig.Name,
		Success: true,
	}
}

func (a *Adapter) GetConsoleURL(consoleConfig machine.ConsoleConfig) (machine.ConsoleResult, error) {
	return a.Underlying.GetConsoleURL(consoleConfig)
}

func (a *Adapter) Start(startConfig machine.StartConfig) (machine.StartResult, error) {
	return a.Underlying.Start(startConfig)
}

func (a *Adapter) Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error) {
	return a.Underlying.Status(statusConfig)
}

func (a *Adapter) Stop(stopConfig machine.StopConfig) (machine.StopResult, error) {
	return a.Underlying.Stop(stopConfig)
}

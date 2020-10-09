package api

import (
	"github.com/code-ready/crc/pkg/crc/machine"
)

type AdaptedClient interface {
	Delete(deleteConfig machine.DeleteConfig) Result
	GetConsoleURL(consoleConfig machine.ConsoleConfig) (machine.ConsoleResult, error)
	Start(startConfig machine.StartConfig) StartResult
	Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error)
	Stop(stopConfig machine.StopConfig) Result
}

type Result struct {
	Name    string
	Success bool
	Error   string
}

type StartResult struct {
	Name           string
	Status         string
	Error          string
	ClusterConfig  machine.ClusterConfig
	KubeletStarted bool
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

func (a *Adapter) Start(startConfig machine.StartConfig) StartResult {
	res, err := a.Underlying.Start(startConfig)
	if err != nil {
		return StartResult{
			Name:  startConfig.Name,
			Error: err.Error(),
		}
	}
	return StartResult{
		Name:           startConfig.Name,
		Status:         res.Status,
		ClusterConfig:  res.ClusterConfig,
		KubeletStarted: res.KubeletStarted,
	}
}

func (a *Adapter) Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error) {
	return a.Underlying.Status(statusConfig)
}

func (a *Adapter) Stop(stopConfig machine.StopConfig) Result {
	_, err := a.Underlying.Stop(stopConfig)
	if err != nil {
		return Result{
			Name:    stopConfig.Name,
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    stopConfig.Name,
		Success: true,
	}
}

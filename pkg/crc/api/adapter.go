package api

import (
	"github.com/code-ready/crc/pkg/crc/machine"
)

type AdaptedClient interface {
	Delete(deleteConfig machine.DeleteConfig) Result
	GetConsoleURL(consoleConfig machine.ConsoleConfig) ConsoleResult
	Start(startConfig machine.StartConfig) StartResult
	Status(statusConfig machine.ClusterStatusConfig) ClusterStatusResult
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

type ClusterStatusResult struct {
	Name             string
	CrcStatus        string
	OpenshiftStatus  string
	OpenshiftVersion string
	DiskUse          int64
	DiskSize         int64
	Error            string
	Success          bool
}

type ConsoleResult struct {
	ClusterConfig machine.ClusterConfig
	Success       bool
	Error         string
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

func (a *Adapter) GetConsoleURL(consoleConfig machine.ConsoleConfig) ConsoleResult {
	res, err := a.Underlying.GetConsoleURL(consoleConfig)
	if err != nil {
		return ConsoleResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	return ConsoleResult{
		ClusterConfig: res.ClusterConfig,
		Success:       true,
	}
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

func (a *Adapter) Status(statusConfig machine.ClusterStatusConfig) ClusterStatusResult {
	res, err := a.Underlying.Status(statusConfig)
	if err != nil {
		return ClusterStatusResult{
			Name:    statusConfig.Name,
			Error:   err.Error(),
			Success: false,
		}
	}
	return ClusterStatusResult{
		Name:             statusConfig.Name,
		CrcStatus:        res.CrcStatus,
		OpenshiftStatus:  res.OpenshiftStatus,
		OpenshiftVersion: res.OpenshiftVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		Success:          true,
	}
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

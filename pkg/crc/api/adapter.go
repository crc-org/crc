package api

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine"
)

type AdaptedClient interface {
	Delete() Result
	GetConsoleURL() ConsoleResult
	Start(startConfig machine.StartConfig) StartResult
	Status() ClusterStatusResult
	Stop() Result
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

func (a *Adapter) Delete() Result {
	err := a.Underlying.Delete(constants.DefaultName)
	if err != nil {
		return Result{
			Name:    constants.DefaultName,
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    constants.DefaultName,
		Success: true,
	}
}

func (a *Adapter) GetConsoleURL() ConsoleResult {
	res, err := a.Underlying.GetConsoleURL(constants.DefaultName)
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

func (a *Adapter) Status() ClusterStatusResult {
	res, err := a.Underlying.Status(constants.DefaultName)
	if err != nil {
		return ClusterStatusResult{
			Name:    constants.DefaultName,
			Error:   err.Error(),
			Success: false,
		}
	}
	return ClusterStatusResult{
		Name:             constants.DefaultName,
		CrcStatus:        res.CrcStatus,
		OpenshiftStatus:  res.OpenshiftStatus,
		OpenshiftVersion: res.OpenshiftVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		Success:          true,
	}
}

func (a *Adapter) Stop() Result {
	_, err := a.Underlying.Stop(constants.DefaultName)
	if err != nil {
		return Result{
			Name:    constants.DefaultName,
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    constants.DefaultName,
		Success: true,
	}
}

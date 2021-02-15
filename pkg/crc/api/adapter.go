package api

import (
	"context"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
)

type AdaptedClient interface {
	GetName() string

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

func (a *Adapter) GetName() string {
	return a.Underlying.GetName()
}

func (a *Adapter) Delete() Result {
	err := a.Underlying.Delete()
	if err != nil {
		logging.Error(err)
		return Result{
			Name:    a.Underlying.GetName(),
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    a.Underlying.GetName(),
		Success: true,
	}
}

func (a *Adapter) GetConsoleURL() ConsoleResult {
	res, err := a.Underlying.GetConsoleURL()
	if err != nil {
		logging.Error(err)
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
	res, err := a.Underlying.Start(context.Background(), startConfig)
	if err != nil {
		logging.Error(err)
		return StartResult{
			Name:  a.Underlying.GetName(),
			Error: err.Error(),
		}
	}
	return StartResult{
		Name:           a.Underlying.GetName(),
		Status:         res.Status.String(),
		ClusterConfig:  res.ClusterConfig,
		KubeletStarted: res.KubeletStarted,
	}
}

func (a *Adapter) Status() ClusterStatusResult {
	res, err := a.Underlying.Status()
	if err != nil {
		logging.Error(err)
		return ClusterStatusResult{
			Name:    a.Underlying.GetName(),
			Error:   err.Error(),
			Success: false,
		}
	}
	return ClusterStatusResult{
		Name:             a.Underlying.GetName(),
		CrcStatus:        res.CrcStatus.String(),
		OpenshiftStatus:  res.OpenshiftStatus,
		OpenshiftVersion: res.OpenshiftVersion,
		DiskUse:          res.DiskUse,
		DiskSize:         res.DiskSize,
		Success:          true,
	}
}

func (a *Adapter) Stop() Result {
	_, err := a.Underlying.Stop()
	if err != nil {
		logging.Error(err)
		return Result{
			Name:    a.Underlying.GetName(),
			Success: false,
			Error:   err.Error(),
		}
	}
	return Result{
		Name:    a.Underlying.GetName(),
		Success: true,
	}
}

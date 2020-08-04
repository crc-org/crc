package cmd

import (
	"errors"

	"github.com/code-ready/crc/pkg/crc/machine"
)

type mockClient struct {
}

func (c mockClient) Delete(deleteConfig machine.DeleteConfig) (machine.DeleteResult, error) {
	return machine.DeleteResult{}, errors.New("not implemented")
}

func (c mockClient) GetConsoleURL(consoleConfig machine.ConsoleConfig) (machine.ConsoleResult, error) {
	return machine.ConsoleResult{}, errors.New("not implemented")
}

func (c mockClient) IP(ipConfig machine.IPConfig) (machine.IPResult, error) {
	return machine.IPResult{}, errors.New("not implemented")
}

func (c mockClient) PowerOff(powerOff machine.PowerOffConfig) (machine.PowerOffResult, error) {
	return machine.PowerOffResult{}, errors.New("not implemented")
}

func (c mockClient) Start(startConfig machine.StartConfig) (machine.StartResult, error) {
	return machine.StartResult{}, errors.New("not implemented")
}

func (c mockClient) Stop(stopConfig machine.StopConfig) (machine.StopResult, error) {
	return machine.StopResult{}, errors.New("not implemented")
}

func (mockClient) Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error) {
	return machine.ClusterStatusResult{
		Name:             "crc",
		CrcStatus:        "Running",
		OpenshiftStatus:  "Running",
		OpenshiftVersion: "4.5.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		Success:          true,
	}, nil
}

func (mockClient) Exists(name string) (bool, error) {
	return true, nil
}

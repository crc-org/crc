package cmd

import "github.com/code-ready/crc/pkg/crc/machine"

type client interface {
	Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error)
	Exists(name string) (bool, error)
}

type libmachineClient struct{}

func (*libmachineClient) Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error) {
	return machine.Status(statusConfig)
}

func (*libmachineClient) Exists(name string) (bool, error) {
	return machine.Exists(name)
}

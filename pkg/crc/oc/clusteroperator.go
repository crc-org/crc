package oc

import (
	"encoding/json"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/logging"
	"time"
)

var ignoreClusterOperators = []string{"monitoring", "machine-config", "marketplace"}

type ClusterOperator struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Conditions []struct {
				LastTransitionTime time.Time `json:"lastTransitionTime"`
				Reason             string    `json:"reason"`
				Status             string    `json:"status"`
				Type               string    `json:"type"`
			} `json:"conditions"`
		} `json:"status,omitempty"`
	} `json:"items"`
}

func GetClusterOperatorStatus(oc OcConfig) (bool, error) {
	allAvailable := true
	data, stderr, err := oc.RunOcCommand("get", "co", "-ojson")
	if err != nil {
		return false, fmt.Errorf("%s - %v", stderr, err)
	}

	var co ClusterOperator

	err = json.Unmarshal([]byte(data), &co)
	if err != nil {
		return false, err
	}
	for _, c := range co.Items {
		if contains(c.Metadata.Name, ignoreClusterOperators) {
			continue
		}
		for _, con := range c.Status.Conditions {
			switch con.Type {
			case "Available":
				if con.Status != "True" {
					logging.Debug(c.Metadata.Name, " operator not available, Reason: ", con.Reason)
					allAvailable = false
				}
			case "Degraded":
				if con.Status != "False" {
					logging.Debug(c.Metadata.Name, " operator is degraded, Reason: ", con.Reason)
					allAvailable = false
				}
			case "Progressing":
				if con.Status != "False" {
					logging.Debug(c.Metadata.Name, " operator is still progressing, Reason: ", con.Reason)
					allAvailable = false
				}
			case "Upgradeable":
				continue
			default:
				logging.Debugf("Unexpected operator status for %s: %s", c.Metadata.Name, con.Type)
			}
		}
	}
	return allAvailable, nil
}

func contains(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

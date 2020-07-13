package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

var ignoreClusterOperators = []string{"monitoring", "machine-config", "marketplace", "insights"}

type K8sResource struct {
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

// https://github.com/openshift/cluster-version-operator/blob/master/docs/dev/clusteroperator.md#what-should-an-operator-report-with-clusteroperator-custom-resource
type ClusterStatus struct {
	Available   bool
	Degraded    bool
	Progressing bool
	Disabled    bool
}

func GetClusterOperatorStatus(ocConfig oc.OcConfig, operator string) (*ClusterStatus, error) {
	return getStatus(ocConfig, []string{operator})
}

func GetClusterOperatorsStatus(ocConfig oc.OcConfig) (*ClusterStatus, error) {
	return getStatus(ocConfig, []string{})
}

func getStatus(ocConfig oc.OcConfig, selector []string) (*ClusterStatus, error) {
	cs := &ClusterStatus{
		Available: true,
	}

	data, stderr, err := ocConfig.RunOcCommand("get", "co", "-ojson")
	if err != nil {
		return cs, fmt.Errorf("%s", stderr)
	}

	var co K8sResource
	if err := json.Unmarshal([]byte(data), &co); err != nil {
		return cs, err
	}

	found := false
	for _, c := range co.Items {
		if contains(c.Metadata.Name, ignoreClusterOperators) {
			continue
		}
		if len(selector) > 0 && !contains(c.Metadata.Name, selector) {
			continue
		}
		found = true
		for _, con := range c.Status.Conditions {
			switch con.Type {
			case "Available":
				if con.Status != "True" {
					logging.Debug(c.Metadata.Name, " operator not available, Reason: ", con.Reason)
					cs.Available = false
				}
			case "Degraded":
				if con.Status == "True" {
					logging.Debug(c.Metadata.Name, " operator is degraded, Reason: ", con.Reason)
					cs.Degraded = true
				}
			case "Progressing":
				if con.Status == "True" {
					logging.Debug(c.Metadata.Name, " operator is still progressing, Reason: ", con.Reason)
					cs.Progressing = true
				}
			case "Upgradeable":
				continue
			case "Disabled":
				if con.Status == "True" {
					logging.Debug(c.Metadata.Name, " operator is disabled, Reason: ", con.Reason)
					cs.Disabled = true
				}
			default:
				logging.Debugf("Unexpected operator status for %s: %s", c.Metadata.Name, con.Type)
			}
		}
	}
	if !found {
		return nil, errors.New("no cluster operator found")
	}
	return cs, nil
}

func contains(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

package cluster

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"

	openshiftapi "github.com/openshift/api/config/v1"
)

var ignoreClusterOperators = []string{"monitoring", "machine-config", "marketplace", "insights"}

// https://github.com/openshift/cluster-version-operator/blob/master/docs/dev/clusteroperator.md#what-should-an-operator-report-with-clusteroperator-custom-resource
type Status struct {
	Available   bool
	Degraded    bool
	Progressing bool
	Disabled    bool
}

func GetClusterOperatorStatus(ocConfig oc.Config, operator string) (*Status, error) {
	return getStatus(ocConfig, []string{operator})
}

func GetClusterOperatorsStatus(ocConfig oc.Config) (*Status, error) {
	return getStatus(ocConfig, []string{})
}

func getStatus(ocConfig oc.Config, selector []string) (*Status, error) {
	cs := &Status{
		Available: true,
	}

	data, _, err := ocConfig.RunOcCommandPrivate("get", "co", "-ojson")
	if err != nil {
		return cs, err
	}

	var co openshiftapi.ClusterOperatorList
	if err := json.Unmarshal([]byte(data), &co); err != nil {
		return cs, err
	}

	found := false
	for _, c := range co.Items {
		if contains(c.ObjectMeta.Name, ignoreClusterOperators) {
			continue
		}
		if len(selector) > 0 && !contains(c.ObjectMeta.Name, selector) {
			continue
		}
		found = true
		for _, con := range c.Status.Conditions {
			switch con.Type {
			case "Available":
				if con.Status != "True" {
					logging.Debug(c.ObjectMeta.Name, " operator not available, Reason: ", con.Reason)
					cs.Available = false
				}
			case "Degraded":
				if con.Status == "True" {
					logging.Debug(c.ObjectMeta.Name, " operator is degraded, Reason: ", con.Reason)
					cs.Degraded = true
				}
			case "Progressing":
				if con.Status == "True" {
					logging.Debug(c.ObjectMeta.Name, " operator is still progressing, Reason: ", con.Reason)
					cs.Progressing = true
				}
			case "Upgradeable":
				continue
			case "Disabled":
				if con.Status == "True" {
					logging.Debug(c.ObjectMeta.Name, " operator is disabled, Reason: ", con.Reason)
					cs.Disabled = true
				}
			default:
				logging.Debugf("Unexpected operator status for %s: %s", c.ObjectMeta.Name, con.Type)
			}
		}
	}
	if !found {
		return cs, errors.New("no cluster operator found")
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

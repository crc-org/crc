package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/oc"
	v1 "github.com/openshift/api/config/v1"
	log "github.com/sirupsen/logrus"
)

func StartMonitoring(ocConfig oc.Config) error {
	data, _, err := ocConfig.RunOcCommand("get", "clusterversion/version", "-o", "json")
	if err != nil {
		return err
	}

	var cv v1.ClusterVersion
	if err := json.Unmarshal([]byte(data), &cv); err != nil {
		return err
	}

	pos := -1
	for i, override := range cv.Spec.Overrides {
		if override.Name == "cluster-monitoring-operator" {
			pos = i
			break
		}
	}
	if pos == -1 {
		log.Debug("monitoring operator not found in cluster version overrides")
		return nil
	}

	_, _, err = ocConfig.RunOcCommand("patch", "clusterversion/version",
		"--type", "json",
		"--patch", fmt.Sprintf(`'[{"op":"remove", "path":"/spec/overrides/%d"}]'`, pos))
	return err
}

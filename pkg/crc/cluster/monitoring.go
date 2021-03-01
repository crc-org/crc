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

	indexForClusterMonitoringDeploymentKind := getIndexInOverridesForObjectName(cv, "cluster-monitoring-operator")
	indexForClusterMonitoringCVOKind := getIndexInOverridesForObjectName(cv, "monitoring")

	if indexForClusterMonitoringDeploymentKind != -1 && indexForClusterMonitoringCVOKind != -1 {
		_, _, err = ocConfig.RunOcCommand("patch", "clusterversion/version",
			"--type", "json",
			"--patch", fmt.Sprintf(`'[{"op":"remove", "path":"/spec/overrides/%d"},{"op":"remove", "path":"/spec/overrides/%d"}]'`,
				indexForClusterMonitoringDeploymentKind, indexForClusterMonitoringCVOKind-1))
	}
	return err
}

func getIndexInOverridesForObjectName(cv v1.ClusterVersion, objectName string) int {
	pos := -1
	for i, override := range cv.Spec.Overrides {
		if override.Name == objectName {
			pos = i
			break
		}
	}
	if pos == -1 {
		log.Debugf("%s not found in cluster version overrides", objectName)
	}
	return pos
}

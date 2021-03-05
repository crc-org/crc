package cluster

import (
	"encoding/json"
	"strings"

	"github.com/code-ready/crc/pkg/os"
)

type crictlContainersList struct {
	Containers []crictlContainer `json:"containers"`
}

type crictlContainer struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	State  string `json:"state"`
	Labels struct {
		Namespace string `json:"io.kubernetes.pod.namespace"`
	} `json:"labels"`
}

func VerifyOperatorContainersAreRunning(ssh os.CommandRunner, status *Status) error {
	if !status.IsReady() {
		return nil
	}

	operatorContainers := map[string]string{
		"authentication":               "authentication-operator",
		"dns":                          "dns-operator",
		"etcd":                         "etcd-operator",
		"network":                      "network-operator",
		"openshift-apiserver":          "openshift-apiserver-operator",
		"service-ca":                   "service-ca-operator",
		"config-operator":              "openshift-config-operator",
		"console":                      "console-operator",
		"image-registry":               "cluster-image-registry-operator",
		"ingress":                      "ingress-operator",
		"kube-apiserver":               "kube-apiserver-operator",
		"kube-controller-manager":      "kube-controller-manager-operator",
		"kube-scheduler":               "kube-scheduler-operator-container",
		"marketplace":                  "marketplace-operator",
		"node-tuning":                  "cluster-node-tuning-operator",
		"openshift-controller-manager": "openshift-controller-manager-operator",
		"openshift-samples":            "cluster-samples-operator",
		"operator-lifecycle-manager":   "olm-operator",
	}

	ctrs, err := runningOpenShiftContainers(ssh)
	if err != nil {
		return err
	}

	for operator, container := range operatorContainers {
		if !contains(operator, status.unavailable) && !contains(container, ctrs) {
			status.Available = false
			status.unavailable = append(status.unavailable, operator)
		}
	}

	return nil
}

func runningOpenShiftContainers(runner os.CommandRunner) ([]string, error) {
	output, _, err := runner.RunPrivileged("crictl ps", "timeout", "5s", "crictl", "ps", "-o", "json")
	if err != nil {
		return nil, err
	}
	var list crictlContainersList
	if err := json.Unmarshal([]byte(output), &list); err != nil {
		return nil, err
	}
	var ret []string
	for _, container := range list.Containers {
		if container.State == "CONTAINER_RUNNING" && strings.HasPrefix(container.Labels.Namespace, "openshift-") {
			ret = append(ret, container.Metadata.Name)
		}
	}
	return ret, nil
}

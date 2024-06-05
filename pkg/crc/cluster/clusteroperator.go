package cluster

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	clientset "github.com/openshift/client-go/config/clientset/versioned"
	k8sclient "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcstrings "github.com/crc-org/crc/v2/pkg/strings"
	openshiftapi "github.com/openshift/api/config/v1"
	k8sapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// https://github.com/openshift/cluster-version-operator/blob/master/docs/dev/clusteroperator.md#what-should-an-operator-report-with-clusteroperator-custom-resource
type Status struct {
	Available   bool
	Degraded    bool
	Progressing bool
	Disabled    bool

	progressing []string
	degraded    []string
	unavailable []string
}

const maxNames = 5

func (status *Status) String() string {
	if len(status.progressing) == 1 {
		return fmt.Sprintf("Operator %s is progressing", status.progressing[0])
	} else if len(status.progressing) > 1 {
		return fmt.Sprintf("%d operators are progressing: %s", len(status.progressing), joinWithLimit(status.progressing, maxNames))
	}

	if len(status.degraded) == 1 {
		return fmt.Sprintf("Operator %s is degraded", status.degraded[0])
	} else if len(status.degraded) > 0 {
		return fmt.Sprintf("%d operators are degraded: %s", len(status.degraded), joinWithLimit(status.degraded, maxNames))
	}

	if len(status.unavailable) == 1 {
		return fmt.Sprintf("Operator %s is not yet available", status.unavailable[0])
	} else if len(status.unavailable) > 0 {
		return fmt.Sprintf("%d operators are not available: %s", len(status.unavailable), joinWithLimit(status.unavailable, maxNames))
	}

	if status.IsReady() {
		return "All operators are ready"
	}
	return "Operators are not ready yet"
}

func joinWithLimit(names []string, maxNames int) string {
	sort.Strings(names)
	ret := strings.Join(names[0:min(len(names), maxNames)], ", ")
	if len(names) > maxNames {
		ret += ", ..."
	}
	return ret
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (status *Status) IsReady() bool {
	return status.Available && !status.Progressing && !status.Degraded && !status.Disabled
}

func GetClusterOperatorsStatus(ctx context.Context, ip string, kubeconfigFilePath string) (*Status, error) {
	lister, err := openshiftClient(ip, kubeconfigFilePath)
	if err != nil {
		return nil, err
	}
	return getStatus(ctx, lister.ConfigV1().ClusterOperators(), []string{})
}

func getStatus(ctx context.Context, lister operatorLister, selector []string) (*Status, error) {
	cs := &Status{
		Available: true,
	}

	co, err := lister.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	found := false
	for _, c := range co.Items {
		if len(selector) > 0 && !crcstrings.Contains(selector, c.ObjectMeta.Name) {
			continue
		}
		found = true
		for _, con := range c.Status.Conditions {
			switch con.Type {
			case openshiftapi.OperatorAvailable:
				if con.Status != openshiftapi.ConditionTrue {
					logging.Debug(c.ObjectMeta.Name, " operator not available, Reason: ", con.Reason)
					cs.unavailable = append(cs.unavailable, c.ObjectMeta.Name)
					cs.Available = false
				}
			case openshiftapi.OperatorDegraded:
				if con.Status == openshiftapi.ConditionTrue {
					logging.Debug(c.ObjectMeta.Name, " operator is degraded, Reason: ", con.Reason)
					cs.degraded = append(cs.degraded, c.ObjectMeta.Name)
					cs.Degraded = true
				}
			case openshiftapi.OperatorProgressing:
				if con.Status == openshiftapi.ConditionTrue {
					logging.Debug(c.ObjectMeta.Name, " operator is still progressing, Reason: ", con.Reason)
					cs.progressing = append(cs.progressing, c.ObjectMeta.Name)
					cs.Progressing = true
				}
			case openshiftapi.OperatorUpgradeable:
				continue
			case "Disabled": // non official status, used by insights and cluster baremetal operators
				if con.Status == openshiftapi.ConditionTrue {
					logging.Debug(c.ObjectMeta.Name, " operator is disabled, Reason: ", con.Reason)
					cs.Disabled = true
				}
			case "ManagementStateDegraded": // only for the network operator
				continue
			case "RecentBackup": // only for etcd operator
				continue
			default:
				logging.Debugf("Unexpected operator status for %s: %s", c.ObjectMeta.Name, con.Type)
			}
		}
	}
	if !found {
		return nil, errors.New("no cluster operator found")
	}
	return cs, nil
}

func GetClusterNodeStatus(ctx context.Context, ip string, kubeconfigFilePath string) (*Status, error) {
	status := &Status{
		Available: true,
	}
	clientSet, err := kubernetesClient(ip, kubeconfigFilePath)
	if err != nil {
		return nil, err
	}
	nodes, err := clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ns k8sapi.ConditionStatus
	for _, c := range nodes.Items[0].Status.Conditions {
		if c.Type == k8sapi.NodeReady {
			ns = c.Status
		}
	}
	switch ns {
	case k8sapi.ConditionTrue:
		status.Available = true
	case k8sapi.ConditionFalse:
		status.Progressing = true
	case k8sapi.ConditionUnknown:
		status.Degraded = true
	default:
		logging.Debugf("Unexpected node status for %s", ns)
	}
	return status, nil
}

type operatorLister interface {
	List(ctx context.Context, opts metav1.ListOptions) (*openshiftapi.ClusterOperatorList, error)
}

func openshiftClient(ip string, kubeconfigFilePath string) (*clientset.Clientset, error) {
	config, err := kubernetesClientConfiguration(ip, kubeconfigFilePath)
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

func kubernetesClient(ip string, kubeconfigFilePath string) (*k8sclient.Clientset, error) {
	config, err := kubernetesClientConfiguration(ip, kubeconfigFilePath)
	if err != nil {
		return nil, err
	}
	return k8sclient.NewForConfig(config)
}

func kubernetesClientConfiguration(ip string, kubeconfigFilePath string) (*restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		return nil, err
	}
	// override dial to directly use the IP of the VM
	config.Dial = func(ctx context.Context, _, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "tcp", fmt.Sprintf("%s:6443", ip))
	}
	// discard any proxy configuration of the host
	config.Proxy = func(_ *http.Request) (*url.URL, error) {
		return nil, nil
	}
	config.Timeout = 5 * time.Second
	return config, nil
}

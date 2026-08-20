package cluster

import (
	"context"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/hostsapi"
	k8sapi "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureHostsAPITokenSecret creates or updates the Secret that routes-controller
// mounts as CRC_HOSTS_API_TOKEN.
func EnsureHostsAPITokenSecret(ctx context.Context, ip, kubeconfigFilePath, token string) error {
	if token == "" {
		return fmt.Errorf("hosts API token must not be empty")
	}

	client, err := kubernetesClient(ip, kubeconfigFilePath)
	if err != nil {
		return err
	}

	secrets := client.CoreV1().Secrets(hostsapi.SecretNamespace)
	desired := &k8sapi.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostsapi.SecretName,
			Namespace: hostsapi.SecretNamespace,
		},
		Type: k8sapi.SecretTypeOpaque,
		Data: map[string][]byte{
			hostsapi.SecretKey: []byte(token),
		},
	}

	_, err = secrets.Create(ctx, desired, metav1.CreateOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating hosts API token secret: %w", err)
	}

	existing, err := secrets.Get(ctx, hostsapi.SecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting hosts API token secret: %w", err)
	}
	existing.Data = desired.Data
	existing.Type = desired.Type
	_, err = secrets.Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating hosts API token secret: %w", err)
	}
	return nil
}

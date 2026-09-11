package cluster

import (
	"context"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/restapi"
	k8sapi "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureRestAPITokenSecret creates or updates the Secret that routes-controller
// mounts as CRC_REST_API_TOKEN.
func EnsureRestAPITokenSecret(ctx context.Context, ip, kubeconfigFilePath, token string) error {
	if token == "" {
		return fmt.Errorf("REST API token must not be empty")
	}

	client, err := kubernetesClient(ip, kubeconfigFilePath)
	if err != nil {
		return err
	}

	secrets := client.CoreV1().Secrets(restapi.SecretNamespace)
	desired := &k8sapi.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      restapi.SecretName,
			Namespace: restapi.SecretNamespace,
		},
		Type: k8sapi.SecretTypeOpaque,
		Data: map[string][]byte{
			restapi.SecretKey: []byte(token),
		},
	}

	_, err = secrets.Create(ctx, desired, metav1.CreateOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating REST API token secret: %w", err)
	}

	existing, err := secrets.Get(ctx, restapi.SecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting REST API token secret: %w", err)
	}
	if existing.Data == nil {
		existing.Data = make(map[string][]byte)
	}
	existing.Data[restapi.SecretKey] = []byte(token)
	existing.Type = desired.Type
	_, err = secrets.Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating REST API token secret: %w", err)
	}
	return nil
}

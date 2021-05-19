package cluster

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	v1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	available = &Status{
		Available: true,
	}
	progressing = &Status{
		Available:   true,
		Progressing: true,
		progressing: []string{"authentication"},
	}
)

func TestGetClusterOperatorsStatus(t *testing.T) {
	status, err := getStatus(context.Background(), lister("co.json"), []string{})
	assert.NoError(t, err)
	assert.Equal(t, available, status)
}

func TestGetClusterOperatorsStatusProgressing(t *testing.T) {
	status, err := getStatus(context.Background(), lister("co-progressing.json"), []string{})
	assert.NoError(t, err)
	assert.Equal(t, progressing, status)
}

type mockLister struct {
	file string
}

func (r *mockLister) List(ctx context.Context, opts metav1.ListOptions) (*v1.ClusterOperatorList, error) {
	bin, err := ioutil.ReadFile(r.file)
	if err != nil {
		return nil, err
	}
	var list v1.ClusterOperatorList
	return &list, json.Unmarshal(bin, &list)
}

func lister(s string) *mockLister {
	return &mockLister{
		file: filepath.Join("testdata", s),
	}
}

package validation

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegistriesFromPullSecret(t *testing.T) {
	pullSecret := `{"auths": {"cloud.openshift.com": {"auth": "abcd", "email": "xyz@mail.com"}, "quay.io": {"auth": "sdfghrrret", "email": "abc@mail.net"}}}` //nolint:gosec

	registries, err := getRegistriesFromPullSecret(pullSecret)
	sort.Strings(registries)
	assert.NoError(t, err)
	assert.Equal(t,
		[]string{
			"cloud.openshift.com",
			"quay.io",
		},
		registries,
	)
}

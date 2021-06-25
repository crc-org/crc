package machine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dummyKubeconfigFileContent = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURPRENDQWlDZ0F3SUJBZ0lJUlZmQ0tOVWExd0l3RFFZSktvWklodmNOQVFFTEJRQXdKakVrTUNJR0ExVUUKQXd3YmFXNW5jbVZ6Y3kxdmNHVnlZWFJ2Y2tBeE5UazVPVGt4TURjNE1CNFhEVEl3TURreE16QTVOVGd3T0ZvWApEVEl5TURreE16QTVOVGd3T1Zvd0hURWJNQmtHQTFVRUF3d1NLaTVoY0hCekxXTnlZeTUwWlhOMGFXNW5NSUlCCklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF0Q2xZbk0xRllpSCt3Y1lpR0lVbkx2TmkKekxXV2pzcUpyUS80U25udTYxSERhVTV3M0FuOXlEWm9qeWFKZENtV1VnNkNLWERpQ29KQitseE1GZHlhb2xYVQpkb2hKOXZyMnd0Nml1TmZzaG10eFV3aUJoSTlaQnNWaHp0V2R1M2NnblVjWVc4S015VW1hamlFeVhEOE5wdmJhClo0aWZVc2pBWUUxTEJ5WnpQSWttQm1LUG52MGZGWnUxZWpnZzdIdVFsWW1mTi9wSkh1ZFdNdndaSjhiM1pocW4Kc0EwY0RzYUNEVTFTSms2Y1F2TkpGZ0YzOCtJV05oSnBpUWxHZkZ3S2thbDY0L1JZY1I5eXY4L3BNVis3amoxegpCUEk5eXhiRmdMdXhMeGx5UzhLeG56OWxuLzRWMFpCMTJxVTNjNmFNNEV3M3VoWXY4WkRTT25wVFVTanJSd0lECkFRQUJvM013Y1RBT0JnTlZIUThCQWY4RUJBTUNCYUF3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdFd0RBWUQKVlIwVEFRSC9CQUl3QURBZEJnTlZIUTRFRmdRVVR0TkNCVXl1UHBmNE00WWhsWUJxSmhaR1FZWXdIUVlEVlIwUgpCQll3RklJU0tpNWhjSEJ6TFdOeVl5NTBaWE4wYVc1bk1BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRQmtHWml2CkJxcUVES1VUaWlmRmZUeFFYSnpPNU9CQlRVRHFTbnRya25BdzBzaWRQZ245QThhMmdHZENyN21LRUgxNkZRN04KMVNwa2pYV1pnaEUvMXBaVGFTN0pWY1Q2K1l1cGxmeTVSZW9pbS9ObHU4dWxERGZCU1NwOVM5Qk8rUS9sRkpkTQpvbXd5SXUwc1NYSTc5Q29saExpckRpbkVteVFNREhkbVRKK2tKZmN0cFNIMjdMNGo1b3NKeHROU0FFQUJtVWdYCjFIUTdGWXpyTm15czVPY0s5U2xEbjB2Q3lPVEh5cUJieFJuRjZMOUVjMmJVNktEMER2RUJ1ckJwQVlHTWM5c3MKQkQwRlhkSEVQZzdIUDBOZXA3OWpoZTEwSVhHcmdodGVwM0Q1alVqYmoxSTRER1NzUTh5NGRmMnZvNklGZDk2QQpmMHVVUy83M3NuY042NGdsCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://api.crc.testing:6443
  name: crc
contexts:
- context:
    cluster: crc
    user: test
  name: test
current-context: test
kind: Config
preferences: {}
users:
- name: test
  user:
    token: 85YUSkhhtVmXvB1ltw-Gx92sWhzjMm_5h6Na4A1IZK8
`

func TestCertificateAuthority(t *testing.T) {
	f, err := ioutil.TempFile("", "kubeconfig")
	assert.NoError(t, err, "")
	_, err = f.WriteString(dummyKubeconfigFileContent)
	assert.NoError(t, err, "")
	st, err := certificateAuthority(f.Name())
	assert.NoError(t, err, "")
	expectedString := "-----BEGIN CERTIFICATE-----\nMIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJKoZIhvcNAQELBQAwJjEkMCIGA1UE\nAwwbaW5ncmVzcy1vcGVyYXRvckAxNTk5OTkxMDc4MB4XDTIwMDkxMzA5NTgwOFoX\nDTIyMDkxMzA5NTgwOVowHTEbMBkGA1UEAwwSKi5hcHBzLWNyYy50ZXN0aW5nMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtClYnM1FYiH+wcYiGIUnLvNi\nzLWWjsqJrQ/4Snnu61HDaU5w3An9yDZojyaJdCmWUg6CKXDiCoJB+lxMFdyaolXU\ndohJ9vr2wt6iuNfshmtxUwiBhI9ZBsVhztWdu3cgnUcYW8KMyUmajiEyXD8Npvba\nZ4ifUsjAYE1LByZzPIkmBmKPnv0fFZu1ejgg7HuQlYmfN/pJHudWMvwZJ8b3Zhqn\nsA0cDsaCDU1SJk6cQvNJFgF38+IWNhJpiQlGfFwKkal64/RYcR9yv8/pMV+7jj1z\nBPI9yxbFgLuxLxlyS8Kxnz9ln/4V0ZB12qU3c6aM4Ew3uhYv8ZDSOnpTUSjrRwID\nAQABo3MwcTAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYD\nVR0TAQH/BAIwADAdBgNVHQ4EFgQUTtNCBUyuPpf4M4YhlYBqJhZGQYYwHQYDVR0R\nBBYwFIISKi5hcHBzLWNyYy50ZXN0aW5nMA0GCSqGSIb3DQEBCwUAA4IBAQBkGZiv\nBqqEDKUTiifFfTxQXJzO5OBBTUDqSntrknAw0sidPgn9A8a2gGdCr7mKEH16FQ7N\n1SpkjXWZghE/1pZTaS7JVcT6+Yuplfy5Reoim/Nlu8ulDDfBSSp9S9BO+Q/lFJdM\nomwyIu0sSXI79ColhLirDinEmyQMDHdmTJ+kJfctpSH27L4j5osJxtNSAEABmUgX\n1HQ7FYzrNmys5OcK9SlDn0vCyOTHyqBbxRnF6L9Ec2bU6KD0DvEBurBpAYGMc9ss\nBD0FXdHEPg7HP0Nep79jhe10IXGrghtep3D5jUjbj1I4DGSsQ8y4df2vo6IFd96A\nf0uUS/73sncN64gl\n-----END CERTIFICATE-----\n"
	assert.Equal(t, expectedString, string(st), "")
}

func TestCleanKubeconfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "clean")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	assert.NoError(t, cleanKubeconfig(filepath.Join("testdata", "kubeconfig.in"), filepath.Join(dir, "kubeconfig")))
	actual, err := ioutil.ReadFile(filepath.Join(dir, "kubeconfig"))
	assert.NoError(t, err)
	expected, err := ioutil.ReadFile(filepath.Join("testdata", "kubeconfig.out"))
	assert.NoError(t, err)
	assert.YAMLEq(t, string(expected), string(actual))
}

func TestUpdateUserCaAndKeyToKubeconfig(t *testing.T) {
	f, err := ioutil.TempFile("", "kubeconfig")
	assert.NoError(t, err, "")
	err = updateClientCrtAndKeyToKubeconfig([]byte("dummykey"), []byte("dummycert"), filepath.Join("testdata", "kubeconfig.in"), f.Name())
	assert.NoError(t, err)
	userClientCA, err := adminClientCertificate(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "dummycert", userClientCA)
}

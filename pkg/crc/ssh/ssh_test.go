package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/code-ready/machine/drivers/errdriver"
	machinessh "github.com/code-ready/machine/libmachine/ssh"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestRunner(t *testing.T) {
	dir, err := ioutil.TempDir("", "ssh")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	clientKeyFile := filepath.Join(dir, "private.key")
	writePrivateKey(t, clientKeyFile, clientKey)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	createSSHServer(t, listener, clientKey, func(input string) (byte, string) {
		escaped := fmt.Sprintf("%q", input)
		if escaped == `"echo hello"` {
			return 0, "hello"
		}
		if escaped == `"sudo install -m 0644 /dev/null /hello && cat <<EOF | base64 --decode | sudo tee /hello\naGVsbG8gd29ybGQ=\nEOF"` {
			return 0, ""
		}
		return 1, fmt.Sprintf("unexpected command: %q", input)
	})

	for _, clientType := range []machinessh.ClientType{machinessh.External, machinessh.Native} {
		machinessh.SetDefaultClient(clientType)
		runner := CreateRunnerWithPrivateKey(&mockDriver{
			addr: listener.Addr().String(),
		}, clientKeyFile)

		bin, err := runner.Run("echo hello")
		assert.NoError(t, err)
		assert.Equal(t, "hello", bin)
		assert.NoError(t, runner.CopyData([]byte(`hello world`), "/hello", 0644))

		cmdRunner := NewRemoteCommandRunner(runner)
		stdout, stderr, err := cmdRunner.Run("echo", "hello")
		assert.NoError(t, err)
		assert.Equal(t, "hello", stdout)
		assert.Empty(t, stderr)
	}
}

func createSSHServer(t *testing.T, listener net.Listener, clientKey *rsa.PrivateKey, fun func(string) (byte, string)) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			pub, err := ssh.NewPublicKey(&clientKey.PublicKey)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(pubKey.Marshal(), pub.Marshal()) && c.User() == "core" {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := ssh.NewSignerFromKey(serverKey)
	require.NoError(t, err)
	config.AddHostKey(signer)

	go func() {
		for {
			nConn, err := listener.Accept()
			if err != nil {
				logrus.Debugf("cannot accept connection: %v", err)
				return
			}

			conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
			require.NoError(t, err)
			defer conn.Close()

			logrus.Debugf("logged in with key %s\n", conn.Permissions.Extensions["pubkey-fp"])

			go ssh.DiscardRequests(reqs)

			for newChannel := range chans {
				if newChannel.ChannelType() != "session" {
					_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}

				channel, requests, err := newChannel.Accept()
				require.NoError(t, err)

				go func(in <-chan *ssh.Request) {
					for req := range in {
						command := string(req.Payload[4 : req.Payload[3]+4])
						logrus.Debugf("received command: %s", command)
						_ = req.Reply(req.Type == "exec", nil)

						ret, out := fun(command)
						_, _ = channel.Write([]byte(out))
						_, _ = channel.SendRequest("exit-status", false, []byte{0, 0, 0, ret})
						_ = channel.Close()
					}
				}(requests)
			}
		}
	}()
}

func writePrivateKey(t *testing.T, clientKeyFile string, clientKey *rsa.PrivateKey) {
	privateKeyFile, err := os.OpenFile(clientKeyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer privateKeyFile.Close()
	require.NoError(t, pem.Encode(privateKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	}))
}

type mockDriver struct {
	addr string

	errdriver.Driver
}

func (d *mockDriver) GetSSHHostname() (string, error) {
	return strings.Split(d.addr, ":")[0], nil
}

func (mockDriver) GetSSHKeyPath() string {
	return ""
}

func (d *mockDriver) GetSSHPort() (int, error) {
	return strconv.Atoi(strings.Split(d.addr, ":")[1])
}

func (mockDriver) GetSSHUsername() string {
	return "core"
}

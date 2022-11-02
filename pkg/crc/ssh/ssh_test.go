package ssh

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestRunner(t *testing.T) {
	dir := t.TempDir()

	clientKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err)

	clientKeyFile := filepath.Join(dir, "private.key")
	writePrivateKey(t, clientKeyFile, clientKey)

	cancel, runner, _ := createListenerAndSSHServer(t, clientKey, clientKeyFile)

	assert.NoError(t, err)
	defer runner.Close()

	bin, _, err := runner.Run("echo hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", bin)
	cancel()
	// Expect error when sending data over close ssh server channel
	assert.Error(t, runner.CopyDataPrivileged([]byte(`hello world`), "/hello", 0644))

	_, runner, totalConn := createListenerAndSSHServer(t, clientKey, clientKeyFile)
	assert.NoError(t, runner.CopyDataPrivileged([]byte(`hello world`), "/hello", 0644))
	assert.NoError(t, runner.CopyDataPrivileged([]byte(`hello world`), "/hello", 0644))
	assert.NoError(t, runner.CopyDataPrivileged([]byte(`hello world`), "/hello", 0644))
	assert.NoError(t, runner.CopyData([]byte(`hello world`), "/home/core/hello", 0644))
	assert.Equal(t, 1, *totalConn)
}

func createListenerAndSSHServer(t *testing.T, clientKey *ecdsa.PrivateKey, clientKeyFile string) (context.CancelFunc, *Runner, *int) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	addr := listener.Addr().String()
	runner, err := CreateRunner(ipFor(addr), portFor(addr), clientKeyFile)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	totalConn := createSSHServer(ctx, t, listener, clientKey, func(input string) (byte, string) {
		escaped := fmt.Sprintf("%q", input)
		if escaped == `"echo hello"` {
			return 0, "hello"
		}
		if escaped == `"sudo install -m 0644 /dev/null /hello && cat <<EOF | base64 --decode | sudo tee /hello\naGVsbG8gd29ybGQ=\nEOF"` {
			return 0, ""
		}
		if escaped == `"install -m 0644 /dev/null /home/core/hello && cat <<EOF | base64 --decode | tee /home/core/hello\naGVsbG8gd29ybGQ=\nEOF"` {
			return 0, ""
		}
		return 1, fmt.Sprintf("unexpected command: %q", input)
	})
	return cancel, runner, totalConn
}

func createSSHServer(ctx context.Context, t *testing.T, listener net.Listener, clientKey *ecdsa.PrivateKey, fun func(string) (byte, string)) *int {
	totalConn := 0
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

	serverKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
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
			totalConn++

			conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
			require.NoError(t, err)
			defer conn.Close()

			logrus.Debugf("logged in with key %s\n", conn.Permissions.Extensions["pubkey-fp"])

			go ssh.DiscardRequests(reqs)

			for newChannel := range chans {
				select {
				case <-ctx.Done():
					return
				default:
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
		}
	}()
	return &totalConn
}

func writePrivateKey(t *testing.T, clientKeyFile string, clientKey *ecdsa.PrivateKey) {
	privateKeyFile, err := os.OpenFile(clientKeyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer privateKeyFile.Close()
	bytes, _ := x509.MarshalPKCS8PrivateKey(clientKey)
	require.NoError(t, pem.Encode(privateKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytes,
	}))
}

func ipFor(addr string) string {
	return strings.Split(addr, ":")[0]
}

func portFor(addr string) int {
	port, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	return port
}

func TestGenerateSSHKey(t *testing.T) {
	tmpDir := t.TempDir()

	filename := filepath.Join(tmpDir, "sshkey")

	if err := GenerateSSHKey(filename); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("expected ssh key at %s", filename)
	}
}

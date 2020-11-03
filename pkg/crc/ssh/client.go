package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/code-ready/machine/libmachine/log"
	"golang.org/x/crypto/ssh"
)

type Client interface {
	Output(command string) (string, error)
}

type ExternalClient struct {
	BaseArgs   []string
	BinaryPath string
}

type NativeClient struct {
	Config   ssh.ClientConfig
	Hostname string
	Port     int
}

type Auth struct {
	Keys []string
}

type ClientType string

const (
	External ClientType = "external"
	Native   ClientType = "native"
)

var (
	baseSSHArgs = []string{
		"-F", "/dev/null",
		"-o", "ConnectionAttempts=3", // retry 3 times if SSH connection fails
		"-o", "ConnectTimeout=10", // timeout after 10 seconds
		"-o", "ControlMaster=no", // disable ssh multiplexing
		"-o", "ControlPath=none",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
		"-o", "PasswordAuthentication=no",
		"-o", "ServerAliveInterval=60", // prevents connection to be dropped if command takes too long
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}
	defaultClientType = External
)

func SetDefaultClient(clientType ClientType) {
	// Allow over-riding of default client type, so that even if ssh binary
	// is found in PATH we can still use the Go native implementation if
	// desired.
	switch clientType {
	case External:
		defaultClientType = External
	case Native:
		defaultClientType = Native
	}
}

func NewClient(user string, host string, port int, auth *Auth) (Client, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		log.Debug("SSH binary not found, using native Go implementation")
		client, err := NewNativeClient(user, host, port, auth)
		log.Debug(client)
		return client, err
	}

	if defaultClientType == Native {
		log.Debug("Using SSH client type: native")
		client, err := NewNativeClient(user, host, port, auth)
		log.Debug(client)
		return client, err
	}

	log.Debug("Using SSH client type: external")
	client, err := NewExternalClient(sshBinaryPath, user, host, port, auth)
	log.Debug(client)
	return client, err
}

func NewNativeClient(user, host string, port int, auth *Auth) (Client, error) {
	config, err := NewNativeConfig(user, auth)
	if err != nil {
		return nil, fmt.Errorf("Error getting config for native Go SSH: %s", err)
	}

	return &NativeClient{
		Config:   config,
		Hostname: host,
		Port:     port,
	}, nil
}

func NewNativeConfig(user string, auth *Auth) (ssh.ClientConfig, error) {
	var (
		privateKeys []ssh.Signer
		authMethods []ssh.AuthMethod
	)

	for _, k := range auth.Keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		privateKeys = append(privateKeys, privateKey)
	}

	if len(privateKeys) > 0 {
		authMethods = append(authMethods, ssh.PublicKeys(privateKeys...))
	}

	return ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		// #nosec G106
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Minute,
	}, nil
}

func (client *NativeClient) session() (*ssh.Client, *ssh.Session, error) {
	conn, err := ssh.Dial("tcp", net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port)), &client.Config)
	if err != nil {
		return nil, nil, err
	}
	session, err := conn.NewSession()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, session, err
}

func (client *NativeClient) Output(command string) (string, error) {
	conn, session, err := client.session()
	if err != nil {
		return "", err
	}
	defer closeConn(conn)
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func NewExternalClient(sshBinaryPath, user, host string, port int, auth *Auth) (*ExternalClient, error) {
	client := &ExternalClient{
		BinaryPath: sshBinaryPath,
	}

	args := append(baseSSHArgs, fmt.Sprintf("%s@%s", user, host))

	// If no identities are explicitly provided, also look at the identities
	// offered by ssh-agent
	if len(auth.Keys) > 0 {
		args = append(args, "-o", "IdentitiesOnly=yes")
	}

	// Specify which private keys to use to authorize the SSH request.
	for _, privateKeyPath := range auth.Keys {
		if privateKeyPath != "" {
			// Check each private key before use it
			fi, err := os.Stat(privateKeyPath)
			if err != nil {
				// Abort if key not accessible
				return nil, err
			}
			if runtime.GOOS != "windows" {
				mode := fi.Mode()
				log.Debugf("Using SSH private key: %s (%s)", privateKeyPath, mode)
				// Private key file should have strict permissions
				perm := mode.Perm()
				if perm&0400 == 0 {
					return nil, fmt.Errorf("'%s' is not readable", privateKeyPath)
				}
				if perm&0077 != 0 {
					return nil, fmt.Errorf("permissions %#o for '%s' are too open", perm, privateKeyPath)
				}
			}
			args = append(args, "-i", privateKeyPath)
		}
	}

	// Set which port to use for SSH.
	args = append(args, "-p", fmt.Sprintf("%d", port))

	client.BaseArgs = args

	return client, nil
}

func getSSHCmd(binaryPath string, args ...string) *exec.Cmd {
	return exec.Command(binaryPath, args...)
}

func (client *ExternalClient) Output(command string) (string, error) {
	args := append(client.BaseArgs, command)
	cmd := getSSHCmd(client.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func closeConn(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Debugf("Error closing SSH Client: %s", err)
	}
}

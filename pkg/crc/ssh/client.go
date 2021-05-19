package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	log "github.com/code-ready/crc/pkg/crc/logging"
	"golang.org/x/crypto/ssh"
)

const dialTimeout = 10 * time.Second

type Client interface {
	Run(command string) ([]byte, []byte, error)
	RunWithTimeout(command string, timeout time.Duration) ([]byte, []byte, error)
	Close()
}

type NativeClient struct {
	User     string
	Hostname string
	Port     int
	Keys     []string

	sshClient *ssh.Client
	conn      net.Conn
}

func NewClient(user string, host string, port int, keys ...string) (Client, error) {
	return &NativeClient{
		User:     user,
		Hostname: host,
		Port:     port,
		Keys:     keys,
	}, nil
}

func clientConfig(user string, keys []string) (*ssh.ClientConfig, error) {
	var (
		privateKeys []ssh.Signer
		keyPaths    []string
	)

	for _, k := range keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			continue
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		privateKeys = append(privateKeys, privateKey)
		keyPaths = append(keyPaths, k)
	}

	if len(privateKeys) == 0 {
		return nil, errors.New("no ssh private keys available")
	}
	log.Debugf("Using ssh private keys: %v", keyPaths)

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(privateKeys...)},
		// #nosec G106
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         dialTimeout,
	}, nil
}

func (client *NativeClient) session(timeout time.Duration) (*ssh.Session, error) {
	if client.sshClient == nil {
		var err error
		config, err := clientConfig(client.User, client.Keys)
		if err != nil {
			return nil, fmt.Errorf("Error getting config for native Go SSH: %s", err)
		}
		addr := net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port))
		client.conn, err = net.DialTimeout("tcp", addr, config.Timeout)
		if err != nil {
			return nil, err
		}

		if err := client.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return nil, err
		}
		defer func() {
			_ = client.conn.SetDeadline(time.Time{})
		}()

		c, chans, reqs, err := ssh.NewClientConn(client.conn, addr, config)
		if err != nil {
			return nil, err
		}
		client.sshClient = ssh.NewClient(c, chans, reqs)
	}

	if err := client.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	defer func() {
		_ = client.conn.SetDeadline(time.Time{})
	}()

	session, err := client.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	return session, err
}

func (client *NativeClient) RunWithTimeout(command string, timeout time.Duration) ([]byte, []byte, error) {
	session, err := client.session(timeout)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	if err := client.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = client.conn.SetDeadline(time.Time{})
	}()

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (client *NativeClient) Run(command string) ([]byte, []byte, error) {
	return client.RunWithTimeout(command, 2*time.Minute)
}

func (client *NativeClient) Close() {
	if client.sshClient == nil {
		return
	}
	err := client.sshClient.Close()
	if err != nil {
		log.Debugf("Error closing ssh client: %s", err)
	}
}

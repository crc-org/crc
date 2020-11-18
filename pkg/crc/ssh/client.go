package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"github.com/code-ready/machine/libmachine/log"
	"golang.org/x/crypto/ssh"
)

type Client interface {
	Run(command string) ([]byte, []byte, error)
	Close()
}

type NativeClient struct {
	User     string
	Hostname string
	Port     int
	Keys     []string

	conn *ssh.Client
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
		authMethods []ssh.AuthMethod
	)

	for _, k := range keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			log.Debugf("Cannot read private ssh key %s", k)
			continue
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		privateKeys = append(privateKeys, privateKey)
	}

	if len(privateKeys) > 0 {
		authMethods = append(authMethods, ssh.PublicKeys(privateKeys...))
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		// #nosec G106
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Minute,
	}, nil
}

func (client *NativeClient) session() (*ssh.Session, error) {
	if client.conn == nil {
		var err error
		config, err := clientConfig(client.User, client.Keys)
		if err != nil {
			return nil, fmt.Errorf("Error getting config for native Go SSH: %s", err)
		}
		client.conn, err = ssh.Dial("tcp", net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port)), config)
		if err != nil {
			return nil, err
		}
	}
	session, err := client.conn.NewSession()
	if err != nil {
		return nil, err
	}
	return session, err
}

func (client *NativeClient) Run(command string) ([]byte, []byte, error) {
	session, err := client.session()
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (client *NativeClient) Close() {
	if client.conn == nil {
		return
	}
	err := client.conn.Close()
	if err != nil {
		log.Debugf("Error closing SSH Client: %s", err)
	}
}

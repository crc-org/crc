package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"github.com/code-ready/machine/libmachine/log"
	"golang.org/x/crypto/ssh"
)

type Client interface {
	Output(command string) (string, error)
}

type NativeClient struct {
	Config   ssh.ClientConfig
	Hostname string
	Port     int
}

type Auth struct {
	Keys []string
}

func NewClient(user string, host string, port int, auth *Auth) (Client, error) {
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

func closeConn(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Debugf("Error closing SSH Client: %s", err)
	}
}

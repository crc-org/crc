package cluster

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"golang.org/x/crypto/bcrypt"
)

// GenerateKubeAdminUserPassword creates and put updated kubeadmin password to ~/.crc/machine/crc/kubeadmin-password
func GenerateKubeAdminUserPassword() error {
	logging.Infof("Generating new password for the kubeadmin user")
	kubeAdminPasswordFile := constants.GetKubeAdminPasswordPath()
	kubeAdminPassword, err := GenerateRandomPasswordHash(23)
	if err != nil {
		return fmt.Errorf("Cannot generate the kubeadmin user password: %w", err)
	}
	return ioutil.WriteFile(kubeAdminPasswordFile, []byte(kubeAdminPassword), 0600)
}

// UpdateKubeAdminUserPassword updates the htpasswd secret
func UpdateKubeAdminUserPassword(ocConfig oc.Config, newPassword string) error {
	if newPassword != "" {
		logging.Infof("Overriding password for kubeadmin user")
		if err := ioutil.WriteFile(constants.GetKubeAdminPasswordPath(), []byte(strings.TrimSpace(newPassword)), 0600); err != nil {
			return err
		}
	}

	kubeAdminPassword, err := GetKubeadminPassword()
	if err != nil {
		return fmt.Errorf("Cannot generate the kubeadmin user password: %w", err)
	}
	credentials := map[string]string{
		"developer": "developer",
		"kubeadmin": kubeAdminPassword,
	}

	given, _, err := ocConfig.RunOcCommandPrivate("get", "secret", "htpass-secret", "-n", "openshift-config", "-o", `jsonpath="{.data.htpasswd}"`)
	if err != nil {
		return err
	}
	ok, err := compareHtpasswd(given, credentials)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	logging.Infof("Changing the password for the kubeadmin user")
	expected, err := getHtpasswd(credentials)
	if err != nil {
		return err
	}
	cmdArgs := []string{"patch", "secret", "htpass-secret", "-p",
		fmt.Sprintf(`'{"data":{"htpasswd":"%s"}}'`, expected),
		"-n", "openshift-config", "--type", "merge"}
	_, stderr, err := ocConfig.RunOcCommandPrivate(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to update kubeadmin password %v: %s", err, stderr)
	}
	return nil
}

func GetKubeadminPassword() (string, error) {
	kubeAdminPasswordFile := constants.GetKubeAdminPasswordPath()
	rawData, err := ioutil.ReadFile(kubeAdminPasswordFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(rawData)), nil
}

// generateRandomPasswordHash generates a hash of a random ASCII password
// 5char-5char-5char-5char
// Copied from openshift/installer https://github.com/openshift/installer/blob/master/pkg/asset/password/password.go
func GenerateRandomPasswordHash(length int) (string, error) {
	const (
		lowerLetters = "abcdefghijkmnopqrstuvwxyz"
		upperLetters = "ABCDEFGHIJKLMNPQRSTUVWXYZ"
		digits       = "23456789"
		all          = lowerLetters + upperLetters + digits
	)
	var password string
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
		if err != nil {
			return "", err
		}
		newchar := string(all[n.Int64()])
		if password == "" {
			password = newchar
		}
		if i < length-1 {
			n, err = rand.Int(rand.Reader, big.NewInt(int64(len(password)+1)))
			if err != nil {
				return "", err
			}
			j := n.Int64()
			password = password[0:j] + newchar + password[j:]
		}
	}
	pw := []rune(password)
	for _, replace := range []int{5, 11, 17} {
		pw[replace] = '-'
	}

	return string(pw), nil
}

func hashBcrypt(password string) (hash string, err error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	return string(passwordBytes), nil
}

func getHtpasswd(credentials map[string]string) (string, error) {
	var ret []string
	for username, password := range credentials {
		hash, err := hashBcrypt(password)
		if err != nil {
			return "", err
		}
		ret = append(ret, fmt.Sprintf("%s:%s", username, hash))
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(ret, "\n"))), nil
}

// source https://github.com/openshift/oauth-server/blob/04985077512fec241a5170074bf767c23592d7e7/pkg/authenticator/password/htpasswd/htpasswd.go
func compareHtpasswd(given string, credentials map[string]string) (bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(given)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(decoded))

	found := 0
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		username := parts[0]
		password := parts[1]

		if expectedPassword, ok := credentials[username]; ok {
			ok, err := testBCryptPassword(expectedPassword, password)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
			found++
		}
	}
	if found != len(credentials) {
		return false, nil
	}
	return true, nil
}

func testBCryptPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

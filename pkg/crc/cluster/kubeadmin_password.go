package cluster

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	"golang.org/x/crypto/bcrypt"
)

// GenerateUserPassword creates and put updated password to ~/.crc/machine/crc/ directory
func GenerateUserPassword(passwordFile string, user string) error {
	logging.Infof("Generating new password for the %s user", user)
	password, err := GenerateRandomPasswordHash(23)
	if err != nil {
		return fmt.Errorf("cannot generate the %s user password: %w", user, err)
	}
	return os.WriteFile(passwordFile, []byte(password), 0600)
}

// UpdateUserPasswords updates the htpasswd secret
func UpdateUserPasswords(ctx context.Context, ocConfig oc.Config, newKubeAdminPassword string, newDeveloperPassword string) error {
	credentials, err := resolveUserPasswords(newKubeAdminPassword, newDeveloperPassword, constants.GetKubeAdminPasswordPath(), constants.GetDeveloperPasswordPath())
	if err != nil {
		return err
	}

	if err := WaitForOpenshiftResource(ctx, ocConfig, "secret"); err != nil {
		return err
	}

	given, stderr, err := ocConfig.RunOcCommandPrivate("get", "secret", "htpass-secret", "-n", "openshift-config", "-o", `jsonpath="{.data.htpasswd}"`)
	if err != nil {
		return fmt.Errorf("%s:%v", stderr, err)
	}
	ok, externals, err := compareHtpasswd(given, credentials)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	logging.Infof("Changing the password for the users")
	expected, err := getHtpasswd(credentials, externals)
	if err != nil {
		return err
	}
	cmdArgs := []string{"patch", "secret", "htpass-secret", "-p",
		fmt.Sprintf(`'{"data":{"htpasswd":"%s"}}'`, expected),
		"-n", "openshift-config", "--type", "merge"}
	_, stderr, err = ocConfig.RunOcCommandPrivate(cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to update user passwords %v: %s", err, stderr)
	}
	return nil
}

func GetUserPassword(passwordFile string) (string, error) {
	rawData, err := os.ReadFile(passwordFile)
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

func getHtpasswd(credentials map[string]string, externals []string) (string, error) {
	var ret []string
	ret = append(ret, externals...)
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
func compareHtpasswd(given string, credentials map[string]string) (bool, []string, error) {
	decoded, err := base64.StdEncoding.DecodeString(given)
	if err != nil {
		return false, nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(decoded))

	var externals []string
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
			if err != nil || !ok {
				continue
			}
			found++
		} else {
			externals = append(externals, line)
		}
	}
	if found != len(credentials) {
		return false, externals, nil
	}
	return true, externals, nil
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

func resolveUserPasswords(newKubeAdminPassword string, newDeveloperPassword string, kubeAdminPasswordPath string, developerPasswordPath string) (map[string]string, error) {
	if newKubeAdminPassword != "" {
		logging.Infof("Overriding password for kubeadmin user")
		if err := os.WriteFile(kubeAdminPasswordPath, []byte(strings.TrimSpace(newKubeAdminPassword)), 0600); err != nil {
			return nil, err
		}
	}
	if newDeveloperPassword != "" {
		logging.Infof("Overriding password for developer user")
		if err := os.WriteFile(developerPasswordPath, []byte(strings.TrimSpace(newDeveloperPassword)), 0600); err != nil {
			return nil, err
		}
	}

	kubeAdminPassword, err := GetUserPassword(kubeAdminPasswordPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read the kubeadmin user password from file: %w", err)
	}
	developerPassword, err := GetUserPassword(developerPasswordPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read the developer user password from file: %w", err)
	}
	return map[string]string{
		"developer": developerPassword,
		"kubeadmin": kubeAdminPassword,
	}, nil
}

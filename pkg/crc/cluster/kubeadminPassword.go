package cluster

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/oc"
	"golang.org/x/crypto/bcrypt"
)

// UpdateKubeAdminUserPassword does following
// - Create and put updated kubeadmin password to ~/.crc/machine/crc/kubeadmin-password
// - Update the htpasswd secret
func UpdateKubeAdminUserPassword(kubeAdminPassword string, ocConfig oc.Config, bundle *bundle.CrcBundleInfo) error {
	// In case of cluster started from stopped state.
	if kubeAdminPassword == "" {
		return nil
	}
	hashDeveloperPasswd, err := hashBcrypt("developer")
	if err != nil {
		return err
	}

	hashKubeAdminPasswd, err := hashBcrypt(kubeAdminPassword)
	if err != nil {
		return err
	}
	base64Data := getBase64(hashDeveloperPasswd, hashKubeAdminPasswd)

	cmdArgs := []string{"patch", "secret", "htpass-secret", "-p",
		fmt.Sprintf(`'{"data":{"htpasswd":"%s"}}'`, base64Data),
		"-n", "openshift-config", "--type", "merge"}
	_, stderr, err := ocConfig.RunOcCommandPrivate(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to update kubeadmin password %v: %s", err, stderr)
	}
	if err := bundle.UpdateKubeadminPassword(kubeAdminPassword); err != nil {
		return err
	}
	return nil
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

func getBase64(developerUserPassword, adminUserPassword string) string {
	s := fmt.Sprintf("developer:%s\nkubeadmin:%s", developerUserPassword, adminUserPassword)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

package cluster

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "crc"
	keyringUser    = "compressed-pull-secret"
)

type PullSecretLoader interface {
	Value() (string, error)
}

type interactivePullSecretLoader struct {
	nonInteractivePullSecretLoader *nonInteractivePullSecretLoader
}

func NewInteractivePullSecretLoader(config crcConfig.Storage) PullSecretLoader {
	return &PullSecretMemoizer{
		Getter: &interactivePullSecretLoader{
			nonInteractivePullSecretLoader: &nonInteractivePullSecretLoader{
				config: config,
			},
		},
	}
}

func (loader *interactivePullSecretLoader) Value() (string, error) {
	fromNonInteractive, err := loader.nonInteractivePullSecretLoader.Value()
	if err == nil {
		return fromNonInteractive, nil
	}

	pullSecret, err := promptUserForSecret()
	if err != nil {
		return "", err
	}

	if err := StoreInKeyring(pullSecret); err != nil {
		logging.Warnf("Cannot add pull secret to keyring: %v", err)
	}
	return pullSecret, nil
}

type nonInteractivePullSecretLoader struct {
	config crcConfig.Storage
	path   string
}

func NewNonInteractivePullSecretLoader(config crcConfig.Storage, path string) PullSecretLoader {
	return &PullSecretMemoizer{
		Getter: &nonInteractivePullSecretLoader{
			config: config,
			path:   path,
		},
	}
}

func (loader *nonInteractivePullSecretLoader) Value() (string, error) {
	// If crc is built from an OKD bundle, then use the fake pull secret in contants.
	if crcversion.IsOkdBuild() {
		return constants.OkdPullSecret, nil
	}

	if loader.path != "" {
		fromPath, err := loadFile(loader.path)
		if err == nil {
			logging.Debugf("Using secret from path %q", loader.path)
			return fromPath, nil
		}
		logging.Debugf("Cannot load secret from path %q: %v", loader.path, err)
	}
	fromConfig, err := loadFile(loader.config.Get(crcConfig.PullSecretFile).AsString())
	if err == nil {
		logging.Debugf("Using secret from configuration")
		return fromConfig, nil
	}
	logging.Debugf("Cannot load secret from configuration: %v", err)

	fromKeyring, err := loadFromKeyring()
	if err == nil {
		logging.Debugf("Using secret from keyring")
		return fromKeyring, nil
	}
	logging.Debugf("Cannot load secret from keyring: %v", err)

	return "", fmt.Errorf("unable to load pull secret from path %q or from configuration", loader.path)
}

func loadFromKeyring() (string, error) {
	pullsecret, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(pullsecret)
	if err != nil {
		return "", err
	}
	decompressor, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	// #nosec G110
	if _, err := io.Copy(&b, decompressor); err != nil {
		return "", err
	}
	if err := decompressor.Close(); err != nil {
		return "", err
	}
	return b.String(), validation.ImagePullSecret(b.String())
}

func StoreInKeyring(pullSecret string) error {
	var b bytes.Buffer

	if err := validation.ImagePullSecret(pullSecret); err != nil {
		return err
	}

	compressor := gzip.NewWriter(&b)
	if _, err := compressor.Write([]byte(pullSecret)); err != nil {
		return err
	}
	if err := compressor.Close(); err != nil {
		return err
	}
	return keyring.Set(keyringService, keyringUser, base64.StdEncoding.EncodeToString(b.Bytes()))
}

func ForgetPullSecret() error {
	_ = keyring.Delete(keyringService, keyringUser)
	return nil
}

func loadFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	pullsecret := strings.TrimSpace(string(data))
	return pullsecret, validation.ImagePullSecret(pullsecret)
}

const helpMessage = `CodeReady Containers requires a pull secret to download content from Red Hat.
You can copy it from the Pull Secret section of %s.
`

// promptUserForSecret can be used for any kind of secret like image pull
// secret or for password.
func promptUserForSecret() (string, error) {
	if !crcos.RunningInTerminal() {
		return "", errors.New("cannot ask for secret, crc not launched by a terminal")
	}

	fmt.Printf(helpMessage, constants.CrcLandingPageURL)
	var secret string
	prompt := &survey.Password{
		Message: "Please enter the pull secret",
	}
	if err := survey.AskOne(prompt, &secret, survey.WithValidator(func(ans interface{}) error {
		return validation.ImagePullSecret(ans.(string))
	})); err != nil {
		return "", err
	}
	return secret, nil
}

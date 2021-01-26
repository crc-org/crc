package cluster

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	terminal "golang.org/x/term"
	"gopkg.in/AlecAivazis/survey.v1"
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

	return promptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.CrcLandingPageURL))
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
			return fromPath, nil
		}
		logging.Debugf("Cannot load secret from path %q: %v", loader.path, err)
	}
	fromConfig, err := loadFile(loader.config.Get(cmdConfig.PullSecretFile).AsString())
	if err == nil {
		return fromConfig, nil
	}
	logging.Debugf("Cannot load secret from configuration: %v", err)
	return "", fmt.Errorf("unable to load pull secret from path %q or from configuration", loader.path)
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

// promptUserForSecret can be used for any kind of secret like image pull
// secret or for password.
func promptUserForSecret(message string, help string) (string, error) {
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("cannot ask for secret, crc not launched by a terminal")
	}

	var secret string
	prompt := &survey.Password{
		Message: message,
		Help:    help,
	}
	if err := survey.AskOne(prompt, &secret, func(ans interface{}) error {
		return validation.ImagePullSecret(ans.(string))
	}); err != nil {
		return "", err
	}
	return secret, nil
}

package cluster

import (
	"fmt"
	"io/ioutil"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
)

type PullSecretLoader interface {
	Value() (string, error)
}

type interactivePullSecretLoader struct {
	config *crcConfig.Config
}

func NewInteractivePullSecretLoader(config *crcConfig.Config) PullSecretLoader {
	return &PullSecretMemoizer{
		Getter: &interactivePullSecretLoader{
			config: config,
		},
	}
}

func (loader *interactivePullSecretLoader) Value() (string, error) {
	var (
		pullsecret string
		err        error
	)

	// If crc is built from an OKD bundle, then use the fake pull secret in contants.
	if crcversion.IsOkdBuild() {
		pullsecret = constants.OkdPullSecret
		return pullsecret, nil
	}
	// In case user doesn't provide a file in start command or in config then ask for it.
	if loader.config.Get(cmdConfig.PullSecretFile).AsString() == "" {
		pullsecret, err = input.PromptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.CrcLandingPageURL))
		// This is just to provide a new line after user enter the pull secret.
		fmt.Println()
		if err != nil {
			return "", err
		}
	} else {
		// Read the file content
		data, err := ioutil.ReadFile(loader.config.Get(cmdConfig.PullSecretFile).AsString())
		if err != nil {
			return "", err
		}
		pullsecret = string(data)
	}
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", err
	}

	return pullsecret, nil
}

type nonInteractivePullSecretLoader struct {
	path string
}

func NewNonInteractivePullSecretLoader(path string) PullSecretLoader {
	return &PullSecretMemoizer{
		Getter: &nonInteractivePullSecretLoader{
			path: path,
		},
	}
}

func (loader *nonInteractivePullSecretLoader) Value() (string, error) {
	data, err := ioutil.ReadFile(loader.path)
	if err != nil {
		return "", err
	}
	pullsecret := string(data)
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", err
	}
	return pullsecret, nil
}

package gpg

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/crc-org/crc/pkg/crc/constants"
)

func Verify(filePath, signatureFilePath string) error {
	data, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer data.Close()

	signature, err := os.Open(signatureFilePath)
	if err != nil {
		return err
	}
	defer signature.Close()

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(constants.GPGPublicKey))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %s", err)
	}

	if _, err = openpgp.CheckArmoredDetachedSignature(keyring, data, signature, nil); err != nil {
		return fmt.Errorf("failed to check signature: %s", err)
	}
	return nil
}

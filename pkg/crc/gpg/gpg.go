package gpg

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
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

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(constants.CrcOrgPublicKey))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	if _, err = openpgp.CheckArmoredDetachedSignature(keyring, data, signature, nil); err != nil {
		return fmt.Errorf("failed to check signature: %w", err)
	}
	return nil
}

func GetVerifiedClearsignedMsg(pubkey, clearSignedMsg string) (string, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(pubkey))
	if err != nil {
		return "", fmt.Errorf("Unable to read pubkey: %w", err)
	}
	block, rest := clearsign.Decode([]byte(clearSignedMsg))
	if block == nil || len(rest) != 0 {
		return "", fmt.Errorf("Error decoding clear signed message")
	}

	signer, err := block.VerifySignature(keyring, nil)
	if err != nil {
		return "", fmt.Errorf("Invalid signature: %w", err)
	}
	logging.Debugf("Got valid signature from key id: %s", signer.PrimaryKey.KeyIdString())
	// block.VerifySignature checked the signature against block.Bytes, the
	// canonical form of the signed text (dash-escaping removed, trailing
	// whitespace stripped, CRLF line endings). block.Plaintext holds the same
	// text with LF line endings.
	return string(block.Plaintext), nil
}

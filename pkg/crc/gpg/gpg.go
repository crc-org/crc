package gpg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	goOpenpgp "golang.org/x/crypto/openpgp"             //nolint
	goClearsign "golang.org/x/crypto/openpgp/clearsign" //nolint
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
		return fmt.Errorf("failed to parse public key: %s", err)
	}

	if _, err = openpgp.CheckArmoredDetachedSignature(keyring, data, signature, nil); err != nil {
		return fmt.Errorf("failed to check signature: %s", err)
	}
	return nil
}

func GetVerifiedClearsignedMsgV3(pubkey, clearSignedMsg string) (string, error) {
	k, err := goOpenpgp.ReadArmoredKeyRing(bytes.NewBufferString(pubkey))
	if err != nil {
		return "", fmt.Errorf("Unable to read pubkey: %w", err)
	}
	block, rest := goClearsign.Decode([]byte(clearSignedMsg))
	if len(rest) != 0 {
		return "", fmt.Errorf("Error decoding clear signed message")
	}
	sig, err := io.ReadAll(block.ArmoredSignature.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading signature: %w", err)
	}

	// CheckDetachedSignature method expects the clear text msg
	// in the canonical format as defined in the pgp spec which
	// says that each line of the text needs to end with \r\n
	clearTextMsg := make([]byte, len(block.Bytes))
	copy(clearTextMsg, block.Bytes)
	canonicalizedMsgText := canonicalize(trimEachLine(string(clearTextMsg)))

	id, err := goOpenpgp.CheckDetachedSignature(k, bytes.NewBufferString(canonicalizedMsgText), bytes.NewBuffer(sig))
	if err != nil {
		return "", fmt.Errorf("Invalid signature: %w", err)
	}
	logging.Debugf("Got valid signature from key id: %s", id.PrimaryKey.KeyIdString())
	return trimEachLine(string(clearTextMsg)), nil
}

// https://github.com/ProtonMail/gopenpgp/blob/5aebf6a366fd8b81e80c337186fdaa0793597354/internal/common.go#L10-L12
func canonicalize(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\n", "\r\n")
}

// https://github.com/ProtonMail/gopenpgp/blob/5aebf6a366fd8b81e80c337186fdaa0793597354/internal/common.go#L14-L22
func trimEachLine(text string) string {
	lines := strings.Split(text, "\n")

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}

	return strings.Join(lines, "\n")
}

package ssh

import (
	"encoding/pem"
	"os"

	"github.com/hectane/go-acl"
	"golang.org/x/sys/windows"
)

// This is based on https://github.com/hectane/go-acl/blob/ca0b05cb1adbf3d91585d8acab7bc804ee0b8583/chmod.go#L11-L38
// but only sets an ACL for the current user, and not for the group or other users
func winChmodUser(name string, fileMode os.FileMode) error {
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	creatorOwnerSID, err := windows.StringToSid("S-1-3-0")
	if err != nil {
		return err
	}

	mode := uint32(fileMode)
	return acl.Apply(
		name,
		true,
		false,
		acl.GrantSid(((mode&0700)<<23)|((mode&0200)<<9), creatorOwnerSID),
	)
}

// WriteToFile writes keypair to files
func (kp *KeyPair) WriteToFile(privateKeyPath string, publicKeyPath string) error {
	files := []struct {
		File  string
		Type  string
		Value []byte
	}{
		{
			File:  privateKeyPath,
			Value: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Headers: nil, Bytes: kp.PrivateKey}),
		},
		{
			File:  publicKeyPath,
			Value: kp.PublicKey,
		},
	}

	for _, v := range files {
		f, err := os.Create(v.File)
		if err != nil {
			return ErrUnableToWriteFile
		}
		defer f.Close()

		if _, err := f.Write(v.Value); err != nil {
			return ErrUnableToWriteFile
		}

		if err := winChmodUser(v.File, 0600); err != nil {
			return err
		}
	}

	return nil
}

package crcsuite

import (
	"fmt"
	"os"
	"path/filepath"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

// DeleteCRC deletes CRC instance
func DeleteCRC() error {

	command := "crc delete"
	_ = clicumber.ExecuteCommand(command)

	fmt.Printf("Deleted CRC instance (if one existed).\n")
	return nil
}

// RemoveCRCHome removes CRC home folder ~/.crc
func RemoveCRCHome() error {

	keepFile := filepath.Join(CRCHome, ".keep")

	_, err := os.Stat(keepFile)
	if err != nil { // cannot get keepFile's status
		err = os.RemoveAll(CRCHome)

		if err != nil {
			fmt.Printf("Problem deleting CRC home folder %s.\n", CRCHome)
			return err
		}

		fmt.Printf("Deleted CRC home folder %s.\n", CRCHome)
		return nil

	}
	// keepFile exists
	return fmt.Errorf("Folder %s not removed as per request: %s present", CRCHome, keepFile)
}

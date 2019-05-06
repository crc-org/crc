// +build integration

package crcsuite

import (
	"fmt"
	"os"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

// Delete CRC instance
func DeleteCRC() error {

	command := "crc delete"
	_ = clicumber.ExecuteCommand(command)

	fmt.Printf("Deleted CRC instance (if one existed).\n")
	return nil
}

// Remove CRC home folder ~/.crc
func RemoveCRCHome() error {

	err := os.RemoveAll(CRCHome)

	if err != nil {
		fmt.Printf("Problem deleting CRC home folder %s.\n", CRCHome)
		return err
	}

	fmt.Printf("Deleted CRC home folder %s.\n", CRCHome)
	return nil

}

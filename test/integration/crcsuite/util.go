// +build integration

package crcsuite

import (
	"fmt"
	"os"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

// Delete CRC instance
func DeleteCRC() {

	command := "crc delete"
	err := clicumber.ExecuteCommandSucceedsOrFails(command, "succeeds")
	if err != nil {
		fmt.Errorf("Failed to delete CRC.\n")
	} else {
		fmt.Printf("Deleted CRC instance (if one existed).\n")
	}
}

// Delete CRC instance
func ForceStopCRC() {

	command := "crc stop -f"
	err := clicumber.ExecuteCommandSucceedsOrFails(command, "succeeds")
	if err != nil {
		fmt.Errorf("Failed to forcibly stop CRC.\n")
	} else {
		fmt.Printf("Forcibly stopped CRC instance (if one was running).\n")
	}
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

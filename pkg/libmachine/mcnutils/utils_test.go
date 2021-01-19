package mcnutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	testStr := "test-machine"

	srcFile, err := ioutil.TempFile("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	srcFi, err := srcFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	_, _ = srcFile.Write([]byte(testStr))
	srcFile.Close()

	srcFilePath := filepath.Join(os.TempDir(), srcFi.Name())

	destFile, err := ioutil.TempFile("", "machine-copy-test-")
	if err != nil {
		t.Fatal(err)
	}

	destFi, err := destFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	destFile.Close()

	destFilePath := filepath.Join(os.TempDir(), destFi.Name())

	if err := CopyFile(srcFilePath, destFilePath); err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(destFilePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testStr {
		t.Fatalf("expected data \"%s\"; received \"%s\"", testStr, string(data))
	}
}

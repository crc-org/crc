package os

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	testStr := "test-machine"

	srcFile, err := os.CreateTemp(os.TempDir(), "machine-test-")
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

	destFile, err := os.CreateTemp(os.TempDir(), "machine-copy-test-")
	if err != nil {
		t.Fatal(err)
	}
	destFi, err := destFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	destFile.Close()

	tempDirRoot, err := os.OpenRoot(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer tempDirRoot.Close()

	destFilePath := filepath.Join(tempDirRoot.Name(), destFi.Name())

	if err := CopyFile(srcFilePath, destFilePath); err != nil {
		t.Fatal(err)
	}

	data, err := tempDirRoot.ReadFile(destFi.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testStr {
		t.Fatalf("expected data \"%s\"; received \"%s\"", testStr, string(data))
	}
}

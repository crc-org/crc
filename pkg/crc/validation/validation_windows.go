//go:build windows

package validation

import "golang.org/x/sys/windows"

func freeDiskSpace(path string) (uint64, error) {
	var freeBytesAvailable uint64
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailable, nil, nil,
	)
	if err != nil {
		return 0, err
	}
	return freeBytesAvailable, nil
}

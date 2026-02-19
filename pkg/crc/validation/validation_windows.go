//go:build windows

package validation

import "golang.org/x/sys/windows"

func freeDiskSpace(path string) (uint64, error) {
	var freeBytesAvailable uint64
	err := windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(path),
		&freeBytesAvailable, nil, nil,
	)
	if err != nil {
		return 0, err
	}
	return freeBytesAvailable, nil
}

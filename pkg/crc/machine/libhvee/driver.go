package libhvee

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ConvertToUnixPath converts a path like c:\users\crc to /mnt/c/users/crc
func ConvertToUnixPath(path string) string {
	/* podman internally converts windows style paths like C:\Users\crc to
	 * /mnt/c/Users/crc so it expects the shared folder to be mounted under
	 * '/mnt' instead of '/' like in the case of macOS and linux
	 * see: https://github.com/containers/podman/blob/468aa6478c73e4acd8708ce8bb0bb5a056f329c2/pkg/specgen/winpath.go#L24-L59
	 */
	if runtime.GOOS == "windows" {
		path = filepath.ToSlash(path)
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	if len(path) > 1 && path[1] == ':' {
		return "/mnt/" + strings.ToLower(path[0:1]) + path[2:]
	}
	return path
}

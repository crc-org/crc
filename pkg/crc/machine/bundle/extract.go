package bundle

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/extract"
)

func Extract(sourcepath string) (*CrcBundleInfo, error) {
	err := extract.Uncompress(sourcepath, constants.MachineCacheDir)
	if err != nil {
		return nil, err
	}

	return GetCachedBundleInfo(filepath.Base(sourcepath))
}

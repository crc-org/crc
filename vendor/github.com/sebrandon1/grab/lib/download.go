package lib

import (
	"context"
)

type DownloadResponse struct {
	Filename string
	Err      error
}

// DownloadBatch downloads multiple files concurrently to the current directory.
// It is a wrapper around GetBatch, returning a channel of DownloadResponse for CLI and library use.
func DownloadBatch(ctx context.Context, urls []string) (<-chan DownloadResponse, error) {
	ch := make(chan DownloadResponse)
	respch, err := GetBatch(0, ".", urls...)
	if err != nil {
		return nil, err
	}
	go func() {
		for resp := range respch {
			ch <- DownloadResponse{
				Filename: resp.Filename,
				Err:      resp.Err(),
			}
		}
		close(ch)
	}()
	return ch, nil
}

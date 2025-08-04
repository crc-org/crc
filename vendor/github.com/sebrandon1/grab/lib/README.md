# grab/lib

This package provides the core Go library for downloading files from the internet. It is designed to be used both by the grab CLI and by other Go programs via module import.

## Public API

### Types

- **DownloadResponse**
  - `Filename string`: The name of the file downloaded.
  - `Err error`: Any error encountered during download (nil if successful).

### Functions

- **DownloadBatch(ctx context.Context, urls []string) (<-chan DownloadResponse, error)**
  - Downloads multiple files concurrently to the current directory.
  - Returns a channel of `DownloadResponse` for each file.

- **GetBatch(workers int, dst string, urls ...string) (<-chan *Response, error)**
  - Lower-level API for advanced use cases. Downloads files to `dst` with a specified number of workers.
  - Returns a channel of `*Response` for each file.

- **Get(dst, url string) (*Response, error)**
  - Downloads a single file to the specified destination.
  - Returns a `*Response` with details about the download.

- **NewClient() *Client**
  - Returns a new file download client for advanced/custom use.

- **NewRequest(dst, url string) (*Request, error)**
  - Creates a new download request for use with a client.

## Usage Examples

### Download a batch of files (simple)

```go
package main

import (
	"context"
	"log"
	"github.com/sebrandon1/grab/lib"
)

func main() {
	urls := []string{"https://example.com/file1.zip", "https://example.com/file2.zip"}
	ch, err := lib.DownloadBatch(context.Background(), urls)
	if err != nil {
		log.Fatal(err)
	}
	for resp := range ch {
		if resp.Err != nil {
			log.Printf("Failed: %s (%v)", resp.Filename, resp.Err)
		} else {
			log.Printf("Downloaded: %s", resp.Filename)
		}
	}
}
```

### Advanced: Use a custom client and request

```go
package main

import (
	"log"
	"github.com/sebrandon1/grab/lib"
)

func main() {
	client := lib.NewClient()
	urls := []string{"https://example.com/file1.zip", "https://example.com/file2.zip"}
	for _, url := range urls {
		req, err := lib.NewRequest(".", url)
		if err != nil {
			log.Printf("Invalid URL: %s (%v)", url, err)
			continue
		}
		resp := client.Do(req)
		<-resp.Done // wait for download to finish
		if err := resp.Err(); err != nil {
			log.Printf("Failed: %s (%v)", resp.Filename, err)
		} else {
			log.Printf("Downloaded: %s", resp.Filename)
		}
	}
}
```

## See Also

- [Main README](../README.md)

## License

MIT

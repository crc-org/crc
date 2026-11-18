# grab/lib

The core Go library for downloading files from the internet. It is used by the
`grab` CLI and can be imported directly by other Go programs.

```bash
go get github.com/sebrandon1/grab/lib
```

```go
import "github.com/sebrandon1/grab/lib"
```

## Entry Points

| Function | Signature | Notes |
|----------|-----------|-------|
| `Get` | `Get(dst, urlStr string) (*Response, error)` | Simplest call. Blocks until the download completes. `dst` is a file path or an existing directory; `""` or `"."` means the current directory. |
| `GetBatch` | `GetBatch(ctx context.Context, workers int, dst string, urls ...string) (<-chan *Response, error)` | Downloads to an **existing** directory. `workers < 1` means one worker per URL. The channel closes when every download has finished. |
| `DownloadBatch` | `DownloadBatch(ctx context.Context, urls []string) (<-chan DownloadResponse, error)` | Convenience wrapper over `GetBatch` that writes to `.` and yields `{Filename, Err}` pairs. |
| `NewClient` | `NewClient() *Client` | Returns a client with defaults (`UserAgent: "grab"`, an `*http.Client`). |
| `DefaultClient` | `var DefaultClient *Client` | The client used by `Get`, `GetBatch`, and `DownloadBatch`. Mutating it affects those helpers. |
| `NewRequest` | `NewRequest(dst, urlStr string) (*Request, error)` | Builds a `GET` request. `dst` is the destination file path or directory. |

`DownloadResponse` is the small result type returned by `DownloadBatch`:

```go
type DownloadResponse struct {
	Filename string
	Err      error
}
```

## Client

`Client` is safe for concurrent use by multiple goroutines.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `HTTPClient` | `HTTPClient` interface | `*http.Client` | Transport used for every request. Replace it to inject a proxy, custom TLS, or a test double. |
| `UserAgent` | `string` | `"grab"` | Sent on requests that do not already set a `User-Agent` header. |
| `BufferSize` | `int` | 32 KiB (applied when ≤ 0) | Transfer buffer size. Larger buffers mean more throughput and fewer progress updates. Overridden per request by `Request.BufferSize`. |
| `RetryLimit` | `int` | `0` (no retries) | Retry attempts for transient failures: transport errors and status codes `429`, `500`, `502`, `503`, `504`. |
| `RetryDelay` | `time.Duration` | `0` | Minimum wait between attempts. After a `429`, a server's `Retry-After` header wins when it asks for longer, capped at 5 minutes. |

### Client methods

| Method | Description |
|--------|-------------|
| `Do(req *Request) *Response` | Starts a transfer and returns before the body is copied. Policy errors, HTTP errors, and IO errors surface on `Response.Err`. |
| `DoBatch(ctx, workers int, requests ...*Request) <-chan *Response` | Runs requests with a bounded worker pool (`workers < 1` = one per request). The channel closes after all transfers finish. |
| `DoChannel(ctx, reqch <-chan *Request, respch chan<- *Response)` | Runs requests one at a time until `reqch` is closed or `ctx` is done. Keep `respch` buffered so a slow receiver cannot stall a transfer. |
| `SetTimeout(d time.Duration)` | Sets the dial, TLS handshake, and response-header timeouts on the underlying transport. |

`SetTimeout` deliberately does **not** cap total transfer time, so a slow but
healthy large download is never killed. Use a `context` (see
[Request options](#request-options)) for a whole-transfer deadline.

Two constraints on `SetTimeout`:

- It is a **no-op** when `d <= 0`, when `HTTPClient` is not an `*http.Client`,
  or when that client's transport is not an `*http.Transport` (for example, a
  test mock).
- It mutates the shared transport, so call it **before** the first request.

`HTTPClient` is the one-method interface `Do(*http.Request) (*http.Response, error)`;
`*http.Client` satisfies it.

## Request options

Create a request with `NewRequest(dst, urlStr)` and set fields before passing it
to `Client.Do`.

| Field | Type | Description |
|-------|------|-------------|
| `Label` | `string` | Arbitrary human-friendly name; grab never interprets it. |
| `Tag` | `interface{}` | Arbitrary value to correlate the request with your own data. |
| `HTTPRequest` | `*http.Request` | The request sent to the server. Set headers (auth, `Accept`, ...) here. |
| `Filename` | `string` | Destination path. Empty or a directory means the name is resolved from `Content-Disposition` or the URL. |
| `SkipExisting` | `bool` | If the destination exists, fail with `ErrFileExists` without checking completeness. |
| `NoResume` | `bool` | Never append to a partial file; download it again from the start. An already-complete file is still accepted. |
| `NoStore` | `bool` | Do not touch the filesystem. The body is buffered in memory and read with `Response.Open` / `Response.Bytes`. |
| `NoCreateDirectories` | `bool` | Do not create missing parent directories. |
| `IgnoreBadStatusCodes` | `bool` | Accept any status code instead of requiring 2xx. Also disables retry-on-transient-status. |
| `IgnoreRemoteTime` | `bool` | Do not apply the remote `Last-Modified` timestamp to the local file. |
| `Size` | `int64` | Expected total size. A mismatch cancels the transfer with `ErrBadLength`. |
| `BufferSize` | `int` | Per-request transfer buffer. Defaults to `Client.BufferSize`, then 32 KiB. |
| `RateLimiter` | `RateLimiter` | Paces the transfer. See [Rate limiting](#rate-limiting). |
| `BeforeCopy` | `Hook` | Called immediately before the body is copied. Returning an error cancels the transfer. |
| `AfterCopy` | `Hook` | Called after the body is copied and before checksum validation. Returning an error cancels the transfer. |

`Hook` is `func(*Response) error`. Hooks run synchronously, so they must not
call blocking `Response` methods such as `Err`, `Wait`, or `Cancel` — that
deadlocks. Return an error to cancel instead.

### Request methods

| Method | Description |
|--------|-------------|
| `Context() context.Context` | The request's context, defaulting to `context.Background()`. |
| `WithContext(ctx) *Request` | Returns a **shallow copy** with a new context. Reassign the result: `req = req.WithContext(ctx)`. |
| `URL() *url.URL` | The URL being downloaded. |
| `SetChecksum(h hash.Hash, sum []byte, deleteOnError bool)` | Validate the result with `h`. A mismatch fails with `ErrBadChecksum`; when `deleteOnError` is true the file is removed. Call `SetChecksum(nil, nil, false)` to disable. The `hash.Hash` must not be shared with other requests. |

## Response

`Client.Do` and `Client.DoBatch` return `*Response` values. All methods are
thread-safe and may be polled while the transfer runs; the transfer itself
happens in a background goroutine.

| Field / method | Description | Blocks until complete |
|----------------|-------------|:---------------------:|
| `Request` | The originating `*Request`. | no |
| `HTTPResponse` | The HTTP response. Its `Body` is consumed and closed by grab — do not read it. | no |
| `Filename` | Final destination path. | no |
| `Start`, `End` | Transfer start and end times. `End` is zero until completion. | no |
| `CanResume` | Server advertised `Accept-Ranges: bytes`. | no |
| `DidResume` | This transfer resumed a partial file. | no |
| `Done` | Channel closed when the transfer finalizes, successfully or not. | no |
| `IsComplete() bool` | Whether the transfer has finalized. | no |
| `Wait()` | Waits for completion. | yes |
| `Err() error` | Waits for completion, then returns the error (nil on success). Returns immediately if already complete. | yes |
| `Cancel() error` | Cancels the transfer and waits, returning the resulting error (usually `context.Canceled`). | yes |
| `Size() int64` | Total size in bytes, or `-1` if unknown and the transfer is incomplete. | no |
| `BytesComplete() int64` | Bytes written so far, including bytes resumed from a previous download. | no |
| `BytesPerSecond() float64` | Five-second moving average rate; the lifetime average once complete. | no |
| `Progress() float64` | `BytesComplete()/Size()` in `[0,1]`; `0` when the size is unknown or zero. Multiply by 100 for a percentage. | no |
| `Duration() time.Duration` | Elapsed time, or the full transfer time once complete. | no |
| `ETA() time.Time` | Projected completion time, or `End` once complete. Zero while the rate is still 0. | no |
| `Open() (io.ReadCloser, error)` | Opens the finished file, or a memory reader with `NoStore`. Caller closes it. | yes |
| `Bytes() ([]byte, error)` | Reads the entire finished content. | yes |

## Errors

| Error | Meaning |
|-------|---------|
| `ErrBadLength` | The response or the existing local file does not match the expected content length. |
| `ErrBadChecksum` | Checksum validation configured with `SetChecksum` failed. |
| `ErrNoFilename` | No filename could be derived from the URL or `Content-Disposition`. |
| `ErrFileExists` | `Request.SkipExisting` was set and the destination exists. |
| `StatusCodeError` | The server returned a status outside 2xx. Carries the code; `IsStatusCodeError` reports it. |

`StatusCodeError` also unwraps to `*httpx.HTTPError`, so either form works:

```go
var status lib.StatusCodeError
if errors.As(err, &status) {
	log.Fatalf("server returned %d", int(status))
}
```

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
			log.Printf("failed: %s (%v)", resp.Filename, resp.Err)
		} else {
			log.Printf("downloaded: %s", resp.Filename)
		}
	}
}
```

### Advanced: custom client, request context, and headers

```go
package main

import (
	"context"
	"log"

	"github.com/sebrandon1/grab/lib"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := lib.NewClient()
	urls := []string{"https://example.com/file1.zip", "https://example.com/file2.zip"}
	for _, url := range urls {
		req, err := lib.NewRequest(".", url)
		if err != nil {
			log.Printf("invalid URL %s: %v", url, err)
			continue
		}
		// WithContext returns a shallow copy, so the result must be reassigned.
		req = req.WithContext(ctx)
		req.HTTPRequest.Header.Set("Authorization", "Bearer <token>")

		resp := client.Do(req)
		if err := resp.Err(); err != nil {
			log.Printf("failed: %s (%v)", resp.Filename, err)
			continue
		}
		log.Printf("downloaded: %s (%d bytes in %s)", resp.Filename, resp.Size(), resp.Duration())
	}
}
```

### Deadline, phase timeouts, and checksum verification

```go
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/sebrandon1/grab/lib"
)

func main() {
	const (
		urlStr = "https://example.com/file.tar.gz"
		sumHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	)

	want, err := hex.DecodeString(sumHex)
	if err != nil {
		log.Fatal(err)
	}

	// Upper bound on the whole download, including the body transfer.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := lib.NewClient()
	// Phase timeouts only: dial, TLS handshake, and response headers.
	client.SetTimeout(15 * time.Second)

	req, err := lib.NewRequest("./downloads", urlStr)
	if err != nil {
		log.Fatal(err)
	}
	req = req.WithContext(ctx)
	// Delete the file when the SHA-256 does not match.
	req.SetChecksum(sha256.New(), want, true)

	resp := client.Do(req)
	if err := resp.Err(); err != nil {
		var status lib.StatusCodeError
		if errors.As(err, &status) {
			log.Fatalf("server returned HTTP %d", int(status))
		}
		log.Fatal(err)
	}
	log.Printf("verified %s (%d bytes)", resp.Filename, resp.Size())
}
```

### Rate limiting

`Request.RateLimiter` only needs `WaitN(ctx context.Context, n int) error`, which
`golang.org/x/time/rate.Limiter` already provides.

grab calls `WaitN` with the number of bytes it just read, so the buffer size must
not exceed the limiter's burst — otherwise a single read asks for more tokens
than the bucket can ever hold and the transfer fails.

```go
package main

import (
	"log"

	"github.com/sebrandon1/grab/lib"
	"golang.org/x/time/rate"
)

func main() {
	const bytesPerSecond = 200 * 1024 // 200 KiB/s

	// Burst must be at least as large as Request.BufferSize below.
	limiter := rate.NewLimiter(rate.Limit(bytesPerSecond), 64*1024)

	req, err := lib.NewRequest(".", "https://example.com/file.tar.gz")
	if err != nil {
		log.Fatal(err)
	}
	req.RateLimiter = limiter
	req.BufferSize = 4 * 1024 // small buffer paces the transfer smoothly

	resp := lib.NewClient().Do(req)
	if err := resp.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("downloaded: %s", resp.Filename)
}
```

### Progress reporting

Poll `Progress()` until `Done` closes. Do not call `Err()` inside the loop — it
waits for completion, which is exactly what the select is already doing.

```go
ticker := time.NewTicker(time.Second)
defer ticker.Stop()

for {
	select {
	case <-ticker.C:
		log.Printf("%.02f%% (%.0f B/s)", resp.Progress()*100, resp.BytesPerSecond())
	case <-resp.Done:
		if err := resp.Err(); err != nil {
			log.Fatal(err)
		}
		return
	}
}
```

## See Also

- [Main README](../README.md)
- [CLI Reference](../docs/cli-reference.md)

## License

MIT

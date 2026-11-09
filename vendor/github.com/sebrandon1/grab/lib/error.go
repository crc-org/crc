package lib

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sebrandon1/go-toolbox/httpx"
)

var (
	// ErrBadLength indicates that the server response or an existing file does
	// not match the expected content length.
	ErrBadLength = errors.New("bad content length")

	// ErrBadChecksum indicates that a downloaded file failed to pass checksum
	// validation.
	ErrBadChecksum = errors.New("checksum mismatch")

	// ErrNoFilename indicates that a reasonable filename could not be
	// automatically determined using the URL or response headers from a server.
	ErrNoFilename = errors.New("no filename could be determined")

	// ErrFileExists indicates that the destination path already exists.
	ErrFileExists = errors.New("file exists")
)

// StatusCodeError indicates that the server response had a status code that
// was not in the 200-299 range (after following any redirects).
type StatusCodeError int

func (err StatusCodeError) Error() string {
	return fmt.Sprintf("server returned %d %s", err, http.StatusText(int(err)))
}

// Unwrap exposes a StatusCodeError as the shared *httpx.HTTPError so callers
// can use errors.As with either type. The status code is carried over; the
// response body is not retained by StatusCodeError, so only the status text is
// available.
func (err StatusCodeError) Unwrap() error {
	return &httpx.HTTPError{
		StatusCode: int(err),
		Body:       http.StatusText(int(err)),
	}
}

// IsStatusCodeError returns true if the given error is of type StatusCodeError.
func IsStatusCodeError(err error) bool {
	_, ok := err.(StatusCodeError)
	return ok
}

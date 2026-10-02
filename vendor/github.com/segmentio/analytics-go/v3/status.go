package analytics

import (
	"strconv"
	"strings"
	"time"
)

// isSuccess reports whether the upload was accepted. Only 2xx counts: net/http
// follows any redirect it can, so a 3xx reaching us means it declined to (no
// Location, a 300, or a 304) and nothing was uploaded. Treating those as success
// would drop the batch silently. The TAPI endpoint does not emit 3xx at all;
// this matters when host points at a customer's proxy or redirector.
func isSuccess(status int) bool {
	return status >= 200 && status < 300
}

// retryableStatus returns whether the given HTTP status code is retryable.
func retryableStatus(status int) bool {
	switch status {
	case 408, 410, 429, 460:
		return true
	case 501, 505, 511:
		return false
	default:
		return status >= 500 && status < 600
	}
}

// parseRetryAfter parses the Retry-After header value.
// Supports integer seconds and HTTP-date format (RFC 7231 §7.1.1.1).
// Returns 0 if the value is absent, invalid, zero, or in the past.
// Caps the value at cap.
func parseRetryAfter(header string, cap int64) int64 {
	if header == "" {
		return 0
	}
	header = strings.TrimSpace(header)
	// Try integer seconds first
	n, err := strconv.ParseInt(header, 10, 64)
	if err == nil {
		if n <= 0 {
			return 0
		}
		if n > cap {
			return cap
		}
		return n
	}
	// Try HTTP-date format (RFC 7231 §7.1.1.1)
	t, err := time.Parse(time.RFC1123, header)
	if err != nil {
		// Also try RFC1123Z (with numeric timezone)
		t, err = time.Parse(time.RFC1123Z, header)
		if err != nil {
			return 0
		}
	}
	seconds := int64(time.Until(t).Seconds())
	if seconds <= 0 {
		return 0
	}
	if seconds > cap {
		return cap
	}
	return seconds
}

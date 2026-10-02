3.4.0 / 2026-09-28
==================

### Upgrade note: new request header

This release sends an `X-Retry-Count` request header on retries. If traffic to
Segment passes through a proxy, gateway or WAF that allowlists request headers,
add it before upgrading or retried uploads will be rejected. The `Authorization`
header is unchanged.

### Retry handling

* Uploads are retried on 408, 410, 429, 460, and 5xx except 501, 505 and 511.
* A `Retry-After` header is honoured on any retryable response, not only 429. Numeric seconds and the RFC 7231 HTTP-date formats are both accepted, and the value is capped at 300 seconds.
* Responses carrying `Retry-After` are retried for up to `Config.MaxRateLimitDuration` and do not consume the retry count. A `Retry-After` that will not fit in what is left of the budget ends the episode rather than being shortened: retrying inside the window the server asked for sends a request it has already declined to serve, and the budget would be spent by then anyway. `Config.MaxTotalBackoffDuration` works the same way. Other failures use exponential backoff from 500ms to a 60 second ceiling, limited by `Config.MaxRetries` and by `Config.MaxTotalBackoffDuration` as an upper bound. Exhausting a duration budget is reported as `ErrRateLimitBudgetExceeded` or `ErrBackoffBudgetExceeded`; exhausting `Config.MaxRetries` reports the last upload error instead, so a caller matching on the sentinels should not expect one for that case.
* New `Config` fields: `MaxRateLimitDuration` (default 30 minutes), `MaxTotalBackoffDuration` (default 12 hours) and `MaxRetries` (default 10).
* New `Config.ShutdownTimeout` (default 75s) bounds how long `Close` waits for retries, and the attempt issued after shutdown begins carries it as a request deadline. `Close` will not discard a batch the server has asked the client to resend, nor block for the full rate-limit budget. A request already in flight when `Close` is called is not covered: it is bounded by the HTTP client's own timeout, 10 seconds by default, and by nothing at all if a `Transport` is supplied that does not support one.
* Negative values for `MaxRetries`, `MaxTotalBackoffDuration`, `MaxRateLimitDuration` and `ShutdownTimeout` are rejected at construction. Zero means "use the default", as elsewhere in `Config`.

### Other changes

* `X-Retry-Count` is sent on retries, allowing the server to distinguish a retry from a first attempt. It is omitted on the first attempt.
* Only 2xx responses count as a successful upload. A 3xx is reported as a failed upload rather than treated as delivered, and is not retried: a redirect `net/http` has already declined to follow will not succeed on one. The Segment endpoint does not redirect, so this affects only custom `Endpoint` values.

v3.3.0 / 2023-10-31
===================

* Add groupId to context so the track events can related to both a user and distinct id and group
  * Note: When updating to this version, verify the groupId is not being found using the Extra map.  The new groupId field will now take precedence.

v3.1.0 / 2019-09-20
===================

  * add consistent panic error message
  * Expose the Message interface Validate method
  * return error if a custom type is enqueued
  * Handle pointer types in Enqueue()
  * message: update maxMessageBytes to 32KB

v3.0.1 / 2018-10-02
===================

* Migrate from Circle V1 format to Circle V2
* Adds CLI for sending segment events
* Vendor packages back-go and uuid instead of using gitsubmodules


v3.0.0 / 2016-06-02
===================

 * 3.0 is a significant rewrite with multiple breaking changes.
 * [Quickstart](https://segment.com/docs/sources/server/go/quickstart/).
 * [Documentation](https://segment.com/docs/sources/server/go/).
 * [GoDocs](https://godoc.org/gopkg.in/segmentio/analytics-go.v3).
 * [What's New in v3](https://segment.com/docs/sources/server/go/#what-s-new-in-v3).


v2.1.0 / 2015-12-28
===================

 * Add ability to set custom timestamps for messages.
 * Add ability to set a custom `net/http` client.
 * Add ability to set a custom logger.
 * Fix edge case when client would try to upload no messages.
 * Properly upload in-flight messages when client is asked to shutdown.
 * Add ability to set `.integrations` field on messages.
 * Fix resource leak with interval ticker after shutdown.
 * Add retries and back-off when uploading messages.
 * Add ability to set  custom flush interval.

v2.0.0 / 2015-02-03
===================

 * rewrite with breaking API changes

v1.2.0 / 2014-09-03
==================

 * add public .Flush() method
 * rename .Stop() to .Close()

v1.1.0 / 2014-09-02
==================

 * add client.Stop() to flash/wait. Closes #7

v1.0.0 / 2014-08-26
==================

 * fix response close
 * change comments to be more go-like
 * change uuid libraries

0.1.2 / 2014-06-11
==================

 * add runnable example
 * fix: close body

0.1.1 / 2014-05-31
==================

 * refactor locking

0.1.0 / 2014-05-22
==================

 * replace Debug option with debug package

0.0.2 / 2014-05-20
==================

 * add .Start()
 * add mutexes
 * rename BufferSize to FlushAt and FlushInterval to FlushAfter
 * lower FlushInterval to 5 seconds
 * lower BufferSize to 20 to match other clients

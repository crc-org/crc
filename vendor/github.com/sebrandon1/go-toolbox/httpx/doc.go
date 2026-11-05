// Package httpx provides small, dependency-free building blocks for writing
// HTTP API clients: client construction with sane transport defaults, request
// and response helpers, bounded body reads, typed errors, and Retry-After
// parsing.
//
// It deliberately does not know about authentication protocols, pagination
// schemes, or any particular API. Those belong in the client that uses it.
package httpx

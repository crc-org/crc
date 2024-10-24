package gosettings

// SelfValidator is an interface for a type that can validate itself.
// This is notably the case of netip.IP and netip.Prefix, and can be
// implemented by the user of this library as well.
type SelfValidator interface {
	// IsValid returns true if the value is valid, false otherwise.
	IsValid() bool
}

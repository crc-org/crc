//go:build go1.18
// +build go1.18

package gosettings

// OverrideWithComparable returns the other argument if it is not
// the zero value, otherwise it returns the existing argument.
// If used with an interface and an implementation of the interface,
// it must be instantiated with the interface type, for example:
// variable := OverrideWithComparable[Interface](variable, &implementation{})
// Avoid using this function for non-interface pointers, use OverrideWithPointer
// instead to create a new pointer.
func OverrideWithComparable[T comparable](existing, other T) (result T) { //nolint:ireturn
	var zero T
	if other == zero {
		return existing
	}
	return other
}

// OverrideWithPointer returns the existing argument if the other argument
// is nil. Otherwise it returns a new pointer to the copied value
// of the other argument value, for added mutation safety.
// To override an interface and an implementation, use OverrideWithComparable.
func OverrideWithPointer[T any](existing, other *T) (result *T) {
	if other == nil {
		return existing
	}
	result = new(T)
	*result = *other
	return result
}

// OverrideWithSlice returns the existing slice argument if the other
// slice argument is nil. Otherwise it returns a new slice with the
// copied values of the other slice argument.
// Note it is preferrable to use this function for added mutation safety
// on the result, but one can use OverrideWithSliceRaw if performance matters.
func OverrideWithSlice[T any](existing, other []T) (result []T) {
	if other == nil {
		return existing
	}
	result = make([]T, len(other))
	copy(result, other)
	return result
}

// OverrideWithValidator returns the existing argument if other is not valid,
// otherwise it returns the other argument.
func OverrideWithValidator[T SelfValidator](existing, other T) ( //nolint:ireturn
	result T) {
	if !other.IsValid() {
		return existing
	}
	return other
}

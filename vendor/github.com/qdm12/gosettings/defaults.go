//go:build go1.18
// +build go1.18

package gosettings

// DefaultComparable returns the existing argument if it is not the zero
// value, otherwise it returns the defaultValue argument.
// If used with an interface and an implementation of the interface,
// it must be instantiated with the interface type, for example:
// variable := DefaultComparable[Interface](variable, &implementation{})
// Avoid using this function for non-interface pointers, use DefaultPointer
// instead to create a new pointer.
func DefaultComparable[T comparable](existing, defaultValue T) (result T) { //nolint:ireturn
	var zero T
	if existing != zero {
		return existing
	}
	return defaultValue
}

// DefaultPointer returns the existing argument if it is not nil.
// Otherwise it returns a new pointer to the defaultValue argument.
// To default an interface to an implementation, use DefaultComparable.
func DefaultPointer[T any](existing *T, defaultValue T) (result *T) {
	if existing != nil {
		return existing
	}
	result = new(T)
	*result = defaultValue
	return result
}

// DefaultSlice returns the existing slice argument if is not nil.
// Otherwise it returns a new slice with the copied values of the
// defaultValue slice argument.
// Note it is preferrable to use this function for added mutation safety
// on the result, but one can use DefaultSliceRaw if performance matters.
func DefaultSlice[T any](existing, defaultValue []T) (result []T) {
	if existing != nil || defaultValue == nil {
		return existing
	}
	result = make([]T, len(defaultValue))
	copy(result, defaultValue)
	return result
}

// DefaultValidator returns the existing argument if it is valid,
// otherwise it returns the defaultValue argument.
func DefaultValidator[T SelfValidator](existing, defaultValue T) ( //nolint:ireturn
	result T) {
	if existing.IsValid() {
		return existing
	}
	return defaultValue
}

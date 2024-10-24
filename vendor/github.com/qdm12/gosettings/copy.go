package gosettings

import (
	"golang.org/x/exp/slices"
)

// CopyPointer returns a new pointer to the copied value of the
// original argument value.
func CopyPointer[T any](original *T) (copied *T) {
	if original == nil {
		return nil
	}
	copied = new(T)
	*copied = *original
	return copied
}

// CopySlice returns a new slice with each element of the
// original slice copied.
func CopySlice[T any](original []T) (copied []T) {
	return slices.Clone(original)
}

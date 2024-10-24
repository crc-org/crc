package gosettings

import (
	"math"
)

// BoolToYesNo returns "yes" if the given boolean is true,
// "no" if the given boolean is false, and an empty string
// if the given boolean pointer is nil.
func BoolToYesNo(b *bool) string {
	switch {
	case b == nil:
		return ""
	case *b:
		return "yes"
	default:
		return "no"
	}
}

// ObfuscateKey returns an obfuscated key for logging purposes.
// If the key is empty, `[not set]` is returned.
// If the key has up to 128 bits of security with 2 characters removed,
// it will be obfuscated as `[set]`.
// If the key has at least 128 bits of security if at least 2
// characters are removed from it, it will be obfuscated as
// `start_of_key...end_of_key`, where the start and end parts
// are each at least 1 character long, and up to 3 characters long.
// Note the security bits are calculated by assuming each unique
// character in the given key is a symbol in the alphabet,
// which gives a worst case scenario regarding the alphabet size.
// This will likely produce lower security bits estimated compared
// to the actual security bits of the key, but it's better this way
// to avoid divulging information about the key when it should not be.
// Finally, the 128 bits security is chosen because this function is
// to be used for logging purposes, which should not be exposed publicly
// either.
func ObfuscateKey(key string) (obfuscatedKey string) {
	if key == "" {
		return "[not set]"
	}

	// Guesstimate on how large the alphabet is for the key
	// given. This is the worst case scenario alphabet size
	// so it's fine to use this security-wise for this purpose.
	uniqueCharacters := make(map[rune]struct{}, len(key))
	for _, r := range key {
		uniqueCharacters[r] = struct{}{}
	}
	numberOfUniqueCharacters := len(uniqueCharacters)

	const minimumSecurityBits = 128
	minimumCharactersCount := int(math.Ceil(minimumSecurityBits / math.Log2(float64(numberOfUniqueCharacters))))

	charactersToShow := len(key) - minimumCharactersCount

	// No point showing less than 2 characters
	const minCharactersToShow = 2
	if charactersToShow < minCharactersToShow {
		// not enough security bits in original key
		// to divulge any information.
		return "[set]"
	}

	// No point showing more than 6 characters
	const maxCharactersToShow = 6
	if charactersToShow > maxCharactersToShow {
		charactersToShow = maxCharactersToShow
	}

	const numberOfParts = 2
	startCharacters := charactersToShow / numberOfParts
	endCharacters := charactersToShow - startCharacters

	startPart := key[:startCharacters]
	endPart := key[len(key)-endCharacters:]
	return startPart + "..." + endPart
}

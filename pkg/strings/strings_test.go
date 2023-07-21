package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type containsTest struct {
	input []string
	match string
	found bool
}

var containsTests = map[string]containsTest{
	"Found": {
		input: []string{"a", "b", "c"},
		match: "b",
		found: true,
	},
	"NotFound": {
		input: []string{"a", "b", "c"},
		match: "d",
		found: false,
	},
	"OneElement": {
		input: []string{"a"},
		match: "a",
		found: true,
	},
	"EmptyArray": {
		input: []string{},
		match: "a",
		found: false,
	},
	"EmptyArrayEmptySearch": {
		input: []string{},
		match: "",
		found: false,
	},
}

func TestContains(t *testing.T) {
	t.Run("Contains", func(t *testing.T) {
		for name, test := range containsTests {
			t.Run(name, func(t *testing.T) {
				found := Contains(test.input, test.match)
				assert.Equal(t, test.found, found)
			})
		}
	})
}

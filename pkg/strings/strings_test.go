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

type splitLinesTest struct {
	input       string
	splitOutput []string
}

var splitLinesTests = map[string]splitLinesTest{
	"ThreeLines": {
		input:       "line1\nline2\nline3\n",
		splitOutput: []string{"line1", "line2", "line3"},
	},
	"ThreeLinesNoFinalEOL": {
		input:       "line1\nline2\nline3",
		splitOutput: []string{"line1", "line2", "line3"},
	},
	"WindowsEOL": {
		input:       "line1\r\nline2\r\nline3\r\n",
		splitOutput: []string{"line1", "line2", "line3"},
	},
	"EmptyString": {
		input:       "",
		splitOutput: []string{},
	},
	"EOLOnly": {
		input:       "\n",
		splitOutput: []string{""},
	},
	"NoEOL": {
		input:       "line1",
		splitOutput: []string{"line1"},
	},
}

func TestSplitLines(t *testing.T) {
	t.Run("SplitLines", func(t *testing.T) {
		for name, test := range splitLinesTests {
			t.Run(name, func(t *testing.T) {
				output := SplitLines(test.input)
				assert.Equal(t, test.splitOutput, output)
			})
		}
	})
}

package pattern

import (
	"regexp"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
)

// List is a collection of compiled regular expressions.
type List []*regexp.Regexp

// NewList compiles patterns into a List.
// Returns error if any pattern is empty or invalid.
func NewList(patterns ...string) (List, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	list := make(List, 0, len(patterns))

	for _, pattern := range patterns {
		re, err := compilePattern(pattern)
		if err != nil {
			return nil, err
		}

		list = append(list, re)
	}

	return list, nil
}

// MatchFullString returns true if any regex matches the entire string.
// Pattern "test" matches "test" but not "testing" or "contest".
func (l List) MatchFullString(target string) bool {
	if len(l) == 0 {
		return false
	}

	for i := range len(l) {
		// A match spanning [0, len(target)) covers every byte of target, so the
		// matched substring is target itself — no need to allocate it for comparison.
		if loc := l[i].FindStringIndex(target); loc != nil && loc[0] == 0 && loc[1] == len(target) {
			return true
		}
	}

	return false
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, e.New("empty regular expression is not allowed")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, e.NewFrom("compile regular expression", err, fields.F("pattern", pattern))
	}

	return re, nil
}

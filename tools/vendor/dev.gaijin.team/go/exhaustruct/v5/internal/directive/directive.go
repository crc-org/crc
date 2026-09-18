package directive

import (
	"slices"
	"strings"

	"dev.gaijin.team/go/golib/e"
)

var (
	ErrEmptyDirective      = e.New("empty directive")
	ErrUnknownDirective    = e.New("unknown directive")
	ErrDuplicateDirectives = e.New("duplicate directives")
)

type Directive string

const (
	// Ignore skips checking for a specific struct literal.
	Ignore Directive = "ignore"
	// Enforce forces check even if type is excluded.
	Enforce Directive = "enforce"
	// Optional marks a field as optional.
	Optional Directive = "optional"
)

func (d Directive) IsValid() bool {
	switch d {
	case Ignore, Enforce, Optional:
		return true

	default:
		return false
	}
}

// Directives represents a collection of directives for a single line.
// Multiple directives can be specified comma-separated: //exhaustruct:enforce,optional.
type Directives []Directive

func (ds Directives) Contains(d Directive) bool {
	return slices.Contains(ds, d)
}

// directivePrefix is the exact prefix for exhaustruct directives.
// Format: //exhaustruct:<directives> [optional comment].
const directivePrefix = "//exhaustruct:"

func Parse(text string) (found bool, result Directives, errs []error) {
	text, found = strings.CutPrefix(text, directivePrefix)
	if !found {
		return false, nil, nil
	}

	if idx := strings.IndexAny(text, " \t\n"); idx > 0 {
		text = text[:idx]
	}

	if text == "" {
		return true, nil, []error{ErrEmptyDirective}
	}

	parts := strings.Split(text, ",")

	result = make(Directives, 0, len(parts))

	var hasDups bool

	for _, part := range parts {
		d := Directive(part)

		if !d.IsValid() {
			errs = append(errs, ErrUnknownDirective.WithField("directive", d))
			continue
		}

		// giving the resulting size, linear search would be most efficient
		if slices.Contains(result, d) {
			hasDups = true
			continue
		}

		result = append(result, d)
	}

	if hasDups {
		errs = append(errs, ErrDuplicateDirectives)
	}

	if len(result) == 0 {
		result = nil
	}

	return true, result, errs
}

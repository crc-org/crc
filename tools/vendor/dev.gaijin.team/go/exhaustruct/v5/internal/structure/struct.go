package structure

import (
	"go/ast"
	"go/token"
	"strings"
)

const AnonymousName = "<anonymous>"

// fieldInfo contains raw field data independent of type name.
type fieldInfo struct {
	// name is the name of the field.
	name string
	// exported indicates if the field is exported.
	exported bool `exhaustruct:"optional"`
	// enforced indicates if the field is enforced via directive.
	enforced bool `exhaustruct:"optional"`
	// optional indicates if the field is optional via directive.
	optional bool `exhaustruct:"optional"`
}

// structFields contains field information for a struct, independent of type name.
type structFields struct {
	// packagePath is the package path where fields are declared.
	packagePath string
	// fields is the list of fields in declaration order.
	fields []fieldInfo
}

type Struct struct {
	Name        string
	FullPath    string
	PackageName string

	Position token.Position `exhaustruct:"optional"`
	Fields   Fields         `exhaustruct:"optional"`

	Enforced bool `exhaustruct:"optional"`
	Ignored  bool `exhaustruct:"optional"`
	Optional bool `exhaustruct:"optional"`

	PatternEnforced bool `exhaustruct:"optional"`
	PatternIgnored  bool `exhaustruct:"optional"`
	PatternOptional bool `exhaustruct:"optional"`

	AllowEmptyDecl bool `exhaustruct:"optional"`

	// Detected via OriginScanner AST inspection before types.Unalias.
	IsAlias   bool `exhaustruct:"optional"`
	IsDerived bool `exhaustruct:"optional"`
}

// PackagePath returns the package path of the struct type.
func (s *Struct) PackagePath() string {
	if idx := strings.LastIndex(s.FullPath, "."); idx >= 0 {
		return s.FullPath[:idx]
	}

	return s.FullPath
}

// IsEnforced returns true if struct is enforced via directive or pattern.
func (s *Struct) IsEnforced() bool {
	return s.Enforced || s.PatternEnforced
}

// IsIgnored returns true if struct is ignored via directive or pattern.
func (s *Struct) IsIgnored() bool {
	return s.Ignored || s.PatternIgnored
}

// IsOptional returns true if struct is optional via directive or pattern.
func (s *Struct) IsOptional() bool {
	return s.Optional || s.PatternOptional
}

// SkippedFields returns missing required fields for a composite literal.
// callerPkgPath is used to determine if unexported fields are accessible.
// For positional literals: returns fields after the last provided element.
// For named literals: returns fields not present in the literal.
func (s *Struct) SkippedFields(lit *ast.CompositeLit, callerPkgPath string) []Field {
	externalPkg := s.Fields.PackagePath != callerPkgPath

	if isNamedLiteral(lit) {
		return s.skippedNamed(lit, externalPkg)
	}

	return s.skippedPositional(len(lit.Elts), externalPkg)
}

// isNamedLiteral checks if a composite literal uses named fields. It treats
// empty literals as not named, since positional literals checks are simpler.
func isNamedLiteral(lit *ast.CompositeLit) bool {
	if len(lit.Elts) == 0 {
		return false
	}

	_, ok := lit.Elts[0].(*ast.KeyValueExpr)

	return ok
}

func (s *Struct) skippedPositional(count int, externalPkg bool) []Field {
	items := s.Fields.Items

	if count >= len(items) {
		return nil
	}

	remaining := items[count:]
	missing := make([]Field, 0, len(remaining))

	for _, f := range remaining {
		if s.isFieldRequired(f, externalPkg) {
			missing = append(missing, f)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return missing
}

func (s *Struct) skippedNamed(lit *ast.CompositeLit, externalPkg bool) []Field {
	present := make(map[string]bool, len(lit.Elts))

	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if k, ok := kv.Key.(*ast.Ident); ok {
				present[k.Name] = true
			}
		}
	}

	missing := make([]Field, 0, len(s.Fields.Items)-len(present))

	for _, f := range s.Fields.Items {
		if !present[f.Name] && s.isFieldRequired(f, externalPkg) {
			missing = append(missing, f)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return missing
}

func (s *Struct) isFieldRequired(f Field, externalPkg bool) bool {
	// explicit field directives win over everything
	if f.Enforced {
		return true
	}

	if f.Optional {
		return false
	}

	// field-level patterns apply only when they specifically target the field —
	// i.e., the pattern matches the field path but not the struct path. A broad
	// pattern that also matches the struct is handled via s.IsEnforced/IsOptional
	// and must not silently promote unrelated fields to required/optional.
	if f.PatternEnforced && !s.PatternEnforced {
		return true
	}

	if f.PatternOptional && !s.PatternOptional {
		return false
	}

	// optionality can be inherited from the structure settings
	if s.IsOptional() {
		return false
	}

	// unexported fields are only required for same-package usage
	if externalPkg && !f.Exported {
		return false
	}

	return true
}

type Field struct {
	Name     string
	Exported bool `exhaustruct:"optional"`
	Enforced bool `exhaustruct:"optional"`
	Optional bool `exhaustruct:"optional"`

	PatternEnforced bool `exhaustruct:"optional"`
	PatternOptional bool `exhaustruct:"optional"`
}

// Fields is a collection of struct fields with shared package metadata.
// Items are in declaration order (required for positional literals).
type Fields struct {
	PackagePath string
	Items       []Field
}

func FormatFieldNames(fields []Field) string {
	switch len(fields) {
	case 0:
		return ""
	case 1:
		return fields[0].Name
	}

	var b strings.Builder

	b.Grow(len(fields))
	b.WriteString(fields[0].Name)

	for _, s := range fields[1:] {
		b.WriteString(", ")
		b.WriteString(s.Name)
	}

	return b.String()
}

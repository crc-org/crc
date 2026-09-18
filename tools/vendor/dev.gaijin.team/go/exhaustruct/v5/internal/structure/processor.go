package structure

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"dev.gaijin.team/go/exhaustruct/v5/internal/cache"
	"dev.gaijin.team/go/exhaustruct/v5/internal/directive"
	"dev.gaijin.team/go/exhaustruct/v5/internal/pattern"
)

type Processor struct {
	directives  *directive.Scanner
	origins     *OriginScanner
	fieldsCache *cache.Cache[*types.Struct, structFields]
	structCache *cache.Cache[token.Position, *Struct]

	enforce    pattern.List `exhaustruct:"optional"`
	ignore     pattern.List `exhaustruct:"optional"`
	optional   pattern.List `exhaustruct:"optional"`
	allowEmpty pattern.List `exhaustruct:"optional"`
}

type Option func(*Processor)

func WithEnforce(patterns pattern.List) Option {
	return func(p *Processor) { p.enforce = patterns }
}

func WithIgnore(patterns pattern.List) Option {
	return func(p *Processor) { p.ignore = patterns }
}

func WithOptional(patterns pattern.List) Option {
	return func(p *Processor) { p.optional = patterns }
}

func WithAllowEmpty(patterns pattern.List) Option {
	return func(p *Processor) { p.allowEmpty = patterns }
}

const cachePreallocSize = 64

func NewProcessor(directives *directive.Scanner, origins *OriginScanner, opts ...Option) *Processor {
	p := &Processor{
		directives:  directives,
		origins:     origins,
		fieldsCache: cache.New[*types.Struct, structFields](cachePreallocSize),
		structCache: cache.New[token.Position, *Struct](cachePreallocSize),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Directives returns the directive scanner the processor was constructed
// with, shared so callers can resolve use-site directives against the same
// cache.
func (p *Processor) Directives() *directive.Scanner {
	return p.directives
}

// ResolveStruct returns Struct metadata for the given type.
// Type resolution (pointers, aliases) is done by the caller.
//
// Parameters:
//   - typeName: the type's TypeName, or nil for anonymous structs
//   - strct: the underlying struct type (required)
//   - pos: position of type definition (from analyzer's AST inspection)
//   - callerPkg: package context, used for anonymous struct path
func (p *Processor) ResolveStruct(
	fset *token.FileSet,
	typeName *types.TypeName,
	strct *types.Struct,
	pos token.Pos,
	callerPkg *types.Package,
) (*Struct, []analysis.Diagnostic) {
	if strct == nil {
		return nil, nil
	}

	position := fset.Position(pos)

	// Check cache before allocating
	if position.IsValid() {
		if cached, ok := p.structCache.Get(position); ok {
			return cached, nil
		}
	}

	s := p.buildStruct(typeName, position, callerPkg)

	diags := p.populateFields(fset, s, strct)
	p.resolveStructOrigin(fset, s)

	diags = append(diags, p.resolveStructDirectives(fset, s)...)
	p.matchStructPatterns(s)

	if s.Position.IsValid() {
		p.structCache.Set(s.Position, s)
	}

	return s, diags
}

// buildStruct creates Struct metadata from type info.
func (*Processor) buildStruct(typeName *types.TypeName, pos token.Position, callerPkg *types.Package) *Struct {
	if typeName != nil {
		pkg := typeName.Pkg()

		return &Struct{
			Name:        typeName.Name(),
			FullPath:    pkg.Path() + "." + typeName.Name(),
			PackageName: pkg.Name(),
			Position:    pos,
		}
	}

	// Anonymous struct
	return &Struct{
		Name:        AnonymousName,
		FullPath:    callerPkg.Path() + "." + AnonymousName,
		PackageName: callerPkg.Name(),
		Position:    pos,
	}
}

func (p *Processor) getStructFields(fset *token.FileSet, strct *types.Struct) (structFields, []analysis.Diagnostic) {
	if fields, ok := p.fieldsCache.Get(strct); ok {
		return fields, nil
	}

	fields, diags := p.resolveStructFields(fset, strct)

	p.fieldsCache.Set(strct, fields)

	return fields, diags
}

func (p *Processor) resolveStructFields(
	fset *token.FileSet,
	strct *types.Struct,
) (structFields, []analysis.Diagnostic) {
	result := structFields{
		packagePath: "",
		fields:      make([]fieldInfo, 0, strct.NumFields()),
	}

	var diags []analysis.Diagnostic

	for f := range strct.Fields() {
		if result.packagePath == "" && f.Pkg() != nil {
			result.packagePath = f.Pkg().Path()
		}

		field := fieldInfo{
			name:     f.Name(),
			exported: f.Exported(),
		}

		if p.directives != nil {
			fieldPos := fset.Position(f.Pos())
			dirs, d := p.directives.Lookup(fset, fieldPos)

			diags = append(diags, d...)

			field.enforced = dirs.Contains(directive.Enforce)
			field.optional = dirs.Contains(directive.Optional)
		}

		result.fields = append(result.fields, field)
	}

	return result, diags
}

func (p *Processor) populateFields(fset *token.FileSet, s *Struct, strct *types.Struct) []analysis.Diagnostic {
	resolved, diags := p.getStructFields(fset, strct)

	// Fields are external when declared in a different package than the struct type.
	// This happens for derived types and aliases from external packages.
	//
	// Rationale behind that filtering is that noone except package that has declared
	// the struct can access unexported fields, therefore we can simply filter them
	// out to save up on storage. Usage of derived type from the package of structure
	// definition is simply impossible since it will cause import cycle - thus, such
	// filtering is safe.
	fieldsExternal := resolved.packagePath != s.PackagePath()

	s.Fields = Fields{
		PackagePath: resolved.packagePath,
		Items:       make([]Field, 0, len(resolved.fields)),
	}

	for _, sf := range resolved.fields {
		if fieldsExternal && !sf.exported {
			continue
		}

		fieldPath := s.FullPath + "#" + sf.name

		s.Fields.Items = append(s.Fields.Items, Field{
			Name:            sf.name,
			Exported:        sf.exported,
			Enforced:        sf.enforced,
			Optional:        sf.optional,
			PatternEnforced: p.enforce.MatchFullString(fieldPath),
			PatternOptional: p.optional.MatchFullString(fieldPath),
		})
	}

	return diags
}

func (p *Processor) resolveStructOrigin(fset *token.FileSet, s *Struct) {
	if !s.Position.IsValid() || s.Name == AnonymousName {
		return
	}

	origin := p.origins.Lookup(fset, s.Position.Filename, s.Name)

	s.IsAlias = origin == OriginAlias
	s.IsDerived = origin == OriginDerived
}

func (p *Processor) resolveStructDirectives(fset *token.FileSet, s *Struct) []analysis.Diagnostic {
	if p.directives == nil || !s.Position.IsValid() {
		return nil
	}

	dirs, diags := p.directives.Lookup(fset, s.Position)

	s.Enforced = dirs.Contains(directive.Enforce)
	s.Ignored = dirs.Contains(directive.Ignore)
	s.Optional = dirs.Contains(directive.Optional)

	return diags
}

func (p *Processor) matchStructPatterns(s *Struct) {
	s.PatternEnforced = p.enforce.MatchFullString(s.FullPath)
	s.PatternIgnored = p.ignore.MatchFullString(s.FullPath)
	s.PatternOptional = p.optional.MatchFullString(s.FullPath)
	s.AllowEmptyDecl = p.allowEmpty.MatchFullString(s.FullPath)
}

func (p *Processor) Stats() (hits, misses, size uint64) {
	return p.structCache.Stats()
}

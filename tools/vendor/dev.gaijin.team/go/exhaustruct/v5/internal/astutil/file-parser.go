// Package astutil provides AST file parsing utilities for the analyzer.
package astutil

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"golang.org/x/tools/go/analysis"
)

type ParseCallback func(fset *token.FileSet, file *ast.File) []analysis.Diagnostic

// FileParser orchestrates AST parsing by triggering registered callbacks.
// Each file is processed only once.
//
// Safe for concurrent use.
type FileParser struct {
	mu         sync.RWMutex `exhaustruct:"optional"`
	parsed     map[string]bool
	callbacks  []ParseCallback `exhaustruct:"optional"`
	parseFlags parser.Mode     `exhaustruct:"optional"`
	hits       atomic.Uint64   `exhaustruct:"optional"`
	misses     atomic.Uint64   `exhaustruct:"optional"`
}

type Option func(*FileParser)

// WithParseFlags sets parser flags for file parsing.
// Default: parser.ParseComments | parser.SkipObjectResolution.
func WithParseFlags(flags parser.Mode) Option {
	return func(p *FileParser) { p.parseFlags = flags }
}

const parsedCachePrealloc = 64

func NewFileParser(opts ...Option) *FileParser {
	p := &FileParser{
		parsed:     make(map[string]bool, parsedCachePrealloc),
		parseFlags: parser.ParseComments | parser.SkipObjectResolution,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *FileParser) OnFileParsed(cb ParseCallback) {
	p.callbacks = append(p.callbacks, cb)
}

// ProcessFiles triggers all callbacks for each provided AST file.
// Already-parsed files are skipped.
func (p *FileParser) ProcessFiles(fset *token.FileSet, files ...*ast.File) []analysis.Diagnostic {
	var allDiags []analysis.Diagnostic

	for _, file := range files {
		filename := fset.Position(file.Pos()).Filename

		p.mu.RLock()

		alreadyParsed := p.parsed[filename]
		p.mu.RUnlock()

		if alreadyParsed {
			p.hits.Add(1)

			continue
		}

		p.mu.Lock()

		if p.parsed[filename] {
			p.mu.Unlock()
			p.hits.Add(1)

			continue
		}

		p.misses.Add(1)

		for _, cb := range p.callbacks {
			allDiags = append(allDiags, cb(fset, file)...)
		}

		p.parsed[filename] = true
		p.mu.Unlock()
	}

	return allDiags
}

// ProcessFilename parses a file from disk and triggers all callbacks.
// Returns nil if already processed or if the file belongs to the Go
// distribution.
func (p *FileParser) ProcessFilename(fset *token.FileSet, filename string) []analysis.Diagnostic {
	if isGoRootFile(filename) {
		return nil
	}

	p.mu.RLock()

	alreadyParsed := p.parsed[filename]
	p.mu.RUnlock()

	if alreadyParsed {
		p.hits.Add(1)

		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.parsed[filename] {
		p.hits.Add(1)

		return nil
	}

	p.misses.Add(1)

	file, err := p.parse(fset, filename)
	if err != nil {
		p.parsed[filename] = true

		return []analysis.Diagnostic{{
			Pos:     token.NoPos,
			Message: err.Error(),
		}}
	}

	var allDiags []analysis.Diagnostic

	for _, cb := range p.callbacks {
		allDiags = append(allDiags, cb(fset, file)...)
	}

	p.parsed[filename] = true

	return allDiags
}

func (p *FileParser) Stats() (hits, misses, size uint64) {
	p.mu.RLock()

	size = uint64(len(p.parsed))

	p.mu.RUnlock()

	return p.hits.Load(), p.misses.Load(), size
}

// goRootPlaceholder is the literal prefix the compiler records instead of the
// real GOROOT for files of the Go distribution, see cmd/internal/objabi.AbsFile.
// Positions loaded from export data carry it as-is, so such paths can never be
// opened.
const goRootPlaceholder = "$GOROOT"

// isGoRootFile reports whether filename belongs to the Go distribution. Its
// sources carry no exhaustruct directives, so parsing them yields nothing and
// their definitions are treated as directive-free.
func isGoRootFile(filename string) bool {
	return hasPathPrefix(filename, goRootPlaceholder) ||
		hasPathPrefix(filename, build.Default.GOROOT)
}

// hasPathPrefix reports whether path starts with prefix at a path element
// boundary. Separators are normalized, since the compiler records slashes on
// every platform while GOROOT uses the native separator. An empty prefix, as
// an unset or root GOROOT normalizes to, matches nothing rather than
// everything.
func hasPathPrefix(path, prefix string) bool {
	prefix = strings.TrimSuffix(filepath.ToSlash(prefix), "/")
	if prefix == "" {
		return false
	}

	path = filepath.ToSlash(path)

	if !strings.HasPrefix(path, prefix) {
		return false
	}

	rest := path[len(prefix):]

	return rest == "" || rest[0] == '/'
}

func (p *FileParser) parse(fset *token.FileSet, filename string) (*ast.File, error) {
	//nolint:gosec // filename is derived from source code, not user input
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, e.NewFrom("read file", err, fields.F("filename", filename))
	}

	file, err := parser.ParseFile(fset, filename, content, p.parseFlags)
	if err != nil {
		return nil, e.NewFrom("parse file", err, fields.F("filename", filename))
	}

	return file, nil
}

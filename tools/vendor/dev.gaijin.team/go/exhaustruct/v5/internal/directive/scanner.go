package directive

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"dev.gaijin.team/go/exhaustruct/v5/internal/astutil"
	"dev.gaijin.team/go/exhaustruct/v5/internal/cache"
)

// Scanner provides thread-safe caching and lookup of file directives.
type Scanner struct {
	parser *astutil.FileParser
	cache  *cache.Cache[string, fileDirectives]
}

const cachePreallocSize = 64

// NewScanner creates a new directive scanner that registers a callback
// with the file parser to extract directives from parsed files.
func NewScanner(parser *astutil.FileParser) *Scanner {
	s := &Scanner{
		parser: parser,
		cache:  cache.New[string, fileDirectives](cachePreallocSize),
	}

	parser.OnFileParsed(s.onFileParsed)

	return s
}

func (s *Scanner) onFileParsed(fset *token.FileSet, file *ast.File) []analysis.Diagnostic {
	filename := fset.Position(file.Pos()).Filename

	fd, diags := s.parseFileDirectives(fset, file)

	s.cache.Set(filename, fd)

	return diags
}

// ProcessFiles pre-populates the cache by delegating to FileParser.ProcessFiles.
// Returns diagnostics from directive parsing.
func (s *Scanner) ProcessFiles(fset *token.FileSet, files ...*ast.File) []analysis.Diagnostic {
	return s.parser.ProcessFiles(fset, files...)
}

// Lookup returns the directives at the given source position.
// If the file is not in cache, triggers FileParser.ProcessFilename to parse it.
func (s *Scanner) Lookup(fset *token.FileSet, pos token.Position) (Directives, []analysis.Diagnostic) {
	if pos.Filename == "" {
		return nil, nil
	}

	if fd, ok := s.cache.Get(pos.Filename); ok {
		return fd[pos.Line], nil
	}

	// Cache miss - parse file (triggers onFileParsed callback, which stores
	// the result via cache.Set and increments the miss counter).
	diags := s.parser.ProcessFilename(fset, pos.Filename)

	// Peek avoids counting this self-induced read as a hit — the miss was
	// already recorded by Set above.
	if fd, ok := s.cache.Peek(pos.Filename); ok {
		return fd[pos.Line], diags
	}

	// Still not in cache means parsing failed.
	return nil, diags
}

func (s *Scanner) Stats() (hits, misses, size uint64) {
	return s.cache.Stats()
}

// fileDirectives holds directives found in a single file, indexed by line number.
type fileDirectives map[int]Directives

// parseFileDirectives parses an AST file and extracts all exhaustruct directives.
// Returns diagnostics in case file parsing errors, directive parsing errors, or
// conflicting directives for the same target line.
func (*Scanner) parseFileDirectives(fset *token.FileSet, file *ast.File) (fileDirectives, []analysis.Diagnostic) {
	directives, diagnostics := parseCommentDirectives(fset, file.Comments)

	if len(directives) == 0 {
		return nil, diagnostics
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch n.(type) {
		case nil, *ast.Comment, *ast.CommentGroup:
			return false
		}

		line := fset.Position(n.Pos()).Line
		if d, ok := directives[line]; ok {
			d.targetLine = line
			directives[line] = d
		}

		return true
	})

	result := make(fileDirectives, len(directives))

	for line, d := range directives {
		_, exists := result[d.targetLine]
		if !exists {
			result[d.targetLine] = d.directives
			continue
		}

		pos := d.pos

		// directives from block comments win over inline comments
		if line != d.targetLine {
			result[d.targetLine] = d.directives
			pos = directives[d.targetLine].pos
		}

		diagnostics = append(diagnostics, analysis.Diagnostic{
			Pos:     pos,
			Message: "directive ignored, conflicting directive already exists for the same target line",
		})
	}

	return result, diagnostics
}

type parsedDirective struct {
	pos        token.Pos
	targetLine int
	directives Directives
}

func parseCommentDirectives(
	fset *token.FileSet,
	comments []*ast.CommentGroup,
) (map[int]parsedDirective, []analysis.Diagnostic) {
	var (
		directives  = make(map[int]parsedDirective)
		diagnostics []analysis.Diagnostic
	)

	for _, cg := range comments {
		hasDirective := false

		for _, comment := range cg.List {
			found, parsed, errs := Parse(comment.Text)
			if !found {
				continue
			}

			pos := comment.Pos()

			for _, err := range errs {
				diagnostics = append(diagnostics, analysis.Diagnostic{
					Pos:     pos,
					Message: err.Error(),
				})
			}

			if len(parsed) == 0 {
				continue
			}

			if hasDirective {
				diagnostics = append(diagnostics, analysis.Diagnostic{
					Pos:     pos,
					Message: "multiple exhaustruct directives in a single comment group, ignoring",
				})

				continue
			}

			hasDirective = true
			directives[fset.Position(pos).Line] = parsedDirective{
				pos:        pos,
				targetLine: fset.Position(cg.End()).Line + 1,
				directives: parsed,
			}
		}
	}

	return directives, diagnostics
}

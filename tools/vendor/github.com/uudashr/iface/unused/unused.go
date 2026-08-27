package unused

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"slices"
	"strings"

	"github.com/uudashr/iface/internal/directive"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects interfaces which are not used anywhere in the same package where they are defined.
var Analyzer = newAnalyzer()

func newAnalyzer() *analysis.Analyzer {
	r := runner{}

	analyzer := &analysis.Analyzer{
		Name:     "unused",
		Doc:      "Detects interfaces which are not used anywhere in the same package where they are defined.",
		URL:      "https://pkg.go.dev/github.com/uudashr/iface/unused",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      r.run,
	}

	analyzer.Flags.BoolVar(&r.debug, "nerd", false, "enable nerd mode")
	analyzer.Flags.StringVar(&r.exclude, "exclude", "", "comma-separated list of packages to exclude from the check")

	return analyzer
}

type ifaceEntry struct {
	ts   *ast.TypeSpec
	decl *ast.GenDecl
}

type runner struct {
	debug   bool
	exclude string
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	var excludes []string

	if r.exclude != "" {
		for _, pkg := range strings.Split(r.exclude, ",") {
			if p := strings.TrimSpace(pkg); p != "" {
				excludes = append(excludes, p)
			}
		}
	}

	if slices.Contains(excludes, pass.Pkg.Path()) {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Collect all interface type declarations
	ifaces := make(map[string]ifaceEntry)

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		decl, ok := n.(*ast.GenDecl)
		if !ok {
			return
		}

		if r.debug {
			fmt.Fprintf(os.Stderr, "GenDecl: %v specs=%d\n", decl.Tok, len(decl.Specs))
		}

		if decl.Tok != token.TYPE {
			return
		}

		if directive.ShouldIgnore(decl.Doc, pass.Analyzer.Name) {
			return
		}

		for i, spec := range decl.Specs {
			r.debugf(" spec[%d]: %v %T\n", i, spec, spec)

			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			_, ok = ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			r.debugln("  -> Interface type declaration:", ts.Name.Name, ts.Pos())

			if directive.ShouldIgnore(ts.Doc, pass.Analyzer.Name) {
				continue
			}

			ifaces[ts.Name.Name] = ifaceEntry{ts: ts, decl: decl}
		}
	})

	if r.debug {
		var ifaceNames []string
		for name := range ifaces {
			ifaceNames = append(ifaceNames, name)
		}

		fmt.Fprintln(os.Stderr, "Declared interfaces:", ifaceNames)
	}

	// Inspect whether the interface is used within the package
	nodeFilter = []ast.Node{
		(*ast.Ident)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return
		}

		entry, ok := ifaces[ident.Name]
		if !ok {
			return
		}

		if entry.ts.Pos() == ident.Pos() {
			return
		}

		delete(ifaces, ident.Name)
	})

	if r.debug {
		fmt.Fprintf(os.Stderr, "Package %s %s\n", pass.Pkg.Path(), pass.Pkg.Name())
	}

	for name, entry := range ifaces {
		ts := entry.ts
		decl := entry.decl

		var start, end token.Pos
		if len(decl.Specs) == 1 {
			start = decl.Pos()
			if decl.Doc != nil {
				start = decl.Doc.Pos()
			}

			end = decl.End()
		} else {
			start = ts.Pos()
			if ts.Doc != nil {
				start = ts.Doc.Pos()
			}

			end = ts.End()
		}

		msg := fmt.Sprintf("interface '%s' is declared but not used within the package", name)
		pass.Report(analysis.Diagnostic{
			Pos:     ts.Pos(),
			Message: msg,
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Remove the unused interface declaration",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     start,
							End:     end,
							NewText: []byte{},
						},
					},
				},
			},
		})
	}

	return nil, nil
}

func (r *runner) debugln(a ...any) {
	if r.debug {
		fmt.Fprintln(os.Stderr, a...)
	}
}

func (r *runner) debugf(format string, a ...any) {
	if r.debug {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

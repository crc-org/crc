// Package exptostd It is an analyzer that detects functions from golang.org/x/exp/ that can be replaced by std functions.
package exptostd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	go123   = 123
	go121   = 121
	goDevel = 666
)

// Result is step analysis results.
type Result struct {
	shouldKeepImport bool
	Diagnostics      []analysis.Diagnostic
}

type stdReplacement struct {
	MinGo     int
	Text      string
	Suggested func(callExpr *ast.CallExpr) (analysis.SuggestedFix, error)
}

type analyzer struct {
	mapsPkgReplacements   map[string]stdReplacement
	slicesPkgReplacements map[string]stdReplacement

	skipGoVersionDetection bool
	goVersion              int
}

// NewAnalyzer create a new Analyzer.
func NewAnalyzer() *analysis.Analyzer {
	_, skip := os.LookupEnv("EXPTOSTD_SKIP_GO_VERSION_CHECK")

	l := &analyzer{
		skipGoVersionDetection: skip,
		mapsPkgReplacements: map[string]stdReplacement{
			"Keys":       {MinGo: go123, Text: "slices.Collect(maps.Keys())", Suggested: suggestedFixForKeysOrValues},
			"Values":     {MinGo: go123, Text: "slices.Collect(maps.Values())", Suggested: suggestedFixForKeysOrValues},
			"Equal":      {MinGo: go121, Text: "maps.Equal()"},
			"EqualFunc":  {MinGo: go121, Text: "maps.EqualFunc()"},
			"Clone":      {MinGo: go121, Text: "maps.Clone()"},
			"Copy":       {MinGo: go121, Text: "maps.Copy()"},
			"DeleteFunc": {MinGo: go121, Text: "maps.DeleteFunc()"},
			"Clear":      {MinGo: go121, Text: "clear()", Suggested: suggestedFixForClear},
		},
		slicesPkgReplacements: map[string]stdReplacement{
			"Equal":        {MinGo: go121, Text: "slices.Equal()"},
			"EqualFunc":    {MinGo: go121, Text: "slices.EqualFunc()"},
			"Compare":      {MinGo: go121, Text: "slices.Compare()"},
			"CompareFunc":  {MinGo: go121, Text: "slices.CompareFunc()"},
			"Index":        {MinGo: go121, Text: "slices.Index()"},
			"IndexFunc":    {MinGo: go121, Text: "slices.IndexFunc()"},
			"Contains":     {MinGo: go121, Text: "slices.Contains()"},
			"ContainsFunc": {MinGo: go121, Text: "slices.ContainsFunc()"},
			"Insert":       {MinGo: go121, Text: "slices.Insert()"},
			"Delete":       {MinGo: go121, Text: "slices.Delete()"},
			"DeleteFunc":   {MinGo: go121, Text: "slices.DeleteFunc()"},
			"Replace":      {MinGo: go121, Text: "slices.Replace()"},
			"Clone":        {MinGo: go121, Text: "slices.Clone()"},
			"Compact":      {MinGo: go121, Text: "slices.Compact()"},
			"CompactFunc":  {MinGo: go121, Text: "slices.CompactFunc()"},
			"Grow":         {MinGo: go121, Text: "slices.Grow()"},
			"Clip":         {MinGo: go121, Text: "slices.Clip()"},
			"Reverse":      {MinGo: go121, Text: "slices.Reverse()"},

			"Sort":             {MinGo: go121, Text: "slices.Sort()"},
			"SortFunc":         {MinGo: go121, Text: "slices.SortFunc()"},
			"SortStableFunc":   {MinGo: go121, Text: "slices.SortStableFunc()"},
			"IsSorted":         {MinGo: go121, Text: "slices.IsSorted()"},
			"IsSortedFunc":     {MinGo: go121, Text: "slices.IsSortedFunc()"},
			"Min":              {MinGo: go121, Text: "slices.Min()"},
			"MinFunc":          {MinGo: go121, Text: "slices.MinFunc()"},
			"Max":              {MinGo: go121, Text: "slices.Max()"},
			"MaxFunc":          {MinGo: go121, Text: "slices.MaxFunc()"},
			"BinarySearch":     {MinGo: go121, Text: "slices.BinarySearch()"},
			"BinarySearchFunc": {MinGo: go121, Text: "slices.BinarySearchFunc()"},
		},
	}

	return &analysis.Analyzer{
		Name:     "exptostd",
		Doc:      "Detects functions from golang.org/x/exp/ that can be replaced by std functions.",
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	a.goVersion = getGoVersion(pass)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.ImportSpec)(nil),
	}

	imports := map[string]*ast.ImportSpec{}

	var shouldKeepExpMaps bool

	var resultExpSlices Result

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if importSpec, ok := n.(*ast.ImportSpec); ok {
			// skip aliases
			if importSpec.Name == nil || importSpec.Name.Name == "" {
				imports[trimImportPath(importSpec)] = importSpec
			}

			return
		}

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return
		}

		switch ident.Name {
		case "maps":
			diagnostic, usage := a.detectPackageUsage(pass, a.mapsPkgReplacements, selExpr, ident, callExpr, "golang.org/x/exp/maps")
			if usage {
				pass.Report(diagnostic)
			}

			shouldKeepExpMaps = shouldKeepExpMaps || !usage

		case "slices":
			diagnostic, usage := a.detectPackageUsage(pass, a.slicesPkgReplacements, selExpr, ident, callExpr, "golang.org/x/exp/slices")

			if usage {
				resultExpSlices.Diagnostics = append(resultExpSlices.Diagnostics, diagnostic)
			}

			resultExpSlices.shouldKeepImport = resultExpSlices.shouldKeepImport || !usage
		}
	})

	a.suggestReplaceImport(pass, imports, shouldKeepExpMaps, "golang.org/x/exp/maps")

	if resultExpSlices.shouldKeepImport {
		for _, diagnostic := range resultExpSlices.Diagnostics {
			pass.Report(diagnostic)
		}
	} else {
		a.suggestReplaceImport(pass, imports, resultExpSlices.shouldKeepImport, "golang.org/x/exp/slices")
	}

	return nil, nil
}

func (a *analyzer) detectPackageUsage(pass *analysis.Pass,
	replacements map[string]stdReplacement,
	selExpr *ast.SelectorExpr, ident *ast.Ident, callExpr *ast.CallExpr,
	importPath string,
) (analysis.Diagnostic, bool) {
	rp, ok := replacements[selExpr.Sel.Name]
	if !ok {
		return analysis.Diagnostic{}, false
	}

	if !a.skipGoVersionDetection && rp.MinGo > a.goVersion {
		return analysis.Diagnostic{}, false
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return analysis.Diagnostic{}, false
	}

	pkg, ok := obj.(*types.PkgName)
	if !ok {
		return analysis.Diagnostic{}, false
	}

	if pkg.Imported().Path() != importPath {
		return analysis.Diagnostic{}, false
	}

	diagnostic := analysis.Diagnostic{
		Pos:     callExpr.Pos(),
		Message: fmt.Sprintf("%s.%s() can be replaced by %s", importPath, selExpr.Sel.Name, rp.Text),
	}

	if rp.Suggested != nil {
		fix, err := rp.Suggested(callExpr)
		if err != nil {
			diagnostic.Message = fmt.Sprintf("Suggested fix error: %v", err)
		} else {
			diagnostic.SuggestedFixes = append(diagnostic.SuggestedFixes, fix)
		}
	}

	return diagnostic, true
}

func (a *analyzer) suggestReplaceImport(pass *analysis.Pass, imports map[string]*ast.ImportSpec, shouldKeep bool, importPath string) {
	imp, ok := imports[importPath]
	if !ok || shouldKeep {
		return
	}

	src := trimImportPath(imp)

	index := strings.LastIndex(src, "/")

	pass.Report(analysis.Diagnostic{
		Pos:     imp.Pos(),
		End:     imp.End(),
		Message: fmt.Sprintf("Import statement '%s' can be replaced by '%s'", src, src[index+1:]),
		SuggestedFixes: []analysis.SuggestedFix{{
			TextEdits: []analysis.TextEdit{{
				Pos:     imp.Path.Pos(),
				End:     imp.Path.End(),
				NewText: []byte(string(imp.Path.Value[0]) + src[index+1:] + string(imp.Path.Value[0])),
			}},
		}},
	})
}

func suggestedFixForClear(callExpr *ast.CallExpr) (analysis.SuggestedFix, error) {
	s := &ast.CallExpr{
		Fun:      ast.NewIdent("clear"),
		Args:     callExpr.Args,
		Ellipsis: callExpr.Ellipsis,
	}

	buf := bytes.NewBuffer(nil)

	err := printer.Fprint(buf, token.NewFileSet(), s)
	if err != nil {
		return analysis.SuggestedFix{}, fmt.Errorf("print suggested fix: %w", err)
	}

	return analysis.SuggestedFix{
		TextEdits: []analysis.TextEdit{{
			Pos:     callExpr.Pos(),
			End:     callExpr.End(),
			NewText: buf.Bytes(),
		}},
	}, nil
}

func suggestedFixForKeysOrValues(callExpr *ast.CallExpr) (analysis.SuggestedFix, error) {
	s := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "slices"},
			Sel: &ast.Ident{Name: "Collect"},
		},
		Args: []ast.Expr{callExpr},
	}

	buf := bytes.NewBuffer(nil)

	err := printer.Fprint(buf, token.NewFileSet(), s)
	if err != nil {
		return analysis.SuggestedFix{}, fmt.Errorf("print suggested fix: %w", err)
	}

	return analysis.SuggestedFix{
		TextEdits: []analysis.TextEdit{{
			Pos:     callExpr.Pos(),
			End:     callExpr.End(),
			NewText: buf.Bytes(),
		}},
	}, nil
}

func getGoVersion(pass *analysis.Pass) int {
	// Prior to go1.22, versions.FileVersion returns only the toolchain version,
	// which is of no use to us,
	// so disable this analyzer on earlier versions.
	if !slices.Contains(build.Default.ReleaseTags, "go1.22") {
		return 0 // false
	}

	pkgVersion := pass.Pkg.GoVersion()
	if pkgVersion == "" {
		// Empty means Go devel.
		return goDevel // true
	}

	raw := strings.TrimPrefix(pkgVersion, "go")

	// prerelease version (go1.24rc1)
	idx := strings.IndexFunc(raw, func(r rune) bool {
		return (r < '0' || r > '9') && r != '.'
	})

	if idx != -1 {
		raw = raw[:idx]
	}

	vParts := strings.Split(raw, ".")

	v, err := strconv.Atoi(strings.Join(vParts[:2], ""))
	if err != nil {
		v = 116
	}

	return v
}

func trimImportPath(spec *ast.ImportSpec) string {
	return spec.Path.Value[1 : len(spec.Path.Value)-1]
}

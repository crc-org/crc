package analyzer

import (
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// runTagMigration scans struct definitions for deprecated exhaustruct tags
// and emits migration diagnostics with suggested fixes, using inspector to
// traverse StructType nodes efficiently.
func runTagMigration(pass *analysis.Pass) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert

	insp.Preorder([]ast.Node{new(ast.StructType)}, func(n ast.Node) {
		visitStructType(pass, n)
	})
}

func visitStructType(pass *analysis.Pass, n ast.Node) {
	st, ok := n.(*ast.StructType)
	if !ok {
		return
	}

	if st.Fields == nil {
		return
	}

	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}

		value, ok := parseExhaustructTag(field.Tag.Value)
		if !ok {
			continue
		}

		pass.Report(buildTagDiagnostic(field, value))
	}
}

const exhaustructTagKey = "exhaustruct"

// parseExhaustructTag extracts value from `exhaustruct:"value"` tag.
// Returns ("", false) if tag not present.
func parseExhaustructTag(tagLiteral string) (string, bool) {
	if len(tagLiteral) < 2 { //nolint:mnd
		return "", false
	}

	// Strip backticks
	inner := tagLiteral[1 : len(tagLiteral)-1]

	return reflect.StructTag(inner).Lookup(exhaustructTagKey)
}

func buildTagDiagnostic(field *ast.Field, tagValue string) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos:            field.Tag.Pos(),
		Message:        `struct tag "exhaustruct" is not supported anymore, use comment directives`,
		SuggestedFixes: []analysis.SuggestedFix{buildTagFix(field, tagValue)},
	}
}

func buildTagFix(field *ast.Field, tagValue string) analysis.SuggestedFix {
	tag := field.Tag
	newTag := removeExhaustructFromTag(tag.Value)

	if tagValue == "optional" {
		if newTag != "" {
			newTag += " "
		}

		newTag += "//exhaustruct:optional"
	}

	// Calculate start position (include leading space if removing entirely)
	startPos := tag.Pos()
	if newTag == "" {
		startPos = field.Type.End()
	}

	return analysis.SuggestedFix{
		Message: "fix",
		TextEdits: []analysis.TextEdit{{
			Pos:     startPos,
			End:     tag.End(),
			NewText: []byte(newTag),
		}},
	}
}

var exhaustructTagPattern = regexp.MustCompile(`\s*exhaustruct:"[^"]*"`)

func removeExhaustructFromTag(tagLiteral string) string {
	tagLiteral = tagLiteral[1 : len(tagLiteral)-1]
	tagLiteral = strings.TrimSpace(exhaustructTagPattern.ReplaceAllString(tagLiteral, ""))

	if tagLiteral == "" {
		return ""
	}

	return "`" + tagLiteral + "`"
}

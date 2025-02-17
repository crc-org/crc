package analyzer

import (
	"golang.org/x/tools/go/analysis"
)

func newAnalysisDiagnostic(
	checker string,
	analysisRange analysis.Range,
	message string,
	suggestedFixes []analysis.SuggestedFix,
) *analysis.Diagnostic {
	d := analysis.Diagnostic{
		Pos:            analysisRange.Pos(),
		End:            analysisRange.End(),
		SuggestedFixes: suggestedFixes,
	}
	if checker != "" {
		d.Category = checker
		d.Message = checker + ": " + message
	} else {
		d.Message = message
	}
	return &d
}

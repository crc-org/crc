// Package max_len provides a checker for maximum line length of godocs.
package max_len

import (
	"fmt"
	gdc "go/doc/comment"
	"strings"
	"unicode/utf8"

	"github.com/godoc-lint/godoc-lint/pkg/model"
	"github.com/godoc-lint/godoc-lint/pkg/util"
)

const maxLenRule = model.MaxLenRule

var ruleSet = model.RuleSet{}.Add(maxLenRule)

// MaxLenChecker checks maximum line length of godocs.
type MaxLenChecker struct{}

// NewMaxLenChecker returns a new instance of the corresponding checker.
func NewMaxLenChecker() *MaxLenChecker {
	return &MaxLenChecker{}
}

// GetCoveredRules implements the corresponding interface method.
func (r *MaxLenChecker) GetCoveredRules() model.RuleSet {
	return ruleSet
}

// Apply implements the corresponding interface method.
func (r *MaxLenChecker) Apply(actx *model.AnalysisContext) error {
	includeTests := actx.Config.GetRuleOptions().MaxLenIncludeTests
	maxLen := int(actx.Config.GetRuleOptions().MaxLenLength)

	docs := make(map[*model.CommentGroup]struct{}, 10*len(actx.InspectorResult.Files))

	for _, ir := range util.AnalysisApplicableFiles(actx, includeTests, model.RuleSet{}.Add(maxLenRule)) {
		if ir.PackageDoc != nil {
			docs[ir.PackageDoc] = struct{}{}
		}

		for _, sd := range ir.SymbolDecl {
			if sd.ParentDoc != nil {
				docs[sd.ParentDoc] = struct{}{}
			}
			if sd.Doc == nil {
				continue
			}
			docs[sd.Doc] = struct{}{}
		}
	}

	for doc := range docs {
		checkMaxLen(actx, doc, maxLen)
	}
	return nil
}

func checkMaxLen(actx *model.AnalysisContext, doc *model.CommentGroup, maxLen int) {
	if doc.DisabledRules.All || doc.DisabledRules.Rules.Has(maxLenRule) {
		return
	}

	linkDefsMap := make(map[string]struct{}, len(doc.Parsed.Links))
	for _, linkDef := range doc.Parsed.Links {
		linkDefLine := fmt.Sprintf("[%s]: %s", linkDef.Text, linkDef.URL)
		linkDefsMap[linkDefLine] = struct{}{}
	}

	nonCodeBlocks := make([]gdc.Block, 0, len(doc.Parsed.Content))
	for _, b := range doc.Parsed.Content {
		if _, ok := b.(*gdc.Code); ok {
			continue
		}
		nonCodeBlocks = append(nonCodeBlocks, b)
	}
	strippedCodeAndLinks := &gdc.Doc{
		Content: nonCodeBlocks,
	}
	text := string((&gdc.Printer{}).Comment(strippedCodeAndLinks))
	linesIter := strings.SplitSeq(removeCarriageReturn(text), "\n")

	for l := range linesIter {
		lineLen := utf8.RuneCountInString(l)
		if lineLen <= maxLen {
			continue
		}
		actx.Pass.ReportRangef(&doc.CG, "godoc line is too long (%d > %d)", lineLen, maxLen)
		break
	}
}

func removeCarriageReturn(s string) string {
	return strings.ReplaceAll(s, "\r", "")
}

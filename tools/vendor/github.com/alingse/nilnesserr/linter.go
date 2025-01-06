package nilnesserr

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

const (
	linterName = "nilnesserr"
	linterDoc  = `This linter reports that it checks for err != nil, but it returns a different nil value error.
powered by nilness and nilerr.`

	linterMessage = "return a nil value error after check error"
)

type LinterSetting struct{}

func NewAnalyzer(setting LinterSetting) (*analysis.Analyzer, error) {
	a, err := newAnalyzer(setting)
	if err != nil {
		return nil, err
	}

	return &analysis.Analyzer{
		Name: linterName,
		Doc:  linterDoc,
		Run:  a.run,
		Requires: []*analysis.Analyzer{
			buildssa.Analyzer,
		},
	}, nil
}

type analyzer struct {
	setting LinterSetting
}

func newAnalyzer(setting LinterSetting) (*analyzer, error) {
	a := &analyzer{setting: setting}

	return a, nil
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	_, _ = a.checkNilnesserr(pass)

	return nil, nil
}

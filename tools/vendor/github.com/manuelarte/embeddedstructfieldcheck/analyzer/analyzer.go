package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/manuelarte/embeddedstructfieldcheck/internal"
)

const ForbidMutexName = "forbid-mutex"

func NewAnalyzer() *analysis.Analyzer {
	var forbidMutex bool

	a := &analysis.Analyzer{
		Name: "embeddedstructfieldcheck",
		Doc: "Embedded types should be at the top of the field list of a struct, " +
			"and there must be an empty line separating embedded fields from regular fields.",
		URL: "https://github.com/manuelarte/embeddedstructfieldcheck",
		Run: func(pass *analysis.Pass) (any, error) {
			run(pass, forbidMutex)

			//nolint:nilnil // impossible case.
			return nil, nil
		},
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.BoolVar(&forbidMutex, ForbidMutexName, false, "Checks that sync.Mutex is not used as an embedded field.")

	return a
}

func run(pass *analysis.Pass, forbidMutex bool) {
	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		return
	}

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		node, ok := n.(*ast.StructType)
		if !ok {
			return
		}

		internal.Analyze(pass, node, forbidMutex)
	})
}

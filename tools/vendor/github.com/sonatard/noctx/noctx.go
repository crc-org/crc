package noctx

import (
	"fmt"
	"maps"
	"slices"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var Analyzer = &analysis.Analyzer{
	Name:             "noctx",
	Doc:              "noctx finds sending http request without context.Context",
	Run:              Run,
	RunDespiteErrors: false,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
	ResultType: nil,
	FactTypes:  nil,
}

func Run(pass *analysis.Pass) (interface{}, error) {
	ngFuncMessages := map[string]string{
		// net/http
		"net/http.Get":                "must not be called. use net/http.NewRequestWithContext and (*net/http.Client).Do(*http.Request)",
		"net/http.Head":               "must not be called. use net/http.NewRequestWithContext and (*net/http.Client).Do(*http.Request)",
		"net/http.Post":               "must not be called. use net/http.NewRequestWithContext and (*net/http.Client).Do(*http.Request)",
		"net/http.PostForm":           "must not be called. use net/http.NewRequestWithContext and (*net/http.Client).Do(*http.Request)",
		"(*net/http.Client).Get":      "must not be called. use (*net/http.Client).Do(*http.Request)",
		"(*net/http.Client).Head":     "must not be called. use (*net/http.Client).Do(*http.Request)",
		"(*net/http.Client).Post":     "must not be called. use (*net/http.Client).Do(*http.Request)",
		"(*net/http.Client).PostForm": "must not be called. use (*net/http.Client).Do(*http.Request)",
		"net/http.NewRequest":         "must not be called. use net/http.NewRequestWithContext",

		// database/sql
		"(*database/sql.DB).Begin":      "must not be called. use (*database/sql.DB).BeginTx",
		"(*database/sql.DB).Exec":       "must not be called. use (*database/sql.DB).ExecContext",
		"(*database/sql.DB).Ping":       "must not be called. use (*database/sql.DB).PingContext",
		"(*database/sql.DB).Prepare":    "must not be called. use (*database/sql.DB).PrepareContext",
		"(*database/sql.DB).Query":      "must not be called. use (*database/sql.DB).QueryContext",
		"(*database/sql.DB).QueryRow":   "must not be called. use (*database/sql.DB).QueryRowContext",
		"(*database/sql.Tx).Exec":       "must not be called. use (*database/sql.Tx).ExecContext",
		"(*database/sql.Tx).Prepare":    "must not be called. use (*database/sql.Tx).PrepareContext",
		"(*database/sql.Tx).Query":      "must not be called. use (*database/sql.Tx).QueryContext",
		"(*database/sql.Tx).QueryRow":   "must not be called. use (*database/sql.Tx).QueryRowContext",
		"(*database/sql.Tx).Stmt":       "must not be called. use (*database/sql.Tx).StmtContext",
		"(*database/sql.Stmt).Exec":     "must not be called. use (*database/sql.Conn).ExecContext",
		"(*database/sql.Stmt).Query":    "must not be called. use (*database/sql.Conn).QueryContext",
		"(*database/sql.Stmt).QueryRow": "must not be called. use (*database/sql.Conn).QueryRowContext",
	}

	ngFuncs := typeFuncs(pass, slices.Collect(maps.Keys(ngFuncMessages)))
	if len(ngFuncs) == 0 {
		return nil, nil
	}

	ssa, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if !ok {
		panic(fmt.Sprintf("%T is not *buildssa.SSA", pass.ResultOf[buildssa.Analyzer]))
	}

	for _, sf := range ssa.SrcFuncs {
		for _, b := range sf.Blocks {
			for _, instr := range b.Instrs {
				for _, ngFunc := range ngFuncs {
					if analysisutil.Called(instr, nil, ngFunc) {
						pass.Reportf(instr.Pos(), "%s %s", ngFunc.FullName(), ngFuncMessages[ngFunc.FullName()])

						break
					}
				}
			}
		}
	}

	return nil, nil
}

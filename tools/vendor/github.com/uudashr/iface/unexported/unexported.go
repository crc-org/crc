package unexported

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"strings"

	"github.com/uudashr/iface/internal/directive"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects unexported interfaces used in exported functions or methods.
var Analyzer = newAnalyzer()

func newAnalyzer() *analysis.Analyzer {
	r := runner{}

	analyzer := &analysis.Analyzer{
		Name:     "unexported",
		Doc:      "Detects interfaces which are not exported but are used as parameters or return values in exported functions or methods.",
		URL:      "https://pkg.go.dev/github.com/uudashr/iface/unexported",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      r.run,
	}

	analyzer.Flags.BoolVar(&r.debug, "nerd", false, "enable nerd mode")

	return analyzer
}

type runner struct {
	debug bool
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		r.debugln("FuncDecl:", funcDecl.Name.Name)

		if directive.ShouldIgnore(funcDecl.Doc, pass.Analyzer.Name) {
			return
		}

		var recvName string

		if recv := funcDecl.Recv; recv != nil {
			recvType := recv.List[0].Type

			if r.debug {
				infoType := pass.TypesInfo.TypeOf(recvType)
				fmt.Fprintf(os.Stderr, " recvType: %v infoType: %v reflectType: %T\n", recvType, infoType, recvType)
			}

			inner := recvType
			if star, ok := inner.(*ast.StarExpr); ok {
				inner = star.X
			}

			recvName = typeName(inner)
		}

		if !funcDecl.Name.IsExported() {
			// skip unexported functions
			r.debugln(" skip non-exported")

			return
		}

		r.debugln(" params:")

		for _, param := range funcDecl.Type.Params.List {
			r.checkType(pass, param.Type, funcDecl, recvName, "parameter")
		}

		r.debugln(" results:")

		if funcDecl.Type.Results == nil {
			r.debugln("  no results")

			return
		}

		for _, result := range funcDecl.Type.Results.List {
			r.checkType(pass, result.Type, funcDecl, recvName, "return value")
		}
	})

	return nil, nil
}

// findIdent recursively unwraps StarExpr, Ellipsis, and ArrayType to locate the innermost
// *ast.Ident or *ast.SelectorExpr. Returns nil if not found.
func findIdent(expr ast.Expr) ast.Expr {
	for {
		switch e := expr.(type) {
		case *ast.StarExpr:
			expr = e.X
		case *ast.Ellipsis:
			expr = e.Elt
		case *ast.ArrayType:
			expr = e.Elt
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		case *ast.ChanType:
			expr = e.Value
		case *ast.MapType:
			expr = e.Value
		case *ast.Ident, *ast.SelectorExpr:
			return e
		default:
			return nil
		}
	}
}

// typeName reconstructs the type string as written in source from an AST expression.
func typeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + typeName(e.Elt)
		}

		return "[" + typeName(e.Len) + "]" + typeName(e.Elt)
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + typeName(e.X)
	case *ast.Ellipsis:
		return "..." + typeName(e.Elt)
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return "chan<- " + typeName(e.Value)
		case ast.RECV:
			return "<-chan " + typeName(e.Value)
		default:
			return "chan " + typeName(e.Value)
		}
	case *ast.MapType:
		return "map[" + typeName(e.Key) + "]" + typeName(e.Value)
	case *ast.SelectorExpr:
		return typeName(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return typeName(e.X) + "[" + typeName(e.Index) + "]"
	case *ast.IndexListExpr:
		var b strings.Builder
		b.WriteString(typeName(e.X))
		b.WriteByte('[')

		for i, idx := range e.Indices {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(typeName(idx))
		}

		b.WriteByte(']')

		return b.String()
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func (r *runner) checkType(pass *analysis.Pass, expr ast.Expr, funcDecl *ast.FuncDecl, recvName, role string) {
	infoType := pass.TypesInfo.TypeOf(expr)

	r.debugf("  %s: %v infoType: %v reflectType: %T\n", role, expr, infoType, expr)

	ident := findIdent(expr)
	if ident == nil {
		r.debugln("   skip non-interface")

		return
	}

	switch typ := ident.(type) {
	case *ast.SelectorExpr:
		r.debugln("   external")

		return
	case *ast.Ident:
		if typ.IsExported() {
			r.debugln("   skip exported")

			return
		}
	}

	identType := pass.TypesInfo.TypeOf(ident)
	if identType == nil {
		r.debugln("   skip unknown type")

		return
	}

	errorType := types.Universe.Lookup("error").Type()
	anyType := types.Universe.Lookup("any").Type()

	if types.Identical(identType, errorType) || types.Identical(identType, anyType) {
		r.debugln("   skip predeclared type")

		return
	}

	if !types.IsInterface(identType) {
		r.debugln("   skip non-interface")

		return
	}

	typ := ident.(*ast.Ident)

	r.debugln("   unexported")

	funcMethod := "function"
	funcMethodName := funcDecl.Name.Name

	if recvName != "" {
		funcMethod = "method"
		funcMethodName = recvName + "." + funcDecl.Name.Name
	}

	pass.Report(analysis.Diagnostic{
		Pos:     typ.Pos(),
		Message: fmt.Sprintf("unexported interface '%s' used as %s in exported %s '%s'", typeName(expr), role, funcMethod, funcMethodName),
	})
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

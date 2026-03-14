package halstead

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// AnalyzeAST parses Go source with the standard library parser and derives metrics
// from the resulting syntax tree.
func AnalyzeAST(src []byte) (Metrics, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "", src, parser.ParseComments)
	if err != nil {
		return Metrics{}, err
	}

	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Types:      map[ast.Expr]types.TypeAndValue{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}

	config := &types.Config{
		Importer: importer.Default(),
	}
	if _, err := config.Check("halstead", fileSet, []*ast.File{file}, info); err != nil {
		return Metrics{}, err
	}

	metrics := Metrics{
		Name:      "go-ast-types",
		Operators: map[string]int{},
		Operands:  map[string]int{},
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			metrics.addOperator(node.Tok.String(), len(node.Lhs))
		case *ast.BinaryExpr:
			metrics.addOperator(node.Op.String(), 1)
		case *ast.UnaryExpr:
			metrics.addOperator(node.Op.String(), 1)
		case *ast.IncDecStmt:
			metrics.addOperator(node.Tok.String(), 1)
		case *ast.BranchStmt:
			metrics.addOperator(node.Tok.String(), 1)
		case *ast.ReturnStmt:
			metrics.addOperator("return", 1)
		case *ast.IfStmt:
			metrics.addOperator("if", 1)
		case *ast.ForStmt:
			metrics.addOperator("for", 1)
		case *ast.RangeStmt:
			metrics.addOperator(node.Tok.String(), 1)
			metrics.addOperator("range", 1)
		case *ast.SwitchStmt:
			metrics.addOperator("switch", 1)
		case *ast.TypeSwitchStmt:
			metrics.addOperator("type-switch", 1)
		case *ast.SelectStmt:
			metrics.addOperator("select", 1)
		case *ast.GoStmt:
			metrics.addOperator("go", 1)
		case *ast.DeferStmt:
			metrics.addOperator("defer", 1)
		case *ast.SendStmt:
			metrics.addOperator("<-", 1)
		case *ast.CallExpr:
			metrics.addOperator("call", 1)
		case *ast.IndexExpr:
			metrics.addOperator("index", 1)
		case *ast.IndexListExpr:
			metrics.addOperator("index", 1)
		case *ast.SliceExpr:
			metrics.addOperator("slice", 1)
		case *ast.SelectorExpr:
			metrics.addOperator(".", 1)
		case *ast.CompositeLit:
			metrics.addOperator("composite", 1)
		case *ast.TypeAssertExpr:
			metrics.addOperator("type-assert", 1)
		case *ast.FuncDecl:
			metrics.addKeywordOperator(token.FUNC)
		case *ast.FuncLit:
			metrics.addKeywordOperator(token.FUNC)
		case *ast.GenDecl:
			metrics.addKeywordOperator(node.Tok)
		case *ast.CaseClause:
			if node.List == nil {
				metrics.addOperator("default", 1)
			} else {
				metrics.addOperator("case", 1)
			}
		case *ast.CommClause:
			if node.Comm == nil {
				metrics.addOperator("default", 1)
			} else {
				metrics.addOperator("case", 1)
			}
		case *ast.Ident:
			if node.Name == "_" {
				return true
			}
			if object := resolvedObject(info, node); object != nil {
				if name, ok := operandName(object); ok {
					metrics.addOperand(name, 1)
				}
			}
		case *ast.BasicLit:
			metrics.addOperand(node.Value, 1)
		}
		return true
	})

	return metrics, nil
}

func (m *Metrics) addOperator(name string, amount int) {
	if amount <= 0 {
		return
	}
	m.Operators[name] += amount
	m.TotalOperators += amount
}

func (m *Metrics) addOperand(name string, amount int) {
	if amount <= 0 {
		return
	}
	m.Operands[name] += amount
	m.TotalOperands += amount
}

func (m *Metrics) addKeywordOperator(tok token.Token) {
	if name, ok := keywordOperatorName(tok); ok {
		m.addOperator(name, 1)
	}
}

func resolvedObject(info *types.Info, ident *ast.Ident) types.Object {
	if object := info.Defs[ident]; object != nil {
		return object
	}
	return info.Uses[ident]
}

func operandName(object types.Object) (string, bool) {
	switch obj := object.(type) {
	case *types.Builtin:
		return "builtin:" + obj.Name(), true
	case *types.Const:
		return "const:" + obj.Name(), true
	case *types.Func:
		return "func:" + obj.Name(), true
	case *types.Label:
		return "label:" + obj.Name(), true
	case *types.Nil:
		return "nil", true
	case *types.PkgName:
		return "pkg:" + obj.Imported().Name(), true
	case *types.TypeName:
		return "type:" + obj.Name(), true
	case *types.Var:
		if obj.IsField() {
			return "field:" + obj.Name(), true
		}
		return "var:" + obj.Name(), true
	default:
		return "", false
	}
}

func keywordOperatorName(tok token.Token) (string, bool) {
	switch tok {
	case token.FUNC:
		return "func", true
	case token.TYPE:
		return "type", true
	case token.VAR:
		return "var", true
	case token.CONST:
		return "const", true
	case token.PACKAGE, token.IMPORT:
		return "", false
	default:
		return "", false
	}
}

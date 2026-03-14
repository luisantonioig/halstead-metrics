package halstead

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

// AnalyzeAST parses Go source with the standard library parser and derives metrics
// from the resulting syntax tree.
func AnalyzeAST(src []byte) (Metrics, error) {
	report, err := AnalyzeASTReport(src)
	if err != nil {
		return Metrics{}, err
	}
	return metricsFromSummary(report.File), nil
}

// AnalyzeASTReport parses Go source and returns file-level and per-function metrics.
func AnalyzeASTReport(src []byte) (AnalysisReport, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "", src, parser.ParseComments)
	if err != nil {
		return AnalysisReport{}, err
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
		return AnalysisReport{}, err
	}

	analyzer := astAnalyzer{
		fileSet: fileSet,
		info:    info,
	}
	fileMetrics := analyzer.analyzeNode(file)

	report := AnalysisReport{
		Analyzer:  fileMetrics.Name,
		File:      fileMetrics.Summary(),
		Functions: analyzer.functionReports(file),
	}

	return report, nil
}

type astAnalyzer struct {
	fileSet *token.FileSet
	info    *types.Info
}

func (a astAnalyzer) analyzeNode(root ast.Node) Metrics {
	metrics := Metrics{
		Name:      "go-ast-types",
		Operators: map[string]int{},
		Operands:  map[string]int{},
	}

	ast.Inspect(root, func(n ast.Node) bool {
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
			if object := resolvedObject(a.info, node); object != nil {
				if name, ok := operandName(object); ok {
					metrics.addOperand(name, 1)
				}
			}
		case *ast.BasicLit:
			metrics.addOperand(node.Value, 1)
		}
		return true
	})

	return metrics
}

func (a astAnalyzer) functionReports(file *ast.File) []FunctionReport {
	var reports []FunctionReport

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			reports = append(reports, a.newFunctionReport(functionDeclName(node), "func_decl", node))
			return true
		case *ast.FuncLit:
			reports = append(reports, a.newFunctionReport(functionLitName(a.fileSet, node), "func_lit", node))
			return false
		default:
			return true
		}
	})

	return reports
}

func (a astAnalyzer) newFunctionReport(name, kind string, node ast.Node) FunctionReport {
	start := a.fileSet.Position(node.Pos())
	end := a.fileSet.Position(node.End())
	metrics := a.analyzeNode(node)
	return FunctionReport{
		Name:    name,
		Kind:    kind,
		Start:   Position{Line: start.Line, Column: start.Column},
		End:     Position{Line: end.Line, Column: end.Column},
		Metrics: metrics.Summary(),
	}
}

func functionDeclName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return receiverName(fn.Recv.List[0].Type) + "." + fn.Name.Name
}

func functionLitName(fileSet *token.FileSet, fn *ast.FuncLit) string {
	pos := fileSet.Position(fn.Type.Func)
	return "func_literal@" + strconv.Itoa(pos.Line) + ":" + strconv.Itoa(pos.Column)
}

func receiverName(expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.Ident:
		return node.Name
	case *ast.StarExpr:
		return "*" + receiverName(node.X)
	case *ast.IndexExpr:
		return receiverName(node.X) + "[" + exprString(node.Index) + "]"
	case *ast.IndexListExpr:
		parts := make([]string, 0, len(node.Indices))
		for _, index := range node.Indices {
			parts = append(parts, exprString(index))
		}
		return receiverName(node.X) + "[" + strings.Join(parts, ",") + "]"
	default:
		return exprString(expr)
	}
}

func exprString(expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.Ident:
		return node.Name
	case *ast.SelectorExpr:
		return exprString(node.X) + "." + node.Sel.Name
	case *ast.StarExpr:
		return "*" + exprString(node.X)
	default:
		return "expr"
	}
}

func metricsFromSummary(summary MetricsSummary) Metrics {
	return Metrics{
		Name:           summary.Name,
		Operators:      cloneCounts(summary.Operators),
		Operands:       cloneCounts(summary.Operands),
		TotalOperators: summary.TotalOperators,
		TotalOperands:  summary.TotalOperands,
	}
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

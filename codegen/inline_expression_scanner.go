package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type InlineExpressionScanner struct {
	classifier             HoistableCallClassifier
	variableInitFilter     *VariableInitCallFilter
	conditionalArgAnalyzer *ConditionalArgumentAnalyzer
	gen                    *generator
}

func NewInlineExpressionScanner(g *generator) *InlineExpressionScanner {
	return &InlineExpressionScanner{
		classifier: NewHoistableCallClassifier(g),
		variableInitFilter: NewVariableInitCallFilter(
			g.taRegistry,
			g.inlineRegistry,
			g.runtimeOnlyFilter,
			g.exprAnalyzer,
		),
		conditionalArgAnalyzer: g.conditionalArgAnalyzer,
		gen:                    g,
	}
}

func (s *InlineExpressionScanner) ScanProgram(program *ast.Program) []CallInfo {
	var hoistable []CallInfo
	callRegistry := make(map[*ast.CallExpression]bool)

	for _, stmt := range program.Body {
		s.scanStatement(stmt, callRegistry, &hoistable)
	}

	return hoistable
}

func (s *InlineExpressionScanner) scanStatement(stmt ast.Node, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	switch node := stmt.(type) {
	case *ast.VariableDeclaration:
		s.scanVariableDeclaration(node, registry, hoistable)

	case *ast.ExpressionStatement:
		s.registerConditionalsInExpression(node.Expression)
		s.scanExpression(node.Expression, registry, hoistable)

	case *ast.IfStatement:
		if node.Test != nil {
			s.registerConditionalsInExpression(node.Test)
		}
		s.scanExpression(node.Test, registry, hoistable)
		for _, conseq := range node.Consequent {
			s.scanStatement(conseq, registry, hoistable)
		}
		for _, alt := range node.Alternate {
			s.scanStatement(alt, registry, hoistable)
		}

	case *ast.ForStatement:
		for _, bodyStmt := range node.Body {
			s.scanStatement(bodyStmt, registry, hoistable)
		}

	case *ast.ForInStatement:
		for _, bodyStmt := range node.Body {
			s.scanStatement(bodyStmt, registry, hoistable)
		}

	case *ast.WhileStatement:
		s.scanExpression(node.Condition, registry, hoistable)
		for _, bodyStmt := range node.Body {
			s.scanStatement(bodyStmt, registry, hoistable)
		}
	}
}

func (s *InlineExpressionScanner) scanVariableDeclaration(decl *ast.VariableDeclaration, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	for _, declarator := range decl.Declarations {
		if declarator.Init == nil {
			continue
		}

		s.scanForSumConditionals(declarator.Init, registry, hoistable)

		nestedCalls := s.gen.exprAnalyzer.FindNestedCalls(declarator.Init)
		filtered := s.variableInitFilter.FilterHoistable(nestedCalls, declarator.Init)
		for _, callInfo := range filtered {
			if !registry[callInfo.Call] {
				*hoistable = append(*hoistable, callInfo)
				registry[callInfo.Call] = true
			}
		}
	}
}

func (s *InlineExpressionScanner) registerConditionalsInExpression(expr ast.Expression) {
	if s.conditionalArgAnalyzer == nil {
		return
	}
	conditionals := s.conditionalArgAnalyzer.FindInExpression(expr)
	for _, condInfo := range conditionals {
		s.gen.tempVarMgr.RegisterConditional(condInfo.ContentHash, condInfo.Conditional)
	}
}

func (s *InlineExpressionScanner) scanExpression(expr ast.Expression, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		/* Bottom-up: recurse into children FIRST so dependencies are hoisted before parent.
		 * e.g. fixnan(ta.sma(close, 14)) → ta.sma hoisted before fixnan */
		for _, arg := range e.Arguments {
			s.scanExpression(arg, registry, hoistable)
		}
		s.processCall(e, registry, hoistable)

	case *ast.BinaryExpression:
		s.scanExpression(e.Left, registry, hoistable)
		s.scanExpression(e.Right, registry, hoistable)

	case *ast.LogicalExpression:
		s.scanExpression(e.Left, registry, hoistable)
		s.scanExpression(e.Right, registry, hoistable)

	case *ast.ConditionalExpression:
		s.scanExpression(e.Test, registry, hoistable)
		s.scanExpression(e.Consequent, registry, hoistable)
		s.scanExpression(e.Alternate, registry, hoistable)

	case *ast.UnaryExpression:
		s.scanExpression(e.Argument, registry, hoistable)

	case *ast.MemberExpression:
		s.scanExpression(e.Object, registry, hoistable)
		if propExpr, ok := e.Property.(ast.Expression); ok {
			s.scanExpression(propExpr, registry, hoistable)
		}

	case *ast.ObjectExpression:
		for _, prop := range e.Properties {
			s.scanExpression(prop.Value, registry, hoistable)
		}

	case *ast.Literal:
		if elemSlice, ok := e.Value.([]ast.Expression); ok {
			for _, elem := range elemSlice {
				s.scanExpression(elem, registry, hoistable)
			}
		}
	}
}

func (s *InlineExpressionScanner) processCall(call *ast.CallExpression, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	if registry[call] {
		return
	}

	/* Pre-register sum() conditional ternary vars (uses same key as SumHandler) */
	s.processSumConditional(call, registry, hoistable)

	if s.classifier.IsHoistable(call) {
		argHash := s.gen.exprAnalyzer.ComputeArgHash(call)
		funcName := s.gen.extractFunctionName(call.Callee)

		callInfo := CallInfo{
			Call:     call,
			FuncName: funcName,
			ArgHash:  argHash,
		}

		*hoistable = append(*hoistable, callInfo)
		registry[call] = true
	}
}

/* processSumConditional pre-registers ternary Series for sum() with conditional args.
 * Uses same CallInfo key as SumHandler.GenerateCode for deduplication via GetOrCreate. */
func (s *InlineExpressionScanner) processSumConditional(call *ast.CallExpression, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	funcName := s.gen.extractFunctionName(call.Callee)
	if funcName != "sum" && funcName != "math.sum" {
		return
	}

	if len(call.Arguments) < 1 {
		return
	}

	condExpr, ok := call.Arguments[0].(*ast.ConditionalExpression)
	if !ok {
		return
	}

	/* Must match SumHandler.GenerateCode: content-based hash for stability */
	hasher := &ExpressionHasher{}
	argHash := hasher.Hash(condExpr)

	callInfo := CallInfo{
		FuncName: "ternary",
		Call:     call,
		ArgHash:  argHash,
	}

	*hoistable = append(*hoistable, callInfo)
	registry[call] = true
}

/* scanForSumConditionals scans expressions for sum() calls with conditional args.
 * Used in VariableDeclaration context where regular TA hoisting is skipped. */
func (s *InlineExpressionScanner) scanForSumConditionals(expr ast.Expression, registry map[*ast.CallExpression]bool, hoistable *[]CallInfo) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		s.processSumConditional(e, registry, hoistable)
		for _, arg := range e.Arguments {
			s.scanForSumConditionals(arg, registry, hoistable)
		}

	case *ast.BinaryExpression:
		s.scanForSumConditionals(e.Left, registry, hoistable)
		s.scanForSumConditionals(e.Right, registry, hoistable)

	case *ast.LogicalExpression:
		s.scanForSumConditionals(e.Left, registry, hoistable)
		s.scanForSumConditionals(e.Right, registry, hoistable)

	case *ast.ConditionalExpression:
		s.scanForSumConditionals(e.Test, registry, hoistable)
		s.scanForSumConditionals(e.Consequent, registry, hoistable)
		s.scanForSumConditionals(e.Alternate, registry, hoistable)

	case *ast.UnaryExpression:
		s.scanForSumConditionals(e.Argument, registry, hoistable)

	case *ast.MemberExpression:
		s.scanForSumConditionals(e.Object, registry, hoistable)
	}
}

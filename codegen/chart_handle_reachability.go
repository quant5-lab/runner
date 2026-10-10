package codegen

import "github.com/quant5-lab/runner/ast"

type chartHandleReachability map[*ast.CallExpression]bool

func newChartHandleReachability(program *ast.Program) chartHandleReachability {
	if program == nil {
		return nil
	}
	analyzer := chartHandleReachabilityAnalyzer{
		producersByVariable: collectChartHandleProducers(program.Body),
		reachable:           chartHandleReachability{},
	}
	analyzer.markHandlesUsedByGetters(program.Body)
	analyzer.markHandlesUsedByBacktestBoundaries(program.Body)
	return analyzer.reachable
}

func (r chartHandleReachability) IsCalculationBearing(call *ast.CallExpression) bool {
	return r[call]
}

type chartHandleReachabilityAnalyzer struct {
	producersByVariable map[string][]*ast.CallExpression
	reachable           chartHandleReachability
}

// The fixpoint only ever adds producers, so a variable rebound to a second
// producer — by reassignment or by a declaration in each branch — widens its
// producer set instead of oscillating.
func collectChartHandleProducers(nodes []ast.Node) map[string][]*ast.CallExpression {
	producersByVariable := map[string][]*ast.CallExpression{}
	for changed := true; changed; {
		changed = false
		walkControlFlow(nodes, func(node ast.Node) {
			decl, ok := node.(*ast.VariableDeclaration)
			if !ok {
				return
			}
			for _, d := range decl.Declarations {
				bound := chartHandleProducersInExpression(d.Init, producersByVariable)
				if len(bound) == 0 {
					continue
				}
				for _, name := range extractPatternNames(d.ID) {
					if addChartHandleProducers(producersByVariable, name, bound) {
						changed = true
					}
				}
			}
		})
	}
	return producersByVariable
}

func addChartHandleProducers(producersByVariable map[string][]*ast.CallExpression, name string, producers []*ast.CallExpression) bool {
	known := producersByVariable[name]
	changed := false
	for _, producer := range producers {
		if containsChartHandleProducer(known, producer) {
			continue
		}
		known = append(known, producer)
		changed = true
	}
	if changed {
		producersByVariable[name] = known
	}
	return changed
}

func containsChartHandleProducer(producers []*ast.CallExpression, target *ast.CallExpression) bool {
	for _, producer := range producers {
		if producer == target {
			return true
		}
	}
	return false
}

func chartHandleProducersInExpression(expr ast.Expression, producersByVariable map[string][]*ast.CallExpression) []*ast.CallExpression {
	switch e := expr.(type) {
	case *ast.CallExpression:
		if isChartHandleProducer(extractCallFunctionName(e)) {
			return []*ast.CallExpression{e}
		}
	case *ast.Identifier:
		return producersByVariable[e.Name]
	}
	return nil
}

func (a chartHandleReachabilityAnalyzer) markHandlesUsedByGetters(nodes []ast.Node) {
	walkCallExpressions(nodes, func(call *ast.CallExpression) {
		if !isCalculationBearingChartCall(extractCallFunctionName(call)) {
			return
		}
		for _, arg := range call.Arguments {
			a.markHandlesInExpression(arg)
		}
	})
}

func (a chartHandleReachabilityAnalyzer) markHandlesUsedByBacktestBoundaries(nodes []ast.Node) {
	for _, node := range nodes {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			call, ok := s.Expression.(*ast.CallExpression)
			if !ok || !isBacktestCriticalSink(extractCallFunctionName(call)) {
				continue
			}
			for _, arg := range backtestCriticalArguments(extractCallFunctionName(call), call.Arguments) {
				a.markHandlesInExpression(arg)
			}
		case *ast.VariableDeclaration:
			for _, d := range s.Declarations {
				if arrow, ok := d.Init.(*ast.ArrowFunctionExpression); ok {
					a.markHandlesUsedByBacktestBoundaries(arrow.Body)
				}
			}
		case *ast.IfStatement:
			if containsBacktestCriticalSink(s.Consequent) || containsBacktestCriticalSink(s.Alternate) {
				a.markHandlesInExpression(s.Test)
			}
			a.markHandlesUsedByBacktestBoundaries(s.Consequent)
			a.markHandlesUsedByBacktestBoundaries(s.Alternate)
		case *ast.ForStatement:
			if containsBacktestCriticalSink(s.Body) {
				a.markHandlesInExpression(s.From)
				a.markHandlesInExpression(s.To)
				a.markHandlesInExpression(s.Step)
			}
			a.markHandlesUsedByBacktestBoundaries(s.Body)
		case *ast.ForInStatement:
			if containsBacktestCriticalSink(s.Body) {
				a.markHandlesInExpression(s.Collection)
			}
			a.markHandlesUsedByBacktestBoundaries(s.Body)
		case *ast.WhileStatement:
			if containsBacktestCriticalSink(s.Body) {
				a.markHandlesInExpression(s.Condition)
			}
			a.markHandlesUsedByBacktestBoundaries(s.Body)
		}
	}
}

func (a chartHandleReachabilityAnalyzer) markHandlesInExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		for _, producer := range a.producersByVariable[e.Name] {
			a.reachable[producer] = true
		}
	case *ast.CallExpression:
		if isChartHandleProducer(extractCallFunctionName(e)) {
			a.reachable[e] = true
		}
		for _, arg := range e.Arguments {
			a.markHandlesInExpression(arg)
		}
	case *ast.BinaryExpression:
		a.markHandlesInExpression(e.Left)
		a.markHandlesInExpression(e.Right)
	case *ast.LogicalExpression:
		a.markHandlesInExpression(e.Left)
		a.markHandlesInExpression(e.Right)
	case *ast.ConditionalExpression:
		a.markHandlesInExpression(e.Test)
		a.markHandlesInExpression(e.Consequent)
		a.markHandlesInExpression(e.Alternate)
	case *ast.UnaryExpression:
		a.markHandlesInExpression(e.Argument)
	case *ast.MemberExpression:
		a.markHandlesInExpression(e.Object)
		if e.Computed {
			a.markHandlesInExpression(e.Property)
		}
	case *ast.ObjectExpression:
		for _, prop := range e.Properties {
			a.markHandlesInExpression(prop.Value)
		}
	}
}

func containsBacktestCriticalSink(nodes []ast.Node) bool {
	found := false
	walkCallExpressions(nodes, func(call *ast.CallExpression) {
		if isBacktestCriticalSink(extractCallFunctionName(call)) {
			found = true
		}
	})
	return found
}

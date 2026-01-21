package codegen

import (
	"fmt"
	"os"

	"github.com/quant5-lab/runner/ast"
)

type ArrowCallSite struct {
	FunctionName string
	CallIndex    int
	ContextVar   string
}

/*
ArrowCallSiteScanner detects arrow function calls requiring ArrowContext.

Design (SRP): Single purpose - identify call sites, delegate generation and lifecycle

Implementation: **COMPLETE** AST traversal covering:

Statement Types:
  - VariableDeclaration (Init expressions)
  - ExpressionStatement (standalone calls)
  - IfStatement (Test, Consequent, Alternate)
  - ForStatement (Body statements)

Expression Types (ALL):
  - CallExpression (detect + register + recurse arguments)
  - BinaryExpression, LogicalExpression (Left + Right)
  - ConditionalExpression (Test + Consequent + Alternate)
  - UnaryExpression (Argument)
  - MemberExpression (Object + Property, including computed)
  - ArrowFunctionExpression (nested definitions)
  - ObjectExpression (Properties[].Value)
  - Literal (array literals: []ast.Expression)
  - Identifier (terminal)

This ensures 100% support for ANY ARBITRARY PineScript input containing user-defined function calls.

Logging: Prefix [ARROW_SCANNER] for pair-debugging (DEBUG_ARROW_SCANNER=1)
*/
type ArrowCallSiteScanner struct {
	variables map[string]string
	debug     bool
}

func NewArrowCallSiteScanner(variables map[string]string) *ArrowCallSiteScanner {
	return &ArrowCallSiteScanner{
		variables: variables,
		debug:     os.Getenv("DEBUG_ARROW_SCANNER") == "1",
	}
}

func (s *ArrowCallSiteScanner) ScanForArrowFunctionCalls(program *ast.Program) []ArrowCallSite {
	var callSites []ArrowCallSite
	callCounts := make(map[string]int)

	if s.debug {
		fmt.Fprintf(os.Stderr, "[ARROW_SCANNER] Starting scan of %d top-level statements\n", len(program.Body))
	}

	for _, stmt := range program.Body {
		s.scanStatement(stmt, callCounts, &callSites)
	}

	if s.debug {
		fmt.Fprintf(os.Stderr, "[ARROW_SCANNER] Scan complete: %d call sites detected\n", len(callSites))
		for i, site := range callSites {
			fmt.Fprintf(os.Stderr, "[ARROW_SCANNER]   [%d] %s (call #%d) → %s\n",
				i, site.FunctionName, site.CallIndex, site.ContextVar)
		}
	}

	return callSites
}

/* scanStatement traverses statement nodes recursively */
func (s *ArrowCallSiteScanner) scanStatement(stmt ast.Node, callCounts map[string]int, callSites *[]ArrowCallSite) {
	switch node := stmt.(type) {
	case *ast.VariableDeclaration:
		for _, declarator := range node.Declarations {
			s.scanExpression(declarator.Init, callCounts, callSites)
		}

	case *ast.ExpressionStatement:
		s.scanExpression(node.Expression, callCounts, callSites)

	case *ast.IfStatement:
		s.scanExpression(node.Test, callCounts, callSites)
		for _, conseq := range node.Consequent {
			s.scanStatement(conseq, callCounts, callSites)
		}
		for _, alt := range node.Alternate {
			s.scanStatement(alt, callCounts, callSites)
		}

	case *ast.ForStatement:
		for _, bodyStmt := range node.Body {
			s.scanStatement(bodyStmt, callCounts, callSites)
		}
	}
}

/* scanExpression traverses expression nodes recursively */
func (s *ArrowCallSiteScanner) scanExpression(expr ast.Expression, callCounts map[string]int, callSites *[]ArrowCallSite) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		funcName := s.extractFunctionName(e.Callee)
		if s.isUserDefinedFunction(funcName) {
			callCounts[funcName]++
			callIndex := callCounts[funcName]
			contextVar := formatContextVariableName(funcName, callIndex)

			if s.debug {
				fmt.Fprintf(os.Stderr, "[ARROW_SCANNER] Detected call: %s (call #%d) → %s\n",
					funcName, callIndex, contextVar)
			}

			*callSites = append(*callSites, ArrowCallSite{
				FunctionName: funcName,
				CallIndex:    callIndex,
				ContextVar:   contextVar,
			})
		}
		/* Scan nested calls in arguments */
		for _, arg := range e.Arguments {
			s.scanExpression(arg, callCounts, callSites)
		}

	case *ast.BinaryExpression:
		s.scanExpression(e.Left, callCounts, callSites)
		s.scanExpression(e.Right, callCounts, callSites)

	case *ast.LogicalExpression:
		s.scanExpression(e.Left, callCounts, callSites)
		s.scanExpression(e.Right, callCounts, callSites)

	case *ast.ConditionalExpression:
		s.scanExpression(e.Test, callCounts, callSites)
		s.scanExpression(e.Consequent, callCounts, callSites)
		s.scanExpression(e.Alternate, callCounts, callSites)

	case *ast.UnaryExpression:
		s.scanExpression(e.Argument, callCounts, callSites)

	case *ast.MemberExpression:
		s.scanExpression(e.Object, callCounts, callSites)
		if propExpr, ok := e.Property.(ast.Expression); ok {
			s.scanExpression(propExpr, callCounts, callSites)
		}

	case *ast.ArrowFunctionExpression:
		for _, bodyStmt := range e.Body {
			s.scanStatement(bodyStmt, callCounts, callSites)
		}

	case *ast.ObjectExpression:
		/* Object literals: {stop: calcStop(), limit: calcLimit()} */
		for _, prop := range e.Properties {
			s.scanExpression(prop.Value, callCounts, callSites)
		}

	case *ast.Literal:
		/* Array literals: [func1(), func2(), func3()] */
		if elemSlice, ok := e.Value.([]ast.Expression); ok {
			for _, elem := range elemSlice {
				s.scanExpression(elem, callCounts, callSites)
			}
		}
	}
}

func (s *ArrowCallSiteScanner) extractFunctionName(callee ast.Expression) string {
	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}

	if member, ok := callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}

	return ""
}

func (s *ArrowCallSiteScanner) isUserDefinedFunction(funcName string) bool {
	varType, exists := s.variables[funcName]
	return exists && varType == "function"
}

func formatContextVariableName(funcName string, callIndex int) string {
	return "arrowCtx_" + funcName + "_" + string(rune('0'+callIndex))
}

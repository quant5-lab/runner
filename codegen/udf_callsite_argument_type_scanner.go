package codegen

import "github.com/quant5-lab/runner/ast"

// UDFParameterTypes maps each positional parameter index to the ParameterUsageType
// inferred from the argument values at the call site.
type UDFParameterTypes = []ParameterUsageType

// UDFCallSiteArgumentTypeScanner closes the "parameter type unknowable from body
// alone" gap: when a UDF parameter is only forwarded to an opaque built-in inside
// the body, the concrete argument values at each call site are the only evidence
// of the parameter type.  The first call site encountered wins; Pine semantics
// guarantee uniform argument types across all call sites for a given position.
type UDFCallSiteArgumentTypeScanner struct {
	constants map[string]interface{}
	variables map[string]string
}

func NewUDFCallSiteArgumentTypeScanner(
	constants map[string]interface{},
	variables map[string]string,
) *UDFCallSiteArgumentTypeScanner {
	return &UDFCallSiteArgumentTypeScanner{
		constants: constants,
		variables: variables,
	}
}

func (s *UDFCallSiteArgumentTypeScanner) ScanProgram(program *ast.Program) map[string]UDFParameterTypes {
	result := make(map[string]UDFParameterTypes)
	for _, stmt := range program.Body {
		s.scanStatement(stmt, result)
	}
	return result
}

func (s *UDFCallSiteArgumentTypeScanner) scanStatement(node ast.Node, out map[string]UDFParameterTypes) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, d := range n.Declarations {
			s.scanExpression(d.Init, out)
		}
	case *ast.ExpressionStatement:
		s.scanExpression(n.Expression, out)
	case *ast.IfStatement:
		s.scanExpression(n.Test, out)
		for _, c := range n.Consequent {
			s.scanStatement(c, out)
		}
		for _, a := range n.Alternate {
			s.scanStatement(a, out)
		}
	case *ast.ForStatement:
		for _, b := range n.Body {
			s.scanStatement(b, out)
		}
	case *ast.ForInStatement:
		for _, b := range n.Body {
			s.scanStatement(b, out)
		}
	case *ast.WhileStatement:
		s.scanExpression(n.Condition, out)
		for _, b := range n.Body {
			s.scanStatement(b, out)
		}
	}
}

func (s *UDFCallSiteArgumentTypeScanner) scanExpression(expr ast.Expression, out map[string]UDFParameterTypes) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.CallExpression:
		funcName := s.calleeName(e.Callee)
		if s.isUDF(funcName) {
			if _, seen := out[funcName]; !seen {
				out[funcName] = s.inferArgumentTypes(e.Arguments)
			}
		}
		for _, arg := range e.Arguments {
			s.scanExpression(arg, out)
		}
	case *ast.BinaryExpression:
		s.scanExpression(e.Left, out)
		s.scanExpression(e.Right, out)
	case *ast.LogicalExpression:
		s.scanExpression(e.Left, out)
		s.scanExpression(e.Right, out)
	case *ast.ConditionalExpression:
		s.scanExpression(e.Test, out)
		s.scanExpression(e.Consequent, out)
		s.scanExpression(e.Alternate, out)
	case *ast.UnaryExpression:
		s.scanExpression(e.Argument, out)
	case *ast.MemberExpression:
		s.scanExpression(e.Object, out)
		if prop, ok := e.Property.(ast.Expression); ok {
			s.scanExpression(prop, out)
		}
	case *ast.IfStatement:
		s.scanExpression(e.Test, out)
		for _, c := range e.Consequent {
			s.scanStatement(c, out)
		}
		for _, a := range e.Alternate {
			s.scanStatement(a, out)
		}
	}
}

func (s *UDFCallSiteArgumentTypeScanner) inferArgumentTypes(args []ast.Expression) UDFParameterTypes {
	types := make(UDFParameterTypes, len(args))
	for i, arg := range args {
		types[i] = s.inferExpressionType(arg)
	}
	return types
}

// Resolution order for identifier arguments:
//  1. Variable type wins when it is explicitly "float" — this covers input.source
//     variables whose constants map entry is the function tag "input.source" (a
//     string) rather than an actual string value.  Without this guard, the scanner
//     would misclassify source parameters as ParameterUsageString.
//  2. Constants map: a string-valued entry means the variable holds a Pine string
//     constant (e.g. Session = input("0900-1700") stores "0900-1700").
//  3. Variable type "string" or "color" as a fallback for variables typed during
//     the main codegen pass (rare at scan time, but handled for completeness).
func (s *UDFCallSiteArgumentTypeScanner) inferExpressionType(expr ast.Expression) ParameterUsageType {
	switch e := expr.(type) {
	case *ast.Literal:
		if _, ok := e.Value.(string); ok {
			return ParameterUsageString
		}
	case *ast.Identifier:
		if varType, ok := s.variables[e.Name]; ok && varType == "float" {
			return ParameterUsageScalar
		}
		if constVal, ok := s.constants[e.Name]; ok {
			if _, isStr := constVal.(string); isStr {
				return ParameterUsageString
			}
		}
		if varType, ok := s.variables[e.Name]; ok {
			if varType == "string" || varType == "color" {
				return ParameterUsageString
			}
		}
	case *ast.BinaryExpression:
		if e.Operator == "+" {
			if s.inferExpressionType(e.Left) == ParameterUsageString ||
				s.inferExpressionType(e.Right) == ParameterUsageString {
				return ParameterUsageString
			}
		}
	case *ast.ConditionalExpression:
		if s.inferExpressionType(e.Consequent) == ParameterUsageString ||
			s.inferExpressionType(e.Alternate) == ParameterUsageString {
			return ParameterUsageString
		}
	}
	return ParameterUsageScalar
}

func (s *UDFCallSiteArgumentTypeScanner) isUDF(name string) bool {
	return s.variables[name] == "function"
}

func (s *UDFCallSiteArgumentTypeScanner) calleeName(callee ast.Expression) string {
	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}

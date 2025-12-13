package request

import (
	"github.com/quant5-lab/runner/ast"
)

/* PivotFunctionType represents the type of pivot function detected */
type PivotFunctionType int

const (
	PivotTypeNone PivotFunctionType = iota
	PivotTypeHigh
	PivotTypeLow
)

/* PivotCallInfo contains extracted parameters from a pivot function call */
type PivotCallInfo struct {
	Type      PivotFunctionType
	Source    ast.Expression // Source series (high/low or custom)
	LeftBars  int
	RightBars int
	HasOffset bool // Whether [offset] subscript is applied
	Offset    int  // The subscript value if present
}

/* PivotDetector identifies pivot function calls in expressions (SRP) */
type PivotDetector struct{}

/* NewPivotDetector creates a new pivot function detector */
func NewPivotDetector() *PivotDetector {
	return &PivotDetector{}
}

/* DetectPivotCall examines an expression to identify pivot function usage */
func (d *PivotDetector) DetectPivotCall(expr ast.Expression) (*PivotCallInfo, bool) {
	if callExpr, ok := expr.(*ast.CallExpression); ok {
		return d.extractFromCall(callExpr)
	}

	if memberExpr, ok := expr.(*ast.MemberExpression); ok && memberExpr.Computed {
		if callExpr, ok := memberExpr.Object.(*ast.CallExpression); ok {
			info, detected := d.extractFromCall(callExpr)
			if detected && memberExpr.Property != nil {
				if offsetLit, ok := memberExpr.Property.(*ast.Literal); ok {
					info.HasOffset = true
					info.Offset = d.extractIntFromLiteral(offsetLit)
				}
			}
			return info, detected
		}
	}

	return nil, false
}

/* extractFromCall extracts pivot parameters from CallExpression */
func (d *PivotDetector) extractFromCall(call *ast.CallExpression) (*PivotCallInfo, bool) {
	funcName := d.extractFunctionName(call)

	pivotType := d.identifyPivotType(funcName)
	if pivotType == PivotTypeNone {
		return nil, false
	}

	info := &PivotCallInfo{
		Type: pivotType,
	}

	argCount := len(call.Arguments)

	if argCount == 2 {
		info.LeftBars = d.extractIntValue(call.Arguments[0])
		info.RightBars = d.extractIntValue(call.Arguments[1])
	} else if argCount == 3 {
		info.Source = call.Arguments[0]
		info.LeftBars = d.extractIntValue(call.Arguments[1])
		info.RightBars = d.extractIntValue(call.Arguments[2])
	} else {
		return nil, false
	}

	return info, true
}

/* extractFunctionName retrieves the full function name from callee */
func (d *PivotDetector) extractFunctionName(call *ast.CallExpression) string {
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		return ident.Name
	}
	if memberExpr, ok := call.Callee.(*ast.MemberExpression); ok {
		if obj, ok := memberExpr.Object.(*ast.Identifier); ok {
			if prop, ok := memberExpr.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}
	return ""
}

/* identifyPivotType maps function name to pivot type */
func (d *PivotDetector) identifyPivotType(funcName string) PivotFunctionType {
	switch funcName {
	case "pivothigh", "ta.pivothigh":
		return PivotTypeHigh
	case "pivotlow", "ta.pivotlow":
		return PivotTypeLow
	default:
		return PivotTypeNone
	}
}

/* extractIntValue safely extracts integer from expression */
func (d *PivotDetector) extractIntValue(expr ast.Expression) int {
	if lit, ok := expr.(*ast.Literal); ok {
		return d.extractIntFromLiteral(lit)
	}
	return 0
}

/* extractIntFromLiteral converts Literal value to int */
func (d *PivotDetector) extractIntFromLiteral(lit *ast.Literal) int {
	switch v := lit.Value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

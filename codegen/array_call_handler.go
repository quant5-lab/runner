package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ArrayMethodCallHandler: array.get(arr, index) and array.size(arr) for array_series variables.
 * arr may be an Identifier (current bar) or a computed MemberExpression arr[n] (history offset n).
 * Dynamic element indices are supported for array.get. */
type ArrayMethodCallHandler struct{}

func (h *ArrayMethodCallHandler) CanHandle(funcName string) bool {
	return funcName == "array.get" || funcName == "array.size"
}

func (h *ArrayMethodCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "array.get":
		return h.generateGet(g, call)
	case "array.size":
		return h.generateSize(g, call)
	}
	return "", nil
}

func (h *ArrayMethodCallHandler) generateGet(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.get requires 2 arguments, got %d", len(call.Arguments))
	}

	seriesName, barOffset, isDynOffset, dynOffsetExpr, err := h.resolveArraySeriesAndOffset(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.get: %w", err)
	}

	indexCode, err := h.resolveIndexCode(g, call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.get: %w", err)
	}

	if isDynOffset {
		return fmt.Sprintf("%sArraySeries.Elem(int(%s), %s)", seriesName, dynOffsetExpr, indexCode), nil
	}
	return fmt.Sprintf("%sArraySeries.Elem(%d, %s)", seriesName, barOffset, indexCode), nil
}

func (h *ArrayMethodCallHandler) generateSize(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.size requires 1 argument, got %d", len(call.Arguments))
	}

	arrayExpr, err := h.resolveArrayExpression(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.size: %w", err)
	}

	return fmt.Sprintf("float64(len(%s))", arrayExpr), nil
}

// resolveArraySeriesAndOffset extracts the array series variable name and bar offset
// from an array argument. Returns (seriesName, barOffset, isDynOffset, dynExpr, err).
// Used by generateGet to build nil-safe Elem calls.
func (h *ArrayMethodCallHandler) resolveArraySeriesAndOffset(g *generator, arg ast.Expression) (seriesName string, barOffset int, isDynOffset bool, dynExpr string, err error) {
	switch e := arg.(type) {
	case *ast.Identifier:
		if !g.isArraySeriesVariable(e.Name) {
			err = fmt.Errorf("variable %q is not an array<float>", e.Name)
			return
		}
		seriesName = e.Name
		return
	case *ast.MemberExpression:
		if e.Computed {
			if id, ok := e.Object.(*ast.Identifier); ok && g.isArraySeriesVariable(id.Name) {
				barOffset, isDynOffset, dynExpr, err = resolveSubscriptOffset(e.Property)
				seriesName = id.Name
				return
			}
		}
	}
	err = fmt.Errorf("variable is not a registered array<float>")
	return
}

func (h *ArrayMethodCallHandler) resolveArrayExpression(g *generator, arg ast.Expression) (string, error) {
	switch e := arg.(type) {
	case *ast.Identifier:
		if !g.isArraySeriesVariable(e.Name) {
			return "", fmt.Errorf("variable %q is not an array<float>", e.Name)
		}
		return fmt.Sprintf("%sArraySeries.Get(0)", e.Name), nil

	case *ast.MemberExpression:
		if !e.Computed {
			break
		}
		id, ok := e.Object.(*ast.Identifier)
		if !ok || !g.isArraySeriesVariable(id.Name) {
			break
		}
		offset, isDynamic, dynExpr, err := resolveSubscriptOffset(e.Property)
		if err != nil {
			return "", err
		}
		if isDynamic {
			return fmt.Sprintf("%sArraySeries.Get(int(%s))", id.Name, dynExpr), nil
		}
		return fmt.Sprintf("%sArraySeries.Get(%d)", id.Name, offset), nil
	}

	exprCode, err := g.generateArrowFunctionExpression(arg)
	if err != nil {
		return "", fmt.Errorf("unsupported array expression: %w", err)
	}
	return exprCode, nil
}

func (h *ArrayMethodCallHandler) resolveIndexCode(g *generator, arg ast.Expression) (string, error) {
	switch e := arg.(type) {
	case *ast.Literal:
		switch v := e.Value.(type) {
		case int:
			return fmt.Sprintf("%d", v), nil
		case float64:
			return fmt.Sprintf("%d", int(v)), nil
		}
	}
	exprCode, err := g.generateArrowFunctionExpression(arg)
	if err != nil {
		return "", fmt.Errorf("index expression: %w", err)
	}
	return fmt.Sprintf("int(%s)", exprCode), nil
}

func resolveSubscriptOffset(prop ast.Expression) (literal int, isDynamic bool, dynExpr string, err error) {
	lit, ok := prop.(*ast.Literal)
	if !ok {
		return 0, true, fmt.Sprintf("%v", prop), nil
	}
	switch v := lit.Value.(type) {
	case int:
		return v, false, "", nil
	case float64:
		return int(v), false, "", nil
	}
	return 0, false, "", fmt.Errorf("unsupported subscript literal type %T", lit.Value)
}

func (g *generator) isArraySeriesVariable(varName string) bool {
	varType, exists := g.variables[varName]
	return exists && varType == "array_series"
}

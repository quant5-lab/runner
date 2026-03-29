package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ArrayMethodCallHandler: Dispatcher for all array.* operations.
 * Delegates to specialized handlers based on function category:
 *   - ArrayConstructorHandler: array.new_*, array.from
 *   - ArrayMutatorCodegen: array.push, array.pop, array.set, etc.
 *   - ArrayReaderCodegen: array.first, array.last, array.sum, etc.
 *   - Direct methods: array.get, array.size (legacy implementation)
 */
type ArrayMethodCallHandler struct {
	constructorHandler *ArrayConstructorHandler
	mutatorCodegen     *ArrayMutatorCodegen
	readerCodegen      *ArrayReaderCodegen
}

func NewArrayMethodCallHandler() *ArrayMethodCallHandler {
	return &ArrayMethodCallHandler{
		constructorHandler: NewArrayConstructorHandler(),
		mutatorCodegen:     NewArrayMutatorCodegen(),
		readerCodegen:      NewArrayReaderCodegen(),
	}
}

func (h *ArrayMethodCallHandler) CanHandle(funcName string) bool {
	return h.isArrayNamespaceFunction(funcName)
}

func (h *ArrayMethodCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	// Handle direct legacy methods first (array.get, array.size)
	switch funcName {
	case "array.get":
		return h.generateGet(g, call)
	case "array.size":
		return h.generateSize(g, call)
	}

	// Delegate to specialized handlers
	if h.constructorHandler.CanHandle(funcName) {
		return h.constructorHandler.GenerateCode(g, call)
	}

	if h.mutatorCodegen.CanHandle(funcName) {
		return h.mutatorCodegen.GenerateCode(g, call)
	}

	if h.readerCodegen.CanHandle(funcName) {
		return h.readerCodegen.GenerateCode(g, call)
	}

	return "", nil
}

func (h *ArrayMethodCallHandler) isArrayNamespaceFunction(funcName string) bool {
	if len(funcName) < 6 {
		return false
	}
	return funcName[:6] == "array."
}

func (h *ArrayMethodCallHandler) generateGet(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.get requires 2 arguments, got %d", len(call.Arguments))
	}

	seriesName, barOffset, isDynOffset, dynOffsetExpr, err := h.resolveArraySeriesAndOffset(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.get: %w", err)
	}

	elemType, ok := g.lookupArrayElementType(seriesName)
	if !ok {
		return "", fmt.Errorf("array.get: variable %q type not found", seriesName)
	}

	indexCode, err := h.resolveIndexCode(g, call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.get: %w", err)
	}

	if isDynOffset {
		return fmt.Sprintf("%s%s.Elem(int(%s), %s)", seriesName, elemType.VariableSuffix(), dynOffsetExpr, indexCode), nil
	}
	return fmt.Sprintf("%s%s.Elem(%d, %s)", seriesName, elemType.VariableSuffix(), barOffset, indexCode), nil
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
			err = fmt.Errorf("variable %q is not an array", e.Name)
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
	err = fmt.Errorf("variable is not a registered array")
	return
}

func (h *ArrayMethodCallHandler) resolveArrayExpression(g *generator, arg ast.Expression) (string, error) {
	switch e := arg.(type) {
	case *ast.Identifier:
		elemType, ok := g.lookupArrayElementType(e.Name)
		if !ok {
			return "", fmt.Errorf("variable %q is not an array", e.Name)
		}
		return fmt.Sprintf("%s%s.Get(0)", e.Name, elemType.VariableSuffix()), nil

	case *ast.MemberExpression:
		if !e.Computed {
			break
		}
		id, ok := e.Object.(*ast.Identifier)
		if !ok {
			break
		}
		elemType, ok := g.lookupArrayElementType(id.Name)
		if !ok {
			break
		}
		offset, isDynamic, dynExpr, err := resolveSubscriptOffset(e.Property)
		if err != nil {
			return "", err
		}
		if isDynamic {
			return fmt.Sprintf("%s%s.Get(int(%s))", id.Name, elemType.VariableSuffix(), dynExpr), nil
		}
		return fmt.Sprintf("%s%s.Get(%d)", id.Name, elemType.VariableSuffix(), offset), nil
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
	if g.arrayVariableRegistry == nil {
		return false
	}
	return g.arrayVariableRegistry.IsArrayVariable(varName)
}

package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ArrayMutatorCodegen struct{}

func NewArrayMutatorCodegen() *ArrayMutatorCodegen {
	return &ArrayMutatorCodegen{}
}

func (h *ArrayMutatorCodegen) CanHandle(funcName string) bool {
	classifier := NewArrayConstructorClassifier()
	return classifier.IsMutatingMethod(funcName)
}

func (h *ArrayMutatorCodegen) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "array.push":
		return h.generatePush(g, call)
	case "array.pop":
		return h.generatePop(g, call)
	case "array.shift":
		return h.generateShift(g, call)
	case "array.unshift":
		return h.generateUnshift(g, call)
	case "array.set":
		return h.generateSet(g, call)
	case "array.insert":
		return h.generateInsert(g, call)
	case "array.remove":
		return h.generateRemove(g, call)
	case "array.clear":
		return h.generateClear(g, call)
	case "array.fill":
		return h.generateFill(g, call)
	case "array.reverse":
		return h.generateReverse(g, call)
	case "array.sort":
		return h.generateSort(g, call)
	case "array.concat":
		return h.generateConcat(g, call)
	}

	return "", nil
}

func (h *ArrayMutatorCodegen) generatePush(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.push requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.push: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.push: value: %w", err)
	}

	return fmt.Sprintf("%s.Push(%s%s, %s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), valueCode), nil
}

func (h *ArrayMutatorCodegen) generatePop(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.pop requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.pop: %w", err)
	}

	return fmt.Sprintf("%s.Pop(%s%s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix()), nil
}

func (h *ArrayMutatorCodegen) generateShift(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.shift requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.shift: %w", err)
	}

	return fmt.Sprintf("%s.Shift(%s%s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix()), nil
}

func (h *ArrayMutatorCodegen) generateUnshift(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.unshift requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.unshift: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.unshift: value: %w", err)
	}

	return fmt.Sprintf("%s.Unshift(%s%s, %s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), valueCode), nil
}

func (h *ArrayMutatorCodegen) generateSet(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 3 {
		return "", fmt.Errorf("array.set requires 3 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.set: %w", err)
	}

	indexCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.set: index: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
	if err != nil {
		return "", fmt.Errorf("array.set: value: %w", err)
	}

	return fmt.Sprintf("%s.SetElement(%s%s, int(%s), %s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), indexCode, valueCode), nil
}

func (h *ArrayMutatorCodegen) generateInsert(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 3 {
		return "", fmt.Errorf("array.insert requires 3 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.insert: %w", err)
	}

	indexCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.insert: index: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
	if err != nil {
		return "", fmt.Errorf("array.insert: value: %w", err)
	}

	return fmt.Sprintf("%s.Insert(%s%s, int(%s), %s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), indexCode, valueCode), nil
}

func (h *ArrayMutatorCodegen) generateRemove(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.remove requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.remove: %w", err)
	}

	indexCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.remove: index: %w", err)
	}

	return fmt.Sprintf("%s.Remove(%s%s, int(%s))",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), indexCode), nil
}

func (h *ArrayMutatorCodegen) generateClear(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.clear requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.clear: %w", err)
	}

	return fmt.Sprintf("%s.Clear(%s%s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix()), nil
}

func (h *ArrayMutatorCodegen) generateFill(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 || len(call.Arguments) > 4 {
		return "", fmt.Errorf("array.fill requires 2-4 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.fill: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.fill: value: %w", err)
	}

	indexFrom := "0"
	indexTo := "-1"

	if len(call.Arguments) >= 3 {
		fromCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
		if err != nil {
			return "", fmt.Errorf("array.fill: indexFrom: %w", err)
		}
		indexFrom = fmt.Sprintf("int(%s)", fromCode)
	}

	if len(call.Arguments) == 4 {
		toCode, err := g.generateArrowFunctionExpression(call.Arguments[3])
		if err != nil {
			return "", fmt.Errorf("array.fill: indexTo: %w", err)
		}
		indexTo = fmt.Sprintf("int(%s)", toCode)
	}

	return fmt.Sprintf("%s.Fill(%s%s, %s, %s, %s)",
		elemType.MutatorConstructor(), arrayVar, elemType.VariableSuffix(), valueCode, indexFrom, indexTo), nil
}

func (h *ArrayMutatorCodegen) generateReverse(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.reverse requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.reverse: %w", err)
	}

	return fmt.Sprintf("%s.Reverse(%s%s)",
		elemType.TransformerConstructor(), arrayVar, elemType.VariableSuffix()), nil
}

func (h *ArrayMutatorCodegen) generateSort(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
		return "", fmt.Errorf("array.sort requires 1-2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.sort: %w", err)
	}

	order := `"order.ascending"`
	if len(call.Arguments) == 2 {
		if memberExpr, ok := call.Arguments[1].(*ast.MemberExpression); ok {
			if objID, ok := memberExpr.Object.(*ast.Identifier); ok && objID.Name == "order" {
				if propID, ok := memberExpr.Property.(*ast.Identifier); ok {
					order = fmt.Sprintf(`"order.%s"`, propID.Name)
				}
			}
		} else {
			orderCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
			if err != nil {
				return "", fmt.Errorf("array.sort: order: %w", err)
			}
			order = orderCode
		}
	}

	return fmt.Sprintf("%s.Sort(%s%s, %s)",
		elemType.TransformerConstructor(), arrayVar, elemType.VariableSuffix(), order), nil
}

func (h *ArrayMutatorCodegen) generateConcat(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.concat requires 2 arguments, got %d", len(call.Arguments))
	}

	array1Var, elemType, err := h.extractArrayVariable(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.concat: first array: %w", err)
	}

	array2Var, _, err := h.extractArrayVariable(g, call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.concat: second array: %w", err)
	}

	return fmt.Sprintf("%s.Concat(%s%s, 0, %s%s, 0)",
		elemType.TransformerConstructor(),
		array1Var, elemType.VariableSuffix(),
		array2Var, elemType.VariableSuffix()), nil
}

func (h *ArrayMutatorCodegen) extractArrayVariable(g *generator, arg ast.Expression) (string, ArrayElementType, error) {
	if id, ok := arg.(*ast.Identifier); ok {
		elemType, ok := g.lookupArrayElementType(id.Name)
		if !ok {
			return "", 0, fmt.Errorf("variable %q is not an array", id.Name)
		}
		return id.Name, elemType, nil
	}
	return "", 0, fmt.Errorf("expected array variable identifier")
}

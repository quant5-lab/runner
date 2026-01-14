package codegen

import "github.com/quant5-lab/runner/ast"

type ConstantResolver struct {
	registry     *PineConstantRegistry
	keyExtractor *ConstantKeyExtractor
}

func NewConstantResolver() *ConstantResolver {
	return &ConstantResolver{
		registry:     NewPineConstantRegistry(),
		keyExtractor: NewConstantKeyExtractor(),
	}
}

func (cr *ConstantResolver) ResolveToBool(expr ast.Expression) (bool, bool) {
	if literalValue, ok := cr.tryLiteralBool(expr); ok {
		return literalValue, true
	}

	if constantValue, ok := cr.tryConstantBool(expr); ok {
		return constantValue, true
	}

	return false, false
}

func (cr *ConstantResolver) ResolveToInt(expr ast.Expression) (int, bool) {
	if literalValue, ok := cr.tryLiteralInt(expr); ok {
		return literalValue, true
	}

	if constantValue, ok := cr.tryConstantInt(expr); ok {
		return constantValue, true
	}

	return 0, false
}

func (cr *ConstantResolver) ResolveToFloat(expr ast.Expression) (float64, bool) {
	if literalValue, ok := cr.tryLiteralFloat(expr); ok {
		return literalValue, true
	}

	if constantValue, ok := cr.tryConstantFloat(expr); ok {
		return constantValue, true
	}

	return 0.0, false
}

func (cr *ConstantResolver) ResolveToString(expr ast.Expression) (string, bool) {
	if literalValue, ok := cr.tryLiteralString(expr); ok {
		return literalValue, true
	}

	if constantValue, ok := cr.tryConstantString(expr); ok {
		return constantValue, true
	}

	return "", false
}

func (cr *ConstantResolver) tryLiteralBool(expr ast.Expression) (bool, bool) {
	if lit, ok := expr.(*ast.Literal); ok {
		if boolVal, ok := lit.Value.(bool); ok {
			return boolVal, true
		}
	}
	return false, false
}

func (cr *ConstantResolver) tryLiteralInt(expr ast.Expression) (int, bool) {
	if lit, ok := expr.(*ast.Literal); ok {
		if intVal, ok := lit.Value.(int); ok {
			return intVal, true
		}
	}
	return 0, false
}

func (cr *ConstantResolver) tryLiteralFloat(expr ast.Expression) (float64, bool) {
	if lit, ok := expr.(*ast.Literal); ok {
		if floatVal, ok := lit.Value.(float64); ok {
			return floatVal, true
		}
		if intVal, ok := lit.Value.(int); ok {
			return float64(intVal), true
		}
	}
	return 0.0, false
}

func (cr *ConstantResolver) tryLiteralString(expr ast.Expression) (string, bool) {
	if lit, ok := expr.(*ast.Literal); ok {
		if strVal, ok := lit.Value.(string); ok {
			return strVal, true
		}
	}
	return "", false
}

func (cr *ConstantResolver) tryConstantBool(expr ast.Expression) (bool, bool) {
	key, keyOk := cr.keyExtractor.ExtractFromExpression(expr)
	if !keyOk {
		return false, false
	}

	constVal, constOk := cr.registry.Get(key)
	if !constOk {
		return false, false
	}

	return constVal.AsBool()
}

func (cr *ConstantResolver) tryConstantInt(expr ast.Expression) (int, bool) {
	key, keyOk := cr.keyExtractor.ExtractFromExpression(expr)
	if !keyOk {
		return 0, false
	}

	constVal, constOk := cr.registry.Get(key)
	if !constOk {
		return 0, false
	}

	return constVal.AsInt()
}

func (cr *ConstantResolver) tryConstantFloat(expr ast.Expression) (float64, bool) {
	key, keyOk := cr.keyExtractor.ExtractFromExpression(expr)
	if !keyOk {
		return 0.0, false
	}

	constVal, constOk := cr.registry.Get(key)
	if !constOk {
		return 0.0, false
	}

	return constVal.AsFloat()
}

func (cr *ConstantResolver) tryConstantString(expr ast.Expression) (string, bool) {
	key, keyOk := cr.keyExtractor.ExtractFromExpression(expr)
	if !keyOk {
		return "", false
	}

	constVal, constOk := cr.registry.Get(key)
	if !constOk {
		return "", false
	}

	return constVal.AsString()
}

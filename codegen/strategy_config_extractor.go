package codegen

import "github.com/quant5-lab/runner/ast"

// StrategyConfigExtractor extracts configuration from strategy() declarations.
type StrategyConfigExtractor struct {
	propertyParser *PropertyParser
}

// NewStrategyConfigExtractor creates an extractor.
func NewStrategyConfigExtractor() *StrategyConfigExtractor {
	return &StrategyConfigExtractor{
		propertyParser: NewPropertyParser(),
	}
}

// ExtractFromCall parses strategy() call arguments into config.
func (e *StrategyConfigExtractor) ExtractFromCall(call *ast.CallExpression) *StrategyConfig {
	config := NewStrategyConfig()

	if len(call.Arguments) == 0 {
		return config
	}

	e.extractNameFromFirstArgument(call.Arguments[0], config)
	e.extractPropertiesFromObjectArguments(call.Arguments, config)

	return config
}

func (e *StrategyConfigExtractor) extractNameFromFirstArgument(arg ast.Expression, config *StrategyConfig) {
	if lit, ok := arg.(*ast.Literal); ok {
		if name, ok := lit.Value.(string); ok {
			config.Name = name
		}
	}
}

func (e *StrategyConfigExtractor) extractPropertiesFromObjectArguments(args []ast.Expression, config *StrategyConfig) {
	for _, arg := range args {
		if obj, ok := arg.(*ast.ObjectExpression); ok {
			e.extractFromObject(obj, config)
		}
	}
}

func (e *StrategyConfigExtractor) extractFromObject(obj *ast.ObjectExpression, config *StrategyConfig) {
	if val, ok := e.propertyParser.ParseFloat(obj, "default_qty_value"); ok {
		config.DefaultQtyValue = val
	}

	if val, ok := e.propertyParser.ParseFloat(obj, "initial_capital"); ok {
		config.InitialCapital = val
	}

	if val, ok := e.propertyParser.ParseIdentifier(obj, "default_qty_type"); ok {
		config.DefaultQtyType = val
	}

	if val, ok := e.propertyParser.ParseInt(obj, "pyramiding"); ok {
		config.Pyramiding = val
	}
}

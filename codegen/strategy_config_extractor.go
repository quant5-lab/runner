package codegen

import "github.com/quant5-lab/runner/ast"

type StrategyConfigExtractor struct {
	propertyParser *PropertyParser
}

func NewStrategyConfigExtractor() *StrategyConfigExtractor {
	return &StrategyConfigExtractor{
		propertyParser: NewPropertyParser(),
	}
}

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

	if val, ok := e.propertyParser.ParseBool(obj, "process_orders_on_close"); ok {
		config.ProcessOrdersOnClose = val
	}

	if val, ok := e.propertyParser.ParseIdentifier(obj, "commission_type"); ok {
		config.CommissionType = normalizeCommissionType(val)
	}

	if val, ok := e.propertyParser.ParseFloat(obj, "commission_value"); ok {
		config.CommissionValue = val
	}
}

/* normalizeCommissionType maps Pine commission.* identifiers to runtime constants. */
func normalizeCommissionType(identifier string) string {
	switch identifier {
	case "strategy.commission.percent", "commission.percent", "percent":
		return "percent"
	case "strategy.commission.cash_per_order", "commission.cash_per_order", "cash_per_order":
		return "cash_per_order"
	case "strategy.commission.cash_per_contract", "commission.cash_per_contract", "cash_per_contract":
		return "cash_per_contract"
	default:
		return identifier
	}
}

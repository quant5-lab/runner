package codegen

import "github.com/quant5-lab/runner/ast"

/* CompositeIndicatorMetadata provides internal series requirements for composite indicators.
 * Enables generic series declaration without hardcoding indicator-specific logic in generator.
 */
type CompositeIndicatorMetadata interface {
	/* GetInternalSeriesNames returns internal series required by the indicator */
	GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error)
}

/* CompositeIndicatorRegistry maps TA functions to their metadata providers */
type CompositeIndicatorRegistry struct {
	providers map[string]CompositeIndicatorMetadata
}

func NewCompositeIndicatorRegistry() *CompositeIndicatorRegistry {
	return &CompositeIndicatorRegistry{
		providers: make(map[string]CompositeIndicatorMetadata),
	}
}

/* Register adds a composite indicator handler to the registry */
func (r *CompositeIndicatorRegistry) Register(funcName string, provider CompositeIndicatorMetadata) {
	r.providers[funcName] = provider
}

/* GetInternalSeriesNames returns internal series for a TA function call.
 * Returns empty slice if function is not registered.
 */
func (r *CompositeIndicatorRegistry) GetInternalSeriesNames(funcName, varName string, call *ast.CallExpression) []string {
	if r == nil || r.providers == nil {
		return []string{}
	}

	normalizedName := funcName
	if len(funcName) > 3 && funcName[:3] != "ta." {
		normalizedName = "ta." + funcName
	}

	provider := r.providers[normalizedName]
	if provider == nil {
		provider = r.providers[funcName]
	}

	if provider == nil {
		return []string{}
	}

	seriesNames, err := provider.GetInternalSeriesNames(varName, call)
	if err != nil {
		return []string{}
	}

	return seriesNames
}

/* IsCompositeIndicator checks if a function requires internal series */
func (r *CompositeIndicatorRegistry) IsCompositeIndicator(funcName string) bool {
	if r == nil || r.providers == nil {
		return false
	}

	normalizedName := funcName
	if len(funcName) > 3 && funcName[:3] != "ta." {
		normalizedName = "ta." + funcName
	}

	return r.providers[normalizedName] != nil || r.providers[funcName] != nil
}

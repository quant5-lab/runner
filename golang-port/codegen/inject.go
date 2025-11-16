package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

/* StrategyCode holds generated Go code for strategy execution */
type StrategyCode struct {
	FunctionBody       string // executeStrategy() function body
	StrategyName       string // Pine Script strategy name
	NeedsSeriesPreCalc bool   // Whether TA pre-calculation imports are needed
}

/* InjectStrategy reads template, injects strategy code, writes output */
func InjectStrategy(templatePath, outputPath string, code *StrategyCode) error {
	// Read template
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	template := string(templateBytes)

	// Generate function with strategy code
	strategyFunc := fmt.Sprintf(`func executeStrategy(ctx *context.Context) (*output.Collector, *strategy.Strategy) {
	collector := output.NewCollector()
	strat := strategy.NewStrategy()

%s

	return collector, strat
}`, code.FunctionBody)

	// Replace placeholders
	output := strings.Replace(template, "{{STRATEGY_FUNC}}", strategyFunc, 1)
	output = strings.Replace(output, "{{STRATEGY_NAME}}", code.StrategyName, 1)
	
	// Conditional imports based on code requirements
	if !code.NeedsSeriesPreCalc {
		// Remove ta and value imports if not used
		output = strings.Replace(output, `"github.com/borisquantlab/pinescript-go/runtime/ta"`, `// ta import not needed`, 1)
		output = strings.Replace(output, `_ "github.com/borisquantlab/pinescript-go/runtime/value"`, `// value import not needed`, 1)
	}

	// Write output file
	err = os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

/* GenerateStrategyCode converts parsed Pine AST to Go code */
func GenerateStrategyCode(astJSON []byte) (*StrategyCode, error) {
	// Parse JSON to ESTree AST
	var program map[string]interface{}
	err := json.Unmarshal(astJSON, &program)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AST JSON: %w", err)
	}

	// For now, return placeholder code
	// TODO: Implement full AST traversal and code generation
	code := &StrategyCode{
		FunctionBody: `	// Strategy code will be generated here
	strat.Call("Generated Strategy", 10000)
	
	for i := 0; i < len(ctx.Data); i++ {
		ctx.BarIndex = i
		strat.OnBarUpdate(i, ctx.Data[i].Open, ctx.Data[i].Time)
		
		// Strategy logic placeholder
		// TODO: Generate from Pine AST
	}`,
	}

	return code, nil
}

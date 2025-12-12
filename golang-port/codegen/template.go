package codegen

import (
	"fmt"
	"os"
	"strings"
)

/* InjectStrategy reads template, injects strategy code, writes output */
func InjectStrategy(templatePath, outputPath string, code *StrategyCode) error {
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	template := string(templateBytes)

	/* Generate function with strategy code (securityContexts map parameter for security() support) */
	strategyFunc := fmt.Sprintf(`func executeStrategy(ctx *context.Context, dataDir string, securityContexts map[string]*context.Context) (*output.Collector, *strategy.Strategy) {
	collector := output.NewCollector()
	strat := strategy.NewStrategy()

%s

	return collector, strat
}`, code.FunctionBody)

	/* Replace placeholders */
	output := strings.Replace(template, "{{STRATEGY_FUNC}}", strategyFunc, 1)
	output = strings.Replace(output, "{{STRATEGY_NAME}}", code.StrategyName, 1)

	if len(code.AdditionalImports) > 0 {
		importMarker := "\t\"github.com/quant5-lab/runner/datafetcher\""
		if strings.Contains(output, importMarker) {
			additionalImportsStr := ""
			for _, imp := range code.AdditionalImports {
				if imp != "github.com/quant5-lab/runner/datafetcher" {
					if imp == "github.com/quant5-lab/runner/ast" {
						if !strings.Contains(code.FunctionBody, "&ast.") {
							continue
						}
					}
					additionalImportsStr += fmt.Sprintf("\t\"%s\"\n", imp)
				}
			}
			output = strings.Replace(output, importMarker, importMarker+"\n"+additionalImportsStr, 1)
		}
	}

	err = os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

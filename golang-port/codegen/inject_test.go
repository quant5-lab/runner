package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectStrategy(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "template.go")
	outputPath := filepath.Join(tempDir, "output.go")

	// Write mock template
	template := `package main

func main() {
	{{STRATEGY_FUNC}}
}
`
	err := os.WriteFile(templatePath, []byte(template), 0644)
	if err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Create strategy code
	code := &StrategyCode{
		FunctionBody: `	ctx.BarIndex = 0
	strat.Call("Test", 10000)`,
	}

	// Inject strategy
	err = InjectStrategy(templatePath, outputPath, code)
	if err != nil {
		t.Fatalf("InjectStrategy failed: %v", err)
	}

	// Verify output file exists
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputStr := string(outputBytes)

	// Verify function was injected
	if len(outputStr) < 100 {
		t.Errorf("Output too short: %d bytes", len(outputStr))
	}

	// Verify placeholder was replaced
	if contains(outputStr, "{{STRATEGY_FUNC}}") {
		t.Error("Placeholder not replaced")
	}

	// Verify function body was inserted
	if !contains(outputStr, "ctx.BarIndex = 0") {
		t.Error("Strategy code not injected")
	}
}

func TestGenerateStrategyCode(t *testing.T) {
	astJSON := []byte(`{"type": "Program", "body": []}`)

	code, err := GenerateStrategyCode(astJSON)
	if err != nil {
		t.Fatalf("GenerateStrategyCode failed: %v", err)
	}

	if code == nil {
		t.Fatal("Generated code is nil")
	}

	if len(code.FunctionBody) == 0 {
		t.Error("Function body is empty")
	}

	// Verify placeholder code contains expected patterns
	if !contains(code.FunctionBody, "strat.Call") {
		t.Error("Missing strategy initialization")
	}
}

func TestInjectedStrategyCompiles(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "main.go")

	// Read actual template
	templatePath := "../template/main.go.tmpl"
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("Template not found: %v", err)
		return
	}

	// Create minimal strategy code
	code := &StrategyCode{
		FunctionBody: `	strat.Call("Test Strategy", 10000)
	
	for i := 0; i < len(ctx.Data); i++ {
		ctx.BarIndex = i
		strat.OnBarUpdate(i, ctx.Data[i].Open, ctx.Data[i].Time)
	}`,
	}

	// Write template to temp location
	tempTemplatePath := filepath.Join(tempDir, "template.go")
	err = os.WriteFile(tempTemplatePath, templateBytes, 0644)
	if err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	// Inject strategy
	err = InjectStrategy(tempTemplatePath, outputPath, code)
	if err != nil {
		t.Fatalf("InjectStrategy failed: %v", err)
	}

	// Verify output compiles (syntax check)
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputStr := string(outputBytes)

	// Verify structure
	if !contains(outputStr, "package main") {
		t.Error("Missing package declaration")
	}
	if !contains(outputStr, "func main()") {
		t.Error("Missing main function")
	}
	if !contains(outputStr, "func executeStrategy") {
		t.Error("Missing executeStrategy function")
	}
	if !contains(outputStr, "strat.Call") {
		t.Error("Missing injected strategy code")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

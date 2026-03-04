package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTAConditionalDeduplication_CompileRegression(t *testing.T) {
	testDir := t.TempDir()

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	strategyPath := filepath.Join(projectRoot, "tests/fixtures/integration/test-ta-conditional-deduplication.pine")
	outputPath := filepath.Join(testDir, "test.go")

	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	cmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputPath,
		"-template", templatePath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Codegen failed: %v\n%s", err, output)
	}

	generatedFile := extractGeneratedPath(t, output)
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	code := string(content)

	rsiCount := strings.Count(code, "ta_rsi_14")
	if rsiCount == 0 {
		t.Fatal("Expected RSI temp var generation, found none")
	}

	if strings.Contains(code, "redeclared") {
		t.Fatal("Generated code contains redeclaration error")
	}

	localFile := filepath.Join(testDir, "test.go")
	if err := os.WriteFile(localFile, content, 0644); err != nil {
		t.Fatalf("Failed to copy generated file: %v", err)
	}

	if err := setupGoMod(localFile, projectRoot); err != nil {
		t.Fatalf("Failed to setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = testDir
	if tidyOutput, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOutput)
	}

	buildCmd := exec.Command("go", "build", localFile)
	buildCmd.Dir = testDir
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("Generated code failed to compile: %v\n%s", buildErr, buildOutput)
	}
}

func TestConditionalBooleanContext_CompileRegression(t *testing.T) {
	testDir := t.TempDir()

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	strategyPath := filepath.Join(projectRoot, "tests/fixtures/integration/test-conditional-boolean-context.pine")
	outputPath := filepath.Join(testDir, "test.go")

	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	cmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputPath,
		"-template", templatePath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Codegen failed: %v\n%s", err, output)
	}

	generatedFile := extractGeneratedPath(t, output)
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	code := string(content)

	if !strings.Contains(code, "value.IsTrue(") {
		t.Fatal("Expected value.IsTrue() wrapper for conditional in if-statement")
	}

	ternaryPattern := "func() float64 { if"
	if !strings.Contains(code, ternaryPattern) {
		t.Fatal("Expected ternary IIFE pattern in generated code")
	}
}

func TestCompositeIndicatorTernary_CompileRegression(t *testing.T) {
	testDir := t.TempDir()

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	strategyPath := filepath.Join(projectRoot, "tests/fixtures/integration/test-composite-indicator-ternary.pine")
	outputPath := filepath.Join(testDir, "test.go")

	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	cmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputPath,
		"-template", templatePath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Codegen failed: %v\n%s", err, output)
	}

	generatedFile := extractGeneratedPath(t, output)
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	code := string(content)

	hasRSI := strings.Contains(code, "ta_rsi")
	hasMFI := strings.Contains(code, "ta_mfi")
	if !hasRSI || !hasMFI {
		t.Errorf("Expected both RSI and MFI indicators, hasRSI=%v hasMFI=%v", hasRSI, hasMFI)
	}

	if strings.Contains(code, "redeclared") {
		t.Fatal("Generated code contains redeclaration error")
	}

	localFile := filepath.Join(testDir, "test.go")
	if err := os.WriteFile(localFile, content, 0644); err != nil {
		t.Fatalf("Failed to copy generated file: %v", err)
	}

	if err := setupGoMod(localFile, projectRoot); err != nil {
		t.Fatalf("Failed to setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = testDir
	if tidyOutput, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOutput)
	}

	buildCmd := exec.Command("go", "build", localFile)
	buildCmd.Dir = testDir
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("Generated code failed to compile: %v\n%s", buildErr, buildOutput)
	}
}

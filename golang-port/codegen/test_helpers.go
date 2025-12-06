package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* contains checks if string s contains substring substr */
func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	if s == substr {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type CodeVerifier struct {
	code string
	t    *testing.T
}

func NewCodeVerifier(code string, t *testing.T) *CodeVerifier {
	return &CodeVerifier{code: code, t: t}
}

func (v *CodeVerifier) MustContain(patterns ...string) *CodeVerifier {
	for _, pattern := range patterns {
		if !strings.Contains(v.code, pattern) {
			v.t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, v.code)
		}
	}
	return v
}

func (v *CodeVerifier) MustNotContain(patterns ...string) *CodeVerifier {
	for _, pattern := range patterns {
		if strings.Contains(v.code, pattern) {
			v.t.Errorf("Found unexpected pattern: %q\nGenerated code:\n%s", pattern, v.code)
		}
	}
	return v
}

func (v *CodeVerifier) MustNotHavePlaceholders() *CodeVerifier {
	return v.MustNotContain("TODO", "math.NaN() //")
}

func (v *CodeVerifier) CountOccurrences(pattern string, expected int) *CodeVerifier {
	count := strings.Count(v.code, pattern)
	if count != expected {
		v.t.Errorf("Expected %d occurrences of %q, found %d", expected, pattern, count)
	}
	return v
}

func generateSecurityExpression(t *testing.T, varName string, expression ast.Expression) string {
	program := buildSecurityTestProgram(varName, expression)
	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	return generated.FunctionBody
}

func generatePlotExpression(t *testing.T, expression ast.Expression) string {
	program := buildPlotTestProgram(expression)
	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	return generated.FunctionBody
}

func generateMultiSecurityProgram(t *testing.T, vars map[string]ast.Expression) string {
	program := buildMultiSecurityTestProgram(vars)
	generated, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}
	return generated.FunctionBody
}

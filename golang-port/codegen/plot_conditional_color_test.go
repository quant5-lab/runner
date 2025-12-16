package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

// Helper builders for plot call AST construction

func PlotCall(valueExpr ast.Expression, options ...ast.Property) *ast.CallExpression {
	args := []ast.Expression{valueExpr}
	if len(options) > 0 {
		args = append(args, &ast.ObjectExpression{Properties: options})
	}
	return &ast.CallExpression{
		Callee:    Ident("plot"),
		Arguments: args,
	}
}

func ColorProp(colorExpr ast.Expression) ast.Property {
	return ast.Property{
		Key:   Ident("color"),
		Value: colorExpr,
	}
}

func TitleProp(title string) ast.Property {
	return ast.Property{
		Key:   Ident("title"),
		Value: Lit(title),
	}
}

func OffsetProp(offset int) ast.Property {
	return ast.Property{
		Key:   Ident("offset"),
		Value: Lit(float64(offset)),
	}
}

func ConditionalExpr(test, consequent, alternate ast.Expression) *ast.ConditionalExpression {
	return &ast.ConditionalExpression{
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func NaIdent() *ast.Identifier {
	return Ident("na")
}

// Tests for buildPlotOptions method

func TestBuildPlotOptions_NoOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{Title: "Test"}
	result := gen.buildPlotOptions(opts)

	if result != "nil" {
		t.Errorf("Expected 'nil', got %q", result)
	}
}

func TestBuildPlotOptions_WithOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(-5)),
	}
	result := gen.buildPlotOptions(opts)

	expected := `map[string]interface{}{"offset": -5}`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildPlotOptions_ZeroOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(0)),
	}
	result := gen.buildPlotOptions(opts)

	if result != "nil" {
		t.Errorf("Expected 'nil' for zero offset, got %q", result)
	}
}

// Tests for buildPlotOptionsWithNullColor method

func TestBuildPlotOptionsWithNullColor_NoOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{Title: "Test"}
	result := gen.buildPlotOptionsWithNullColor(opts)

	expected := `map[string]interface{}{"color": nil}`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildPlotOptionsWithNullColor_WithOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(-10)),
	}
	result := gen.buildPlotOptionsWithNullColor(opts)

	if !strings.Contains(result, `"color": nil`) {
		t.Error("Result should contain color: nil")
	}
	if !strings.Contains(result, `"offset": -10`) {
		t.Error("Result should contain offset: -10")
	}
}

func TestBuildPlotOptionsWithNullColor_PositiveOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(5)),
	}
	result := gen.buildPlotOptionsWithNullColor(opts)

	if !strings.Contains(result, `"offset": 5`) {
		t.Error("Result should contain positive offset: 5")
	}
}

// Tests for buildPlotOptionsWithColor method

func TestBuildPlotOptionsWithColor_ColorOnly(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{Title: "Test"}
	result := gen.buildPlotOptionsWithColor(opts, "#FF0000")

	expected := `map[string]interface{}{"color": "#FF0000"}`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildPlotOptionsWithColor_ColorAndOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(-3)),
	}
	result := gen.buildPlotOptionsWithColor(opts, "#00FF00")

	if !strings.Contains(result, `"color": "#00FF00"`) {
		t.Error("Result should contain color")
	}
	if !strings.Contains(result, `"offset": -3`) {
		t.Error("Result should contain offset")
	}
}

func TestBuildPlotOptionsWithColor_EmptyColor(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{Title: "Test"}
	result := gen.buildPlotOptionsWithColor(opts, "")

	if result != "nil" {
		t.Errorf("Expected 'nil' for empty color, got %q", result)
	}
}

func TestBuildPlotOptionsWithColor_EmptyColorWithOffset(t *testing.T) {
	gen := &generator{
		constEvaluator: validation.NewWarmupAnalyzer(),
	}

	opts := PlotOptions{
		Title:      "Test",
		OffsetExpr: Lit(float64(-2)),
	}
	result := gen.buildPlotOptionsWithColor(opts, "")

	if !strings.Contains(result, `"offset": -2`) {
		t.Error("Result should contain offset")
	}
	if strings.Contains(result, `"color"`) {
		t.Error("Result should not contain color field when color is empty")
	}
}

// Tests for extractColorLiteral method

func TestExtractColorLiteral_StringValue(t *testing.T) {
	gen := &generator{}

	expr := Lit("#FF0000")
	result := gen.extractColorLiteral(expr)

	if result != "#FF0000" {
		t.Errorf("Expected '#FF0000', got %q", result)
	}
}

func TestExtractColorLiteral_NonLiteral(t *testing.T) {
	gen := &generator{}

	expr := Ident("myColor")
	result := gen.extractColorLiteral(expr)

	if result != "" {
		t.Errorf("Expected empty string for non-literal, got %q", result)
	}
}

func TestExtractColorLiteral_NonStringLiteral(t *testing.T) {
	gen := &generator{}

	expr := Lit(123.45)
	result := gen.extractColorLiteral(expr)

	if result != "" {
		t.Errorf("Expected empty string for non-string literal, got %q", result)
	}
}

func TestExtractColorLiteral_NamedColor(t *testing.T) {
	gen := &generator{}

	expr := Lit("#00FF00")
	result := gen.extractColorLiteral(expr)

	if result != "#00FF00" {
		t.Errorf("Expected '#00FF00', got %q", result)
	}
}

// Tests for conditional color code generation from AST

// TestPlotConditionalColor_ConsequentNa_GeneratesNullColorInTrueBranch tests
// pattern: color = condition ? na : #FF0000
func TestPlotConditionalColor_ConsequentNa_GeneratesNullColorInTrueBranch(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							Ident("condition"),
							NaIdent(),
							Lit("#FF0000"),
						),
					},
					{
						Key:   Ident("title"),
						Value: Lit("Test"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("if ")
	verifier.CountOccurrences("collector.Add", 2)
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#FF0000"`)
	verifier.MustContain("null color")
}

// TestPlotConditionalColor_AlternateNa_GeneratesNullColorInFalseBranch tests
// pattern: color = condition ? #0000FF : na
func TestPlotConditionalColor_AlternateNa_GeneratesNullColorInFalseBranch(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							Ident("condition"),
							Lit("#0000FF"),
							NaIdent(),
						),
					},
					{
						Key:   Ident("title"),
						Value: Lit("Alternate"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("if ")
	verifier.CountOccurrences("collector.Add", 2)
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#0000FF"`)
	verifier.MustContain("null color")
}

// TestPlotConditionalColor_WithOffset_BothBranchesHaveOffset tests
// pattern: color = condition ? na : #00FF00, offset = -5
func TestPlotConditionalColor_WithOffset_BothBranchesHaveOffset(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							Ident("condition"),
							NaIdent(),
							Lit("#00FF00"),
						),
					},
					{
						Key:   Ident("offset"),
						Value: Lit(float64(-5)),
					},
					{
						Key:   Ident("title"),
						Value: Lit("OffsetTest"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.CountOccurrences(`"offset": -5`, 2)
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#00FF00"`)
}

// TestPlotConditionalColor_CallExpressionTest_AddsNotEqualZero tests
// pattern: color = change(close) ? na : #FF0000
func TestPlotConditionalColor_CallExpressionTest_AddsNotEqualZero(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							&ast.CallExpression{
								Callee:    Ident("change"),
								Arguments: []ast.Expression{Ident("close")},
							},
							NaIdent(),
							Lit("#FF0000"),
						),
					},
					{
						Key:   Ident("title"),
						Value: Lit("CallTest"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("!= 0")
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#FF0000"`)
}

// TestPlotStaticColor_NoConditional_SingleCollectorAdd tests static color
// without conditional logic
func TestPlotStaticColor_NoConditional_SingleCollectorAdd(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   Ident("color"),
						Value: Lit("#FF00FF"),
					},
					{
						Key:   Ident("title"),
						Value: Lit("Static"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.CountOccurrences("collector.Add", 1)
	verifier.MustNotContain(`"color": nil`)
	verifier.MustNotContain("null color")
}

// TestPlotNoColorOption_DefaultBehavior tests plot without color option
func TestPlotNoColorOption_DefaultBehavior(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   Ident("title"),
						Value: Lit("Default"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("collector.Add")
	verifier.MustNotContain(`"color"`)
}

// Edge case tests

// TestPlotConditionalColor_ComplexBooleanTest tests complex boolean expressions
func TestPlotConditionalColor_ComplexBooleanTest(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							BinaryExpr("and", Ident("a"), Ident("b")),
							Lit("#FFFF00"),
							NaIdent(),
						),
					},
					{
						Key:   Ident("title"),
						Value: Lit("Complex"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("if ")
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#FFFF00"`)
}

// TestPlotConditionalColor_ZeroOffsetIgnored tests zero offset handling
func TestPlotConditionalColor_ZeroOffsetIgnored(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							Ident("condition"),
							NaIdent(),
							Lit("#FFFFFF"),
						),
					},
					{
						Key:   Ident("offset"),
						Value: Lit(float64(0)),
					},
					{
						Key:   Ident("title"),
						Value: Lit("ZeroOffset"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)

	// Should NOT have offset: 0 in output
	verifier.MustNotContain(`"offset": 0`)

	// Should still have conditional color handling
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#FFFFFF"`)
}

// TestPlotConditionalColor_PositiveOffset tests positive offset handling
func TestPlotConditionalColor_PositiveOffset(t *testing.T) {
	gen := createPlotTestGenerator()

	plotCall := &ast.CallExpression{
		Callee: Ident("plot"),
		Arguments: []ast.Expression{
			Ident("value"),
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key: Ident("color"),
						Value: ConditionalExpr(
							Ident("condition"),
							Lit("#00FFFF"),
							NaIdent(),
						),
					},
					{
						Key:   Ident("offset"),
						Value: Lit(float64(3)),
					},
					{
						Key:   Ident("title"),
						Value: Lit("PositiveOffset"),
					},
				},
			},
		},
	}

	code, err := gen.generateVariableInit("test", plotCall)
	if err != nil {
		t.Fatalf("generateVariableInit failed: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.CountOccurrences(`"offset": 3`, 2)
	verifier.MustContain(`"color": nil`)
	verifier.MustContain(`"color": "#00FFFF"`)
}

// Test helpers

// createPlotTestGenerator creates a fully initialized generator for plot testing
func createPlotTestGenerator() *generator {
	gen := &generator{
		variables:               make(map[string]string),
		varInits:                make(map[string]ast.Expression),
		constants:               make(map[string]interface{}),
		taRegistry:              NewTAFunctionRegistry(),
		mathHandler:             NewMathHandler(),
		runtimeOnlyFilter:       NewRuntimeOnlyFunctionFilter(),
		barFieldRegistry:        NewBarFieldSeriesRegistry(),
		constEvaluator:          validation.NewWarmupAnalyzer(),
		inlineConditionRegistry: NewInlineConditionHandlerRegistry(),
		indent:                  1,
	}
	gen.typeSystem = NewTypeInferenceEngine()
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.builtinHandler = NewBuiltinIdentifierHandler()
	gen.boolConverter = NewBooleanConverter(gen.typeSystem)

	// Setup built-in variables
	gen.variables["close"] = "float64"
	gen.variables["open"] = "float64"
	gen.variables["high"] = "float64"
	gen.variables["low"] = "float64"
	gen.variables["volume"] = "float64"
	gen.variables["value"] = "float64"
	gen.variables["condition"] = "bool"
	gen.variables["a"] = "bool"
	gen.variables["b"] = "bool"

	return gen
}

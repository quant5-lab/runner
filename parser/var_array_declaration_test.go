package parser

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// parseOne parses Pine source into a Program AST, failing the test on any error.
func parseOne(t *testing.T, source string) *ast.Program {
	t.Helper()
	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser init: %v", err)
	}
	script, err := p.ParseString("", source)
	if err != nil {
		t.Fatalf("parse %q: %v", source, err)
	}
	conv := NewConverter()
	program, err := conv.ToESTree(script)
	if err != nil {
		t.Fatalf("convert %q: %v", source, err)
	}
	return program
}

// TestVarAssignment_TypedArrayDeclarations validates the `var type[] name = init` grammar
// variant for all supported scalar and drawing-object types.
//
// Scalar types (float, int, bool, string, color) parse and preserve the array declaration.
// Drawing types (line, label, box, table, linefill, polyline) degrade to NaN in the init
// expression while the variable still parses correctly as a var-persisted declaration.
func TestVarAssignment_TypedArrayDeclarations(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varName     string
		persistence string
		nanDegraded bool
	}{
		{
			name:        "var float[] scalar array declaration",
			source:      `var float[] prices = array.new_float()`,
			varName:     "prices",
			persistence: "var",
		},
		{
			name:        "var int[] scalar array declaration",
			source:      `var int[] counts = array.new_int(5)`,
			varName:     "counts",
			persistence: "var",
		},
		{
			name:        "var bool[] scalar array declaration",
			source:      `var bool[] flags = array.new_bool()`,
			varName:     "flags",
			persistence: "var",
		},
		{
			name:        "var string[] scalar array declaration",
			source:      `var string[] labels = array.new_string(0, "")`,
			varName:     "labels",
			persistence: "var",
		},
		{
			name:        "var color[] scalar array declaration",
			source:      `var color[] colors = array.new_color()`,
			varName:     "colors",
			persistence: "var",
		},
		{
			name:        "varip float[] scalar array declaration",
			source:      `varip float[] running = array.new_float()`,
			varName:     "running",
			persistence: "varip",
		},
		{
			name:        "var line[] drawing-type array declaration degrades to NaN",
			source:      `var line[] resistanceLines = array.new_line()`,
			varName:     "resistanceLines",
			persistence: "var",
			nanDegraded: true,
		},
		{
			name:        "var label[] drawing-type array declaration degrades to NaN",
			source:      `var label[] tradeLabels = array.new_label()`,
			varName:     "tradeLabels",
			persistence: "var",
			nanDegraded: true,
		},
		{
			name:        "var box[] drawing-type array declaration degrades to NaN",
			source:      `var box[] zones = array.new_box()`,
			varName:     "zones",
			persistence: "var",
			nanDegraded: true,
		},
		{
			name:        "var table[] drawing-type array declaration degrades to NaN",
			source:      `var table[] dashboards = array.new_table()`,
			varName:     "dashboards",
			persistence: "var",
			nanDegraded: true,
		},
		{
			name:        "var linefill[] drawing-type array declaration degrades to NaN",
			source:      `var linefill[] fills = array.new_linefill()`,
			varName:     "fills",
			persistence: "var",
			nanDegraded: true,
		},
		{
			name:        "var polyline[] drawing-type array declaration degrades to NaN",
			source:      `var polyline[] shapes = array.new_polyline()`,
			varName:     "shapes",
			persistence: "var",
			nanDegraded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseOne(t, tt.source)

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("expected VariableDeclaration for %q, not found in AST", tt.varName)
			}

			if decl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, decl.Persistence)
			}
			if decl.Kind != "let" {
				t.Errorf("Kind: expected \"let\" (declaration), got %q", decl.Kind)
			}
			if len(decl.Declarations) == 0 || decl.Declarations[0].Init == nil {
				t.Error("expected Init expression, got nil")
			}

			if tt.nanDegraded {
				lit, ok := decl.Declarations[0].Init.(*ast.Literal)
				if !ok {
					t.Errorf("drawing type: expected NaN-degraded Literal init, got %T", decl.Declarations[0].Init)
				} else if !strings.Contains(lit.Raw, "math.NaN()") {
					t.Errorf("drawing type: expected NaN marker in Raw, got %q", lit.Raw)
				}
			}
		})
	}
}

// TestVarAssignment_TypedArrayDeclarations_NoRegressionScalar verifies that scalar
// typed declarations WITHOUT the [] suffix still parse correctly after the grammar change.
func TestVarAssignment_TypedArrayDeclarations_NoRegressionScalar(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		varName     string
		persistence string
	}{
		{
			name:        "var float without array suffix still parses",
			source:      `var float x = 0.0`,
			varName:     "x",
			persistence: "var",
		},
		{
			name:        "var int without array suffix still parses",
			source:      `var int counter = 0`,
			varName:     "counter",
			persistence: "var",
		},
		{
			name:        "var bool without array suffix still parses",
			source:      `var bool flag = false`,
			varName:     "flag",
			persistence: "var",
		},
		{
			name:        "var string without array suffix still parses",
			source:      `var string msg = "hello"`,
			varName:     "msg",
			persistence: "var",
		},
		{
			name:        "var color without array suffix still parses",
			source:      `var color lineColor = na`,
			varName:     "lineColor",
			persistence: "var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseOne(t, tt.source)

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("expected VariableDeclaration for %q, not found", tt.varName)
			}
			if decl.Persistence != tt.persistence {
				t.Errorf("Persistence: expected %q, got %q", tt.persistence, decl.Persistence)
			}
			if decl.Kind != "let" {
				t.Errorf("Kind: expected \"let\", got %q", decl.Kind)
			}
		})
	}
}

// TestTypedAssignment_ArraySuffix verifies that typed assignment declarations (without var/varip)
// that include the [] suffix also parse correctly.
func TestTypedAssignment_ArraySuffix(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		varName string
	}{
		{
			name:    "float[] typed assignment",
			source:  `float[] prices = array.new_float()`,
			varName: "prices",
		},
		{
			name:    "int[] typed assignment",
			source:  `int[] counts = array.new_int(5)`,
			varName: "counts",
		},
		{
			name:    "bool[] typed assignment",
			source:  `bool[] flags = array.new_bool()`,
			varName: "flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseOne(t, tt.source)

			decl := findVariableDeclaration(program, tt.varName)
			if decl == nil {
				t.Fatalf("expected VariableDeclaration for %q, not found", tt.varName)
			}
			if decl.Kind != "let" {
				t.Errorf("Kind: expected \"let\", got %q", decl.Kind)
			}
			if len(decl.Declarations) == 0 || decl.Declarations[0].Init == nil {
				t.Error("expected Init expression, got nil")
			}
		})
	}
}

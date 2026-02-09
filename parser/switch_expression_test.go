package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Switch as expression tests */

func TestSwitchExpression_AssignmentContexts(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		varName string
	}{
		{
			name: "simple variable assignment",
			source: `result = switch mode
    1 =>
        10
    2 =>
        20`,
			varName: "result",
		},
		{
			name: "reassignment",
			source: `x = 0
x := switch state
    1 =>
        100
    2 =>
        200`,
			varName: "x",
		},
		{
			name: "var declaration",
			source: `var signal = switch condition
    true =>
        1
    false =>
        0`,
			varName: "signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}

			var foundSwitchExpr bool
			for _, stmt := range program.Body {
				if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
					for _, decl := range varDecl.Declarations {
						if _, ok := decl.Init.(*ast.IfStatement); ok {
							foundSwitchExpr = true
							if ident, ok := decl.ID.(*ast.Identifier); ok {
								if ident.Name != tt.varName {
									t.Errorf("variable name: expected=%q got=%q", tt.varName, ident.Name)
								}
							}
						}
					}
				}
			}

			if !foundSwitchExpr {
				t.Error("switch expression (lowered to IfStatement) not found")
			}
		})
	}
}

func TestSwitchExpression_BranchComplexity(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "arithmetic in branches",
			source: `result = switch mode
    1 =>
        close * 2
    2 =>
        close / 2`,
		},
		{
			name: "function calls in branches",
			source: `result = switch type
    1 =>
        ta.sma(close, 10)
    2 =>
        ta.ema(close, 20)`,
		},
		{
			name: "complex expressions in branches",
			source: `result = switch signal
    1 =>
        (high + low) / 2
    2 =>
        ta.crossover(close, open)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}
		})
	}
}

func TestSwitchExpression_SyntacticContexts(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "standalone assignment",
			source: `result = switch close > open
    true =>
        1
    false =>
        0`,
		},
		{
			name: "in var declaration",
			source: `var direction = switch trend
    1 =>
        1
    =>
        -1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}
		})
	}
}

func TestSwitchExpression_Form1WithComplexSubject(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "subject is binary expression",
			source: `result = switch close - open
    0 =>
        1
    =>
        2`,
		},
		{
			name: "subject is function call",
			source: `result = switch ta.sma(close, 10)
    100 =>
        1
    =>
        2`,
		},
		{
			name: "subject is member expression",
			source: `result = switch strategy.position_size
    0 =>
        1
    =>
        2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}
		})
	}
}

func TestSwitchExpression_Form2WithComplexConditions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "conditions with logical operators",
			source: `result = switch
    close > open and volume > 1000 =>
        1
    close < open or volume < 500 =>
        2`,
		},
		{
			name: "conditions with function calls",
			source: `result = switch
    ta.crossover(close, open) =>
        1
    ta.crossunder(close, open) =>
        2`,
		},
		{
			name: "conditions with member access",
			source: `result = switch
    strategy.position_size > 0 =>
        1
    strategy.position_size < 0 =>
        -1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}
		})
	}
}

func TestSwitchExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "single case with default",
			source: `result = switch val
    1 =>
        10
    =>
        0`,
		},
		{
			name: "many cases",
			source: `result = switch mode
    1 =>
        10
    2 =>
        20
    3 =>
        30
    4 =>
        40
    5 =>
        50
    =>
        0`,
		},
		{
			name: "form2 with complex default body",
			source: `result = switch
    condition1 =>
        value1
    =>
        ta.sma(close, 20) * 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}
		})
	}
}

func TestSwitchExpression_ConverterRobustness(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "form1 preserves subject expression",
			source: `result = switch high[1]
    100 =>
        1`,
		},
		{
			name: "form2 preserves all conditions",
			source: `result = switch
    close > open =>
        1
    close < open =>
        2`,
		},
		{
			name: "preserves case body expressions",
			source: `result = switch mode
    1 =>
        ta.sma(close, 10)
    2 =>
        ta.ema(close, 20)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseString("", tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("conversion produced empty program")
			}
		})
	}
}

package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestPivotHandlers_CanHandle(t *testing.T) {
	highHandler := &PivotHighHandler{}
	lowHandler := &PivotLowHandler{}

	tests := []struct {
		name     string
		funcName string
		wantHigh bool
		wantLow  bool
	}{
		{"namespaced pivothigh", "ta.pivothigh", true, false},
		{"namespaced pivotlow", "ta.pivotlow", false, true},
		{"non-namespaced pivothigh", "pivothigh", true, false},
		{"non-namespaced pivotlow", "pivotlow", false, true},
		{"ta.sma", "ta.sma", false, false},
		{"ta.ema", "ta.ema", false, false},
		{"random function", "myFunc", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHigh := highHandler.CanHandle(tt.funcName)
			gotLow := lowHandler.CanHandle(tt.funcName)

			if gotHigh != tt.wantHigh {
				t.Errorf("PivotHighHandler.CanHandle(%q) = %v, want %v", tt.funcName, gotHigh, tt.wantHigh)
			}
			if gotLow != tt.wantLow {
				t.Errorf("PivotLowHandler.CanHandle(%q) = %v, want %v", tt.funcName, gotLow, tt.wantLow)
			}
		})
	}
}

func TestPivotCodegen_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name      string
		arguments []ast.Expression
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid arguments",
			arguments: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "bar"},
					Property: &ast.Identifier{Name: "High"},
				},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
			expectErr: false,
		},
		{
			name: "valid 2-arg form",
			arguments: []ast.Expression{
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
			expectErr: false,
		},
		{
			name: "too few arguments - only one arg",
			arguments: []ast.Expression{
				&ast.Literal{Value: float64(2)},
			},
			expectErr: true,
			errMsg:    "requires 2 or 3 arguments",
		},
		{
			name:      "no arguments",
			arguments: []ast.Expression{},
			expectErr: true,
			errMsg:    "requires 2 or 3 arguments",
		},
		{
			name: "leftBars zero",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(0)},
				&ast.Literal{Value: float64(2)},
			},
			expectErr: true,
			errMsg:    "must be >= 1",
		},
		{
			name: "leftBars negative",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(-1)},
				&ast.Literal{Value: float64(2)},
			},
			expectErr: true,
			errMsg:    "must be >= 1",
		},
		{
			name: "rightBars zero",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(0)},
			},
			expectErr: true,
			errMsg:    "must be >= 1",
		},
		{
			name: "rightBars negative",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(-2)},
			},
			expectErr: true,
			errMsg:    "must be >= 1",
		},
		{
			name: "both bars zero",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(0)},
				&ast.Literal{Value: float64(0)},
			},
			expectErr: true,
			errMsg:    "must be >= 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			call := &ast.CallExpression{Arguments: tt.arguments}
			_, err := gen.generatePivot("testPivot", call, true)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPivotCodegen_GeneratedCodeStructure(t *testing.T) {
	tests := []struct {
		name             string
		isHigh           bool
		leftBars         int
		rightBars        int
		expectedPatterns []string
	}{
		{
			name:      "symmetric window 2,2",
			isHigh:    true,
			leftBars:  2,
			rightBars: 2,
			expectedPatterns: []string{
				"if i >= 4",
				"centerValue := ",
				"isPivot := true",
				"leftVal := ",
				"rightVal := ",
				"isPivot = false",
				"Series.Set(centerValue",
				"Series.Set(math.NaN()",
				">= centerValue",
			},
		},
		{
			name:      "asymmetric window 3,1",
			isHigh:    false,
			leftBars:  3,
			rightBars: 1,
			expectedPatterns: []string{
				"if i >= 4",
				"centerValue := ",
				"<= centerValue",
			},
		},
		{
			name:      "minimal window 1,1",
			isHigh:    true,
			leftBars:  1,
			rightBars: 1,
			expectedPatterns: []string{
				"if i >= 2",
				"centerValue := ",
				">= centerValue",
			},
		},
		{
			name:      "large asymmetric window 5,3",
			isHigh:    true,
			leftBars:  5,
			rightBars: 3,
			expectedPatterns: []string{
				"if i >= 8",
				"centerValue := ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "bar"},
						Property: &ast.Identifier{Name: "High"},
					},
					&ast.Literal{Value: float64(tt.leftBars)},
					&ast.Literal{Value: float64(tt.rightBars)},
				},
			}

			code, err := gen.generatePivot("testPivot", call, tt.isHigh)
			if err != nil {
				t.Fatalf("generatePivot() failed: %v", err)
			}

			for _, pattern := range tt.expectedPatterns {
				if !strings.Contains(code, pattern) {
					t.Errorf("Generated code missing pattern %q", pattern)
				}
			}
		})
	}
}

func TestPivotCodegen_NoFuturePeek(t *testing.T) {
	tests := []struct {
		name      string
		leftBars  int
		rightBars int
	}{
		{"small symmetric", 2, 2},
		{"large symmetric", 10, 10},
		{"asymmetric left heavy", 5, 2},
		{"asymmetric right heavy", 2, 5},
		{"minimal", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "bar"},
						Property: &ast.Identifier{Name: "High"},
					},
					&ast.Literal{Value: float64(tt.leftBars)},
					&ast.Literal{Value: float64(tt.rightBars)},
				},
			}

			code, err := gen.generatePivot("pivot", call, true)
			if err != nil {
				t.Fatalf("generatePivot() failed: %v", err)
			}

			futurePeekPatterns := []string{
				"ctx.Data[i+",
				"Series.Get(-",
				".Get(-",
			}

			for _, pattern := range futurePeekPatterns {
				if strings.Contains(code, pattern) {
					t.Errorf("Generated code contains future peek pattern %q", pattern)
				}
			}

			backwardPatterns := []string{
				"Series.Get(",
				"ctx.Data[i-",
			}

			hasBackward := false
			for _, pattern := range backwardPatterns {
				if strings.Contains(code, pattern) {
					hasBackward = true
					break
				}
			}

			if !hasBackward {
				t.Error("Generated code does not use any backward access pattern")
			}
		})
	}
}

func TestPivotCodegen_WindowSizeCalculations(t *testing.T) {
	tests := []struct {
		name           string
		leftBars       int
		rightBars      int
		expectedMinBar string
	}{
		{
			name:           "1,1 window",
			leftBars:       1,
			rightBars:      1,
			expectedMinBar: "if i >= 2",
		},
		{
			name:           "2,2 window",
			leftBars:       2,
			rightBars:      2,
			expectedMinBar: "if i >= 4",
		},
		{
			name:           "5,3 window",
			leftBars:       5,
			rightBars:      3,
			expectedMinBar: "if i >= 8",
		},
		{
			name:           "10,10 window",
			leftBars:       10,
			rightBars:      10,
			expectedMinBar: "if i >= 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "bar"},
						Property: &ast.Identifier{Name: "High"},
					},
					&ast.Literal{Value: float64(tt.leftBars)},
					&ast.Literal{Value: float64(tt.rightBars)},
				},
			}

			code, err := gen.generatePivot("pivot", call, true)
			if err != nil {
				t.Fatalf("generatePivot() failed: %v", err)
			}

			if !strings.Contains(code, tt.expectedMinBar) {
				t.Errorf("Expected bar check %q not found in code", tt.expectedMinBar)
			}
		})
	}
}

func TestPivotCodegen_HighVsLowComparison(t *testing.T) {
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "High"},
			},
			&ast.Literal{Value: float64(2)},
			&ast.Literal{Value: float64(2)},
		},
	}

	t.Run("pivot high uses >= comparison", func(t *testing.T) {
		code, err := gen.generatePivot("pivotHigh", call, true)
		if err != nil {
			t.Fatalf("generatePivot(high) failed: %v", err)
		}

		if !strings.Contains(code, ">= centerValue") {
			t.Error("Pivot high should use '>= centerValue' comparison")
		}
		if strings.Contains(code, "<= centerValue") {
			t.Error("Pivot high should not use '<=' comparison")
		}
	})

	t.Run("pivot low uses <= comparison", func(t *testing.T) {
		code, err := gen.generatePivot("pivotLow", call, false)
		if err != nil {
			t.Fatalf("generatePivot(low) failed: %v", err)
		}

		if !strings.Contains(code, "<= centerValue") {
			t.Error("Pivot low should use '<= centerValue' comparison")
		}
		if strings.Contains(code, ">= centerValue") {
			t.Error("Pivot low should not use '>=' comparison")
		}
	})
}

func TestPivotCodegen_EdgeCases(t *testing.T) {
	t.Run("multiple NaN branches", func(t *testing.T) {
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "bar"},
					Property: &ast.Identifier{Name: "High"},
				},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		code, err := gen.generatePivot("pivot", call, true)
		if err != nil {
			t.Fatalf("generatePivot() failed: %v", err)
		}

		nanCount := strings.Count(code, "Series.Set(math.NaN())")
		if nanCount != 3 {
			t.Errorf("Expected 3 NaN branches (bar insufficient, center NaN, not pivot), found %d", nanCount)
		}
	})

	t.Run("NaN check on centerValue", func(t *testing.T) {
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "source"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		code, err := gen.generatePivot("pivot", call, true)
		if err != nil {
			t.Fatalf("generatePivot() failed: %v", err)
		}

		if !strings.Contains(code, "!math.IsNaN(centerValue)") {
			t.Error("Missing NaN check on centerValue")
		}
	})

	t.Run("NaN check on neighbors", func(t *testing.T) {
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "source"},
				&ast.Literal{Value: float64(1)},
				&ast.Literal{Value: float64(1)},
			},
		}

		code, err := gen.generatePivot("pivot", call, true)
		if err != nil {
			t.Fatalf("generatePivot() failed: %v", err)
		}

		if !strings.Contains(code, "!math.IsNaN(leftVal)") {
			t.Error("Missing NaN check on left neighbor")
		}
		if !strings.Contains(code, "!math.IsNaN(rightVal)") {
			t.Error("Missing NaN check on right neighbor")
		}
	})
}

func TestPivotCodegen_SourceExpressions(t *testing.T) {
	tests := []struct {
		name       string
		sourceExpr ast.Expression
		expectPass bool
		desc       string
	}{
		{
			name: "bar.High member expression",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "High"},
			},
			expectPass: true,
			desc:       "Standard bar field access",
		},
		{
			name: "bar.Low member expression",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "Low"},
			},
			expectPass: true,
			desc:       "Alternative bar field",
		},
		{
			name:       "simple identifier",
			sourceExpr: &ast.Identifier{Name: "customSeries"},
			expectPass: true,
			desc:       "User series variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: float64(2)},
					&ast.Literal{Value: float64(2)},
				},
			}

			_, err := gen.generatePivot("pivot", call, true)

			if tt.expectPass && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.desc, err)
			}
			if !tt.expectPass && err == nil {
				t.Errorf("%s: expected error but got nil", tt.desc)
			}
		})
	}
}

func TestPivotCodegen_Integration(t *testing.T) {
	t.Run("handler delegates to generatePivot", func(t *testing.T) {
		gen := createTestGenerator()
		highHandler := &PivotHighHandler{}
		lowHandler := &PivotLowHandler{}

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "bar"},
					Property: &ast.Identifier{Name: "High"},
				},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		highCode, err := highHandler.GenerateCode(gen, "pivotHigh", call)
		if err != nil {
			t.Errorf("PivotHighHandler.GenerateCode() failed: %v", err)
		}
		if highCode == "" {
			t.Error("High handler should generate code")
		}

		lowCode, err := lowHandler.GenerateCode(gen, "pivotLow", call)
		if err != nil {
			t.Errorf("PivotLowHandler.GenerateCode() failed: %v", err)
		}
		if lowCode == "" {
			t.Error("Low handler should generate code")
		}
	})

	t.Run("variable naming consistency", func(t *testing.T) {
		gen := createTestGenerator()

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "source"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		code, err := gen.generatePivot("myPivot", call, true)
		if err != nil {
			t.Fatalf("generatePivot() failed: %v", err)
		}

		if !strings.Contains(code, "myPivotSeries.Set") {
			t.Error("Generated code should use consistent variable name 'myPivotSeries'")
		}
	})
}

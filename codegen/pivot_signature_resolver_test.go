package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestPivotSignatureResolver_SignaturePatterns(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	tests := []struct {
		name            string
		funcName        string
		args            []ast.Expression
		wantSourceName  string
		wantUsesDefault bool
		wantLeftValue   string
		wantRightValue  string
		wantErr         bool
	}{
		{
			name:            "ta.pivothigh 2-arg uses high default",
			funcName:        "ta.pivothigh",
			args:            []ast.Expression{&ast.Literal{Value: "5"}, &ast.Literal{Value: "5"}},
			wantSourceName:  "high",
			wantUsesDefault: true,
			wantLeftValue:   "5",
			wantRightValue:  "5",
			wantErr:         false,
		},
		{
			name:            "ta.pivotlow 2-arg uses low default",
			funcName:        "ta.pivotlow",
			args:            []ast.Expression{&ast.Literal{Value: "3"}, &ast.Literal{Value: "3"}},
			wantSourceName:  "low",
			wantUsesDefault: true,
			wantLeftValue:   "3",
			wantRightValue:  "3",
			wantErr:         false,
		},
		{
			name:            "ta.pivothigh 3-arg explicit source",
			funcName:        "ta.pivothigh",
			args:            []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: "5"}, &ast.Literal{Value: "5"}},
			wantSourceName:  "close",
			wantUsesDefault: false,
			wantLeftValue:   "5",
			wantRightValue:  "5",
			wantErr:         false,
		},
		{
			name:            "ta.pivotlow 3-arg explicit source",
			funcName:        "ta.pivotlow",
			args:            []ast.Expression{&ast.Identifier{Name: "open"}, &ast.Literal{Value: "2"}, &ast.Literal{Value: "2"}},
			wantSourceName:  "open",
			wantUsesDefault: false,
			wantLeftValue:   "2",
			wantRightValue:  "2",
			wantErr:         false,
		},
		{
			name:     "zero args returns error",
			funcName: "ta.pivothigh",
			args:     []ast.Expression{},
			wantErr:  true,
		},
		{
			name:     "one arg returns error",
			funcName: "ta.pivothigh",
			args:     []ast.Expression{&ast.Literal{Value: "5"}},
			wantErr:  true,
		},
		{
			name:     "four args returns error",
			funcName: "ta.pivothigh",
			args: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "extra"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			resolved, err := resolver.Resolve(tt.funcName, call)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Resolve() expected error, got nil")
				}
				if resolved != nil {
					t.Errorf("Resolve() with error should return nil, got %+v", resolved)
				}
				return
			}

			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}

			if resolved == nil {
				t.Fatal("Resolve() returned nil without error")
			}

			if resolved.UsesDefault != tt.wantUsesDefault {
				t.Errorf("UsesDefault = %v, want %v", resolved.UsesDefault, tt.wantUsesDefault)
			}

			sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
			if !ok {
				t.Fatalf("SourceExpr is not *ast.Identifier, got %T", resolved.SourceExpr)
			}
			if sourceIdent.Name != tt.wantSourceName {
				t.Errorf("SourceExpr.Name = %q, want %q", sourceIdent.Name, tt.wantSourceName)
			}

			leftLit, ok := resolved.LeftPeriod.(*ast.Literal)
			if !ok {
				t.Fatalf("LeftPeriod is not *ast.Literal, got %T", resolved.LeftPeriod)
			}
			if leftLit.Value != tt.wantLeftValue {
				t.Errorf("LeftPeriod.Value = %v, want %q", leftLit.Value, tt.wantLeftValue)
			}

			rightLit, ok := resolved.RightPeriod.(*ast.Literal)
			if !ok {
				t.Fatalf("RightPeriod is not *ast.Literal, got %T", resolved.RightPeriod)
			}
			if rightLit.Value != tt.wantRightValue {
				t.Errorf("RightPeriod.Value = %v, want %q", rightLit.Value, tt.wantRightValue)
			}
		})
	}
}

func TestPivotSignatureResolver_AsymmetricPeriods(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	tests := []struct {
		name      string
		funcName  string
		leftBars  string
		rightBars string
	}{
		{"asymmetric left heavy", "ta.pivothigh", "10", "2"},
		{"asymmetric right heavy", "ta.pivothigh", "2", "10"},
		{"extreme asymmetry", "ta.pivotlow", "20", "1"},
		{"single left bar", "ta.pivothigh", "1", "5"},
		{"single right bar", "ta.pivotlow", "5", "1"},
		{"equal periods small", "ta.pivothigh", "1", "1"},
		{"equal periods large", "ta.pivotlow", "50", "50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: tt.leftBars},
					&ast.Literal{Value: tt.rightBars},
				},
			}

			resolved, err := resolver.Resolve(tt.funcName, call)
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}

			if resolved == nil {
				t.Fatal("Resolve() returned nil")
			}

			leftLit := resolved.LeftPeriod.(*ast.Literal)
			if leftLit.Value != tt.leftBars {
				t.Errorf("LeftPeriod = %v, want %q", leftLit.Value, tt.leftBars)
			}

			rightLit := resolved.RightPeriod.(*ast.Literal)
			if rightLit.Value != tt.rightBars {
				t.Errorf("RightPeriod = %v, want %q", rightLit.Value, tt.rightBars)
			}
		})
	}
}

func TestPivotSignatureResolver_ExpressionTypes(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	tests := []struct {
		name       string
		sourceExpr ast.Expression
		leftExpr   ast.Expression
		rightExpr  ast.Expression
		wantSource string
	}{
		{
			name:       "identifier source",
			sourceExpr: &ast.Identifier{Name: "close"},
			leftExpr:   &ast.Literal{Value: "5"},
			rightExpr:  &ast.Literal{Value: "5"},
			wantSource: "close",
		},
		{
			name:       "high source",
			sourceExpr: &ast.Identifier{Name: "high"},
			leftExpr:   &ast.Literal{Value: "3"},
			rightExpr:  &ast.Literal{Value: "3"},
			wantSource: "high",
		},
		{
			name:       "low source",
			sourceExpr: &ast.Identifier{Name: "low"},
			leftExpr:   &ast.Literal{Value: "4"},
			rightExpr:  &ast.Literal{Value: "4"},
			wantSource: "low",
		},
		{
			name:       "open source",
			sourceExpr: &ast.Identifier{Name: "open"},
			leftExpr:   &ast.Literal{Value: "2"},
			rightExpr:  &ast.Literal{Value: "2"},
			wantSource: "open",
		},
		{
			name:       "volume source",
			sourceExpr: &ast.Identifier{Name: "volume"},
			leftExpr:   &ast.Literal{Value: "5"},
			rightExpr:  &ast.Literal{Value: "5"},
			wantSource: "volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{tt.sourceExpr, tt.leftExpr, tt.rightExpr},
			}

			resolved, err := resolver.Resolve("ta.pivothigh", call)
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}

			sourceIdent, ok := resolved.SourceExpr.(*ast.Identifier)
			if !ok {
				t.Fatalf("SourceExpr is not *ast.Identifier, got %T", resolved.SourceExpr)
			}

			if sourceIdent.Name != tt.wantSource {
				t.Errorf("SourceExpr.Name = %q, want %q", sourceIdent.Name, tt.wantSource)
			}

			if resolved.SourceExpr != tt.sourceExpr {
				t.Error("SourceExpr should preserve original expression reference")
			}
		})
	}
}

func TestPivotSignatureResolver_FunctionDetection(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.pivothigh", true},
		{"ta.pivotlow", true},
		{"ta.sma", false},
		{"ta.ema", false},
		{"ta.rma", false},
		{"ta.highest", false},
		{"ta.lowest", false},
		{"pivothigh", false},
		{"pivotlow", false},
		{"pivot", false},
		{"ta.pivot", false},
		{"", false},
		{"ta.", false},
		{"ta.pivothigh_custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			got := resolver.IsPivotFunction(tt.funcName)
			if got != tt.want {
				t.Errorf("IsPivotFunction(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestPivotSignatureResolver_ErrorMessages(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	tests := []struct {
		name          string
		funcName      string
		args          []ast.Expression
		wantErrSubstr string
	}{
		{
			name:          "zero args mentions required count",
			funcName:      "ta.pivothigh",
			args:          []ast.Expression{},
			wantErrSubstr: "2 or 3 arguments",
		},
		{
			name:          "one arg mentions required count",
			funcName:      "ta.pivotlow",
			args:          []ast.Expression{&ast.Literal{Value: "5"}},
			wantErrSubstr: "2 or 3 arguments",
		},
		{
			name:     "four args mentions required count",
			funcName: "ta.pivothigh",
			args: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "extra"},
			},
			wantErrSubstr: "2 or 3 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			_, err := resolver.Resolve(tt.funcName, call)

			if err == nil {
				t.Fatal("Resolve() expected error, got nil")
			}

			if !stringContains(err.Error(), tt.wantErrSubstr) {
				t.Errorf("Error message should contain %q, got: %v", tt.wantErrSubstr, err)
			}
		})
	}
}

func TestPivotSignatureResolver_Consistency(t *testing.T) {
	resolver := NewPivotSignatureResolver()

	t.Run("same input produces same output", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
			},
		}

		resolved1, err1 := resolver.Resolve("ta.pivothigh", call)
		resolved2, err2 := resolver.Resolve("ta.pivothigh", call)

		if err1 != nil || err2 != nil {
			t.Fatalf("Unexpected errors: %v, %v", err1, err2)
		}

		if resolved1.UsesDefault != resolved2.UsesDefault {
			t.Error("UsesDefault should be consistent")
		}

		source1 := resolved1.SourceExpr.(*ast.Identifier).Name
		source2 := resolved2.SourceExpr.(*ast.Identifier).Name
		if source1 != source2 {
			t.Errorf("SourceExpr inconsistent: %q vs %q", source1, source2)
		}
	})

	t.Run("high and low have different defaults", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
			},
		}

		resolvedHigh, _ := resolver.Resolve("ta.pivothigh", call)
		resolvedLow, _ := resolver.Resolve("ta.pivotlow", call)

		sourceHigh := resolvedHigh.SourceExpr.(*ast.Identifier).Name
		sourceLow := resolvedLow.SourceExpr.(*ast.Identifier).Name

		if sourceHigh != "high" {
			t.Errorf("pivothigh default = %q, want \"high\"", sourceHigh)
		}

		if sourceLow != "low" {
			t.Errorf("pivotlow default = %q, want \"low\"", sourceLow)
		}

		if sourceHigh == sourceLow {
			t.Error("pivothigh and pivotlow should have different default sources")
		}
	})
}

package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestBarFieldSeriesRegistry_GetSeriesName tests field-to-series mapping */
func TestBarFieldSeriesRegistry_GetSeriesName(t *testing.T) {
	registry := NewBarFieldSeriesRegistry()

	tests := []struct {
		name       string
		barField   string
		wantName   string
		wantExists bool
	}{
		{
			name:       "Close field",
			barField:   "bar.Close",
			wantName:   "closeSeries",
			wantExists: true,
		},
		{
			name:       "High field",
			barField:   "bar.High",
			wantName:   "highSeries",
			wantExists: true,
		},
		{
			name:       "Low field",
			barField:   "bar.Low",
			wantName:   "lowSeries",
			wantExists: true,
		},
		{
			name:       "Open field",
			barField:   "bar.Open",
			wantName:   "openSeries",
			wantExists: true,
		},
		{
			name:       "Volume field",
			barField:   "bar.Volume",
			wantName:   "volumeSeries",
			wantExists: true,
		},
		{
			name:       "Unknown field",
			barField:   "bar.Unknown",
			wantName:   "",
			wantExists: false,
		},
		{
			name:       "Non-bar field",
			barField:   "close",
			wantName:   "",
			wantExists: false,
		},
		{
			name:       "Empty string",
			barField:   "",
			wantName:   "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExists := registry.GetSeriesName(tt.barField)

			if gotExists != tt.wantExists {
				t.Errorf("GetSeriesName(%q) exists = %v, want %v",
					tt.barField, gotExists, tt.wantExists)
			}

			if gotName != tt.wantName {
				t.Errorf("GetSeriesName(%q) name = %q, want %q",
					tt.barField, gotName, tt.wantName)
			}
		})
	}
}

/* TestBarFieldSeriesRegistry_AllFields tests complete field enumeration */
func TestBarFieldSeriesRegistry_AllFields(t *testing.T) {
	registry := NewBarFieldSeriesRegistry()
	fields := registry.AllFields()

	expectedFields := []string{"Close", "High", "Low", "Open", "Volume"}

	if len(fields) != len(expectedFields) {
		t.Errorf("AllFields() returned %d fields, want %d", len(fields), len(expectedFields))
	}

	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[field] = true
	}

	for _, expected := range expectedFields {
		if !fieldSet[expected] {
			t.Errorf("AllFields() missing expected field %q", expected)
		}
	}
}

/* TestBarFieldSeriesRegistry_AllSeriesNames tests Series name enumeration */
func TestBarFieldSeriesRegistry_AllSeriesNames(t *testing.T) {
	registry := NewBarFieldSeriesRegistry()
	seriesNames := registry.AllSeriesNames()

	expectedNames := []string{"closeSeries", "highSeries", "lowSeries", "openSeries", "volumeSeries"}

	if len(seriesNames) != len(expectedNames) {
		t.Errorf("AllSeriesNames() returned %d names, want %d", len(seriesNames), len(expectedNames))
	}

	nameSet := make(map[string]bool)
	for _, name := range seriesNames {
		nameSet[name] = true
	}

	for _, expected := range expectedNames {
		if !nameSet[expected] {
			t.Errorf("AllSeriesNames() missing expected name %q", expected)
		}
	}
}

/* TestBarFieldSeriesRegistry_FieldNameConsistency tests field-to-series naming convention */
func TestBarFieldSeriesRegistry_FieldNameConsistency(t *testing.T) {
	registry := NewBarFieldSeriesRegistry()

	tests := []struct {
		field      string
		barField   string
		seriesName string
	}{
		{"Close", "bar.Close", "closeSeries"},
		{"High", "bar.High", "highSeries"},
		{"Low", "bar.Low", "lowSeries"},
		{"Open", "bar.Open", "openSeries"},
		{"Volume", "bar.Volume", "volumeSeries"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			gotName, exists := registry.GetSeriesName(tt.barField)

			if !exists {
				t.Errorf("Field %q not found in registry", tt.barField)
			}

			if gotName != tt.seriesName {
				t.Errorf("Field %q mapped to %q, want %q (naming convention violated)",
					tt.barField, gotName, tt.seriesName)
			}
		})
	}
}

/* TestBarFieldSeriesRegistry_Immutability tests that registry state doesn't change */
func TestBarFieldSeriesRegistry_Immutability(t *testing.T) {
	registry := NewBarFieldSeriesRegistry()

	// Get initial state
	fields1 := registry.AllFields()
	names1 := registry.AllSeriesNames()
	close1, exists1 := registry.GetSeriesName("bar.Close")

	// Call methods multiple times
	_ = registry.AllFields()
	_ = registry.AllSeriesNames()
	_, _ = registry.GetSeriesName("bar.High")
	_, _ = registry.GetSeriesName("bar.Unknown")

	// Get state again
	fields2 := registry.AllFields()
	names2 := registry.AllSeriesNames()
	close2, exists2 := registry.GetSeriesName("bar.Close")

	// Verify immutability
	if len(fields1) != len(fields2) {
		t.Error("AllFields() changed after multiple calls")
	}

	if len(names1) != len(names2) {
		t.Error("AllSeriesNames() changed after multiple calls")
	}

	if close1 != close2 || exists1 != exists2 {
		t.Error("GetSeriesName() changed after multiple calls")
	}
}

/* TestBarFieldSeriesCodegen_Declarations tests that bar field Series are declared */
func TestBarFieldSeriesCodegen_Declarations(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "mySignal"},
						Init: &ast.BinaryExpression{
							Operator: ">",
							Left: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "bar"},
								Property: &ast.Identifier{Name: "Close"},
							},
							Right: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "bar"},
								Property: &ast.Identifier{Name: "Open"},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	expectedDeclarations := []string{
		"var closeSeries *series.Series",
		"var highSeries *series.Series",
		"var lowSeries *series.Series",
		"var openSeries *series.Series",
		"var volumeSeries *series.Series",
	}

	for _, expected := range expectedDeclarations {
		if !strings.Contains(code.FunctionBody, expected) {
			t.Errorf("Expected declaration %q not found in generated code", expected)
		}
	}
}

/* TestBarFieldSeriesCodegen_Initialization tests Series initialization before bar loop */
func TestBarFieldSeriesCodegen_Initialization(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "signal"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	expectedInits := []string{
		"closeSeries = series.NewSeries(len(ctx.Data))",
		"highSeries = series.NewSeries(len(ctx.Data))",
		"lowSeries = series.NewSeries(len(ctx.Data))",
		"openSeries = series.NewSeries(len(ctx.Data))",
		"volumeSeries = series.NewSeries(len(ctx.Data))",
	}

	for _, expected := range expectedInits {
		if !strings.Contains(code.FunctionBody, expected) {
			t.Errorf("Expected initialization %q not found in generated code", expected)
		}
	}
}

/* TestBarFieldSeriesCodegen_Population tests Series.Set() in bar loop */
func TestBarFieldSeriesCodegen_Population(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "dummy"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	expectedPopulations := []string{
		"closeSeries.Set(bar.Close)",
		"highSeries.Set(bar.High)",
		"lowSeries.Set(bar.Low)",
		"openSeries.Set(bar.Open)",
		"volumeSeries.Set(bar.Volume)",
	}

	for _, expected := range expectedPopulations {
		if !strings.Contains(code.FunctionBody, expected) {
			t.Errorf("Expected population %q not found in generated code", expected)
		}
	}
}

/* TestBarFieldSeriesCodegen_CursorAdvancement tests Series.Next() calls */
func TestBarFieldSeriesCodegen_CursorAdvancement(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "value"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	expectedAdvancements := []string{
		"closeSeries.Next()",
		"highSeries.Next()",
		"lowSeries.Next()",
		"openSeries.Next()",
		"volumeSeries.Next()",
	}

	for _, expected := range expectedAdvancements {
		if !strings.Contains(code.FunctionBody, expected) {
			t.Errorf("Expected cursor advancement %q not found in generated code", expected)
		}
	}
}

/* TestBarFieldSeriesCodegen_OrderingLifecycle tests correct lifecycle ordering */
func TestBarFieldSeriesCodegen_OrderingLifecycle(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "test"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	body := code.FunctionBody

	// Find positions
	declPos := strings.Index(body, "var closeSeries *series.Series")
	initPos := strings.Index(body, "closeSeries = series.NewSeries")
	populatePos := strings.Index(body, "closeSeries.Set(bar.Close)")
	nextPos := strings.Index(body, "closeSeries.Next()")

	if declPos == -1 || initPos == -1 || populatePos == -1 || nextPos == -1 {
		t.Fatal("Missing expected bar field Series lifecycle statements")
	}

	// Verify ordering: declare → initialize → populate → advance
	if !(declPos < initPos && initPos < populatePos && populatePos < nextPos) {
		t.Errorf("Bar field Series lifecycle out of order: decl=%d, init=%d, populate=%d, next=%d",
			declPos, initPos, populatePos, nextPos)
	}
}

/* TestBarFieldSeriesCodegen_AlwaysGenerated tests bar fields exist regardless of variable usage */
func TestBarFieldSeriesCodegen_AlwaysGenerated(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
	}{
		{
			name: "Only strategy calls",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "strategy"},
								Property: &ast.Identifier{Name: "entry"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "long"},
							},
						},
					},
				},
			},
		},
		{
			name: "Bar field in conditional without variables",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.BinaryExpression{
							Operator: ">",
							Left: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "bar"},
								Property: &ast.Identifier{Name: "Close"},
							},
							Right: &ast.Literal{Value: 100.0},
						},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "entry"},
									},
									Arguments: []ast.Expression{
										&ast.Literal{Value: "long"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			requiredElements := []struct {
				pattern string
				reason  string
			}{
				{"var closeSeries *series.Series", "declaration"},
				{"closeSeries = series.NewSeries(len(ctx.Data))", "initialization"},
				{"closeSeries.Set(bar.Close)", "population in bar loop"},
				{"closeSeries.Next()", "cursor advancement"},
			}

			for _, elem := range requiredElements {
				if !strings.Contains(code.FunctionBody, elem.pattern) {
					t.Errorf("Missing bar field Series %s: %q", elem.reason, elem.pattern)
				}
			}
		})
	}
}

/* TestBarFieldSeriesCodegen_WithSingleVariable tests bar fields generated with any variable */
func TestBarFieldSeriesCodegen_WithSingleVariable(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   ast.Identifier{Name: "x"},
						Init: &ast.Literal{Value: 42.0},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Bar field Series should be generated once any variable exists
	if !strings.Contains(code.FunctionBody, "var closeSeries *series.Series") {
		t.Error("Program with variables should generate bar field Series declarations")
	}

	if !strings.Contains(code.FunctionBody, "closeSeries = series.NewSeries") {
		t.Error("Program with variables should initialize bar field Series")
	}
}

/* TestBarFieldSeriesCodegen_AllFieldsPresent tests all OHLCV fields always generated together */
func TestBarFieldSeriesCodegen_AllFieldsPresent(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "useClose"},
						Init: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "bar"},
							Property: &ast.Identifier{Name: "Close"},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// All OHLCV fields must be present even if only Close is used
	allFields := []string{"closeSeries", "highSeries", "lowSeries", "openSeries", "volumeSeries"}

	for _, field := range allFields {
		if !strings.Contains(code.FunctionBody, "var "+field+" *series.Series") {
			t.Errorf("Field %s declaration missing (all OHLCV fields should be generated)", field)
		}

		if !strings.Contains(code.FunctionBody, field+" = series.NewSeries") {
			t.Errorf("Field %s initialization missing", field)
		}

		if !strings.Contains(code.FunctionBody, field+".Next()") {
			t.Errorf("Field %s cursor advancement missing", field)
		}
	}
}

/* TestBarFieldSeriesInLookback_MultipleOccurrences tests bar fields in repeated lookback contexts */
func TestBarFieldSeriesInLookback_MultipleOccurrences(t *testing.T) {
	g := newTestGenerator()

	highAccess := g.convertSeriesAccessToOffset("bar.High", "lookbackOffset")
	if highAccess != "highSeries.Get(lookbackOffset)" {
		t.Errorf("bar.High conversion = %q, want %q", highAccess, "highSeries.Get(lookbackOffset)")
	}

	lowAccess := g.convertSeriesAccessToOffset("bar.Low", "lookbackOffset")
	if lowAccess != "lowSeries.Get(lookbackOffset)" {
		t.Errorf("bar.Low conversion = %q, want %q", lowAccess, "lowSeries.Get(lookbackOffset)")
	}

	arrayStyleHigh := "ctx.Data[i-lookbackOffset].High"
	arrayStyleLow := "ctx.Data[i-lookbackOffset].Low"

	if highAccess == arrayStyleHigh {
		t.Error("bar.High should use ForwardSeriesBuffer, not array paradigm")
	}

	if lowAccess == arrayStyleLow {
		t.Error("bar.Low should use ForwardSeriesBuffer, not array paradigm")
	}
}

/* TestBarFieldSeriesInLookback_MixedBarAndUserSeries tests bar fields alongside user variables */
func TestBarFieldSeriesInLookback_MixedBarAndUserSeries(t *testing.T) {
	g := newTestGenerator()

	userAccess := g.convertSeriesAccessToOffset("myValueSeries.GetCurrent()", "lookbackOffset")
	if userAccess != "myValueSeries.Get(lookbackOffset)" {
		t.Errorf("User variable conversion = %q, want %q", userAccess, "myValueSeries.Get(lookbackOffset)")
	}

	barAccess := g.convertSeriesAccessToOffset("bar.Close", "lookbackOffset")
	if barAccess != "closeSeries.Get(lookbackOffset)" {
		t.Errorf("Bar field conversion = %q, want %q", barAccess, "closeSeries.Get(lookbackOffset)")
	}

	if !strings.Contains(userAccess, ".Get(") || !strings.Contains(barAccess, ".Get(") {
		t.Error("ForwardSeriesBuffer paradigm requires .Get() for both user and bar field Series")
	}

	if strings.Contains(barAccess, "ctx.Data[i-") {
		t.Error("Bar field should not use array paradigm (ForwardSeriesBuffer consistency violated)")
	}
}

/* TestBarFieldSeriesInLookback_OffsetVariableNames tests different offset variable names */
func TestBarFieldSeriesInLookback_OffsetVariableNames(t *testing.T) {
	g := newTestGenerator()

	tests := []struct {
		barField   string
		offsetVar  string
		wantSeries string
	}{
		{"bar.Close", "i", "closeSeries.Get(i)"},
		{"bar.High", "offset", "highSeries.Get(offset)"},
		{"bar.Low", "lookback", "lowSeries.Get(lookback)"},
		{"bar.Open", "n", "openSeries.Get(n)"},
		{"bar.Volume", "idx", "volumeSeries.Get(idx)"},
	}

	for _, tt := range tests {
		t.Run(tt.barField+"_"+tt.offsetVar, func(t *testing.T) {
			got := g.convertSeriesAccessToOffset(tt.barField, tt.offsetVar)
			if got != tt.wantSeries {
				t.Errorf("convertSeriesAccessToOffset(%q, %q) = %q, want %q",
					tt.barField, tt.offsetVar, got, tt.wantSeries)
			}
		})
	}
}

/* TestBarFieldSeries_EdgeCases tests boundary conditions and error cases */
func TestBarFieldSeries_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
		check   func(*testing.T, string)
	}{
		{
			name: "Nested bar field access in complex expression",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "signal"},
								Init: &ast.BinaryExpression{
									Operator: "&&",
									Left: &ast.BinaryExpression{
										Operator: ">",
										Left: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Close"},
										},
										Right: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Open"},
										},
									},
									Right: &ast.BinaryExpression{
										Operator: ">",
										Left: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Volume"},
										},
										Right: &ast.Literal{Value: 1000.0},
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, code string) {
				if !strings.Contains(code, "closeSeries.Set(bar.Close)") {
					t.Error("Bar field Series should be populated for Close")
				}
				if !strings.Contains(code, "volumeSeries.Set(bar.Volume)") {
					t.Error("Bar field Series should be populated for Volume")
				}
			},
		},
		{
			name: "Multiple bar fields in same statement",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "range"},
								Init: &ast.BinaryExpression{
									Operator: "-",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "bar"},
										Property: &ast.Identifier{Name: "High"},
									},
									Right: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "bar"},
										Property: &ast.Identifier{Name: "Low"},
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, code string) {
				allBarFields := []string{"closeSeries", "highSeries", "lowSeries", "openSeries", "volumeSeries"}
				for _, field := range allBarFields {
					if !strings.Contains(code, "var "+field+" *series.Series") {
						t.Errorf("All bar fields should be declared, missing: %s", field)
					}
				}
			},
		},
		{
			name: "Bar fields with user variables",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   ast.Identifier{Name: "myVar"},
								Init: &ast.Literal{Value: 1.0},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "comparison"},
								Init: &ast.BinaryExpression{
									Operator: ">",
									Left: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "bar"},
										Property: &ast.Identifier{Name: "Close"},
									},
									Right: &ast.Identifier{Name: "myVar"},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, code string) {
				if !strings.Contains(code, "var myVarSeries *series.Series") {
					t.Error("User variable Series should be declared")
				}
				if !strings.Contains(code, "var closeSeries *series.Series") {
					t.Error("Bar field Series should be declared")
				}
				declPos := strings.Index(code, "var closeSeries")
				userDeclPos := strings.Index(code, "var myVarSeries")
				if declPos == -1 || userDeclPos == -1 {
					t.Fatal("Missing expected declarations")
				}
				if declPos > userDeclPos {
					t.Error("Bar field Series should be declared before user variable Series")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			tt.check(t, code.FunctionBody)
		})
	}
}

/* TestBarFieldSeries_Integration tests bar fields work with complete strategy patterns */
func TestBarFieldSeries_Integration(t *testing.T) {
	tests := []struct {
		name        string
		program     *ast.Program
		wantContain []string
	}{
		{
			name: "Bar fields with TA indicators",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "sma20"},
								Init: &ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Close"},
										},
										&ast.Literal{Value: 20.0},
									},
								},
							},
						},
					},
				},
			},
			wantContain: []string{
				"var closeSeries *series.Series",
				"var sma20Series *series.Series",
				"closeSeries.Set(bar.Close)",
				"sma20Series.Set(",
				"closeSeries.Next()",
				"sma20Series.Next()",
			},
		},
		{
			name: "Bar fields with conditional logic",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: ast.Identifier{Name: "signal"},
								Init: &ast.BinaryExpression{
									Operator: "&&",
									Left: &ast.BinaryExpression{
										Operator: ">",
										Left: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Close"},
										},
										Right: &ast.Literal{Value: 100.0},
									},
									Right: &ast.BinaryExpression{
										Operator: ">",
										Left: &ast.MemberExpression{
											Object:   &ast.Identifier{Name: "bar"},
											Property: &ast.Identifier{Name: "Volume"},
										},
										Right: &ast.Literal{Value: 1000.0},
									},
								},
							},
						},
					},
				},
			},
			wantContain: []string{
				"var closeSeries *series.Series",
				"var volumeSeries *series.Series",
				"closeSeries = series.NewSeries(len(ctx.Data))",
				"volumeSeries = series.NewSeries(len(ctx.Data))",
				"closeSeries.Set(bar.Close)",
				"volumeSeries.Set(bar.Volume)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(code.FunctionBody, want) {
					t.Errorf("Expected pattern %q not found in generated code", want)
				}
			}
		})
	}
}

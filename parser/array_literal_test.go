package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrayLiteral_EmptyArray(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
arr = []
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	literal, ok := varDecl.Declarations[0].Init.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected Literal for empty array, got %T", varDecl.Declarations[0].Init)
	}

	elements, ok := literal.Value.([]ast.Expression)
	if !ok {
		t.Fatalf("Expected []ast.Expression, got %T", literal.Value)
	}

	if len(elements) != 0 {
		t.Errorf("Expected empty array, got %d elements", len(elements))
	}
}

func TestArrayLiteral_NumericElements(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []interface{}
	}{
		{
			name:     "integers",
			code:     "arr = [1, 2, 3]",
			expected: []interface{}{1.0, 2.0, 3.0},
		},
		{
			name:     "floats",
			code:     "arr = [1.5, 2.7, 3.14]",
			expected: []interface{}{1.5, 2.7, 3.14},
		},
		{
			name:     "mixed int and float",
			code:     "arr = [1, 2.5, 3]",
			expected: []interface{}{1.0, 2.5, 3.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[1].(*ast.VariableDeclaration)
			literal, ok := varDecl.Declarations[0].Init.(*ast.Literal)
			if !ok {
				t.Fatalf("Expected Literal, got %T", varDecl.Declarations[0].Init)
			}

			elements, ok := literal.Value.([]ast.Expression)
			if !ok {
				t.Fatalf("Expected []ast.Expression, got %T", literal.Value)
			}

			if len(elements) != len(tt.expected) {
				t.Fatalf("Expected %d elements, got %d", len(tt.expected), len(elements))
			}

			for i, expectedVal := range tt.expected {
				elemLiteral, ok := elements[i].(*ast.Literal)
				if !ok {
					t.Errorf("Element[%d]: expected Literal, got %T", i, elements[i])
					continue
				}

				if elemLiteral.Value != expectedVal {
					t.Errorf("Element[%d]: expected %v, got %v", i, expectedVal, elemLiteral.Value)
				}
			}
		})
	}
}

func TestArrayLiteral_StringElements(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
arr = ["a", "b", "c"]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	literal := varDecl.Declarations[0].Init.(*ast.Literal)
	elements := literal.Value.([]ast.Expression)

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		elemLiteral := elements[i].(*ast.Literal)
		if elemLiteral.Value != exp {
			t.Errorf("Element[%d]: expected %q, got %v", i, exp, elemLiteral.Value)
		}
	}
}

func TestArrayLiteral_BooleanElements(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
arr = [true, false, true]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	literal := varDecl.Declarations[0].Init.(*ast.Literal)
	elements := literal.Value.([]ast.Expression)

	expected := []bool{true, false, true}
	for i, exp := range expected {
		elemLiteral := elements[i].(*ast.Literal)
		if elemLiteral.Value != exp {
			t.Errorf("Element[%d]: expected %v, got %v", i, exp, elemLiteral.Value)
		}
	}
}

func TestArrayLiteral_IdentifierElements(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
arr = [close, open, high, low]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	literal := varDecl.Declarations[0].Init.(*ast.Literal)
	elements := literal.Value.([]ast.Expression)

	expected := []string{"close", "open", "high", "low"}
	for i, exp := range expected {
		ident, ok := elements[i].(*ast.Identifier)
		if !ok {
			t.Errorf("Element[%d]: expected Identifier, got %T", i, elements[i])
			continue
		}

		if ident.Name != exp {
			t.Errorf("Element[%d]: expected identifier %q, got %q", i, exp, ident.Name)
		}
	}
}

func TestArrayLiteral_NestedArrays(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedDepth int
		expectedSizes []int
	}{
		{
			name:          "2D array",
			code:          "arr = [[1,2], [3,4]]",
			expectedDepth: 2,
			expectedSizes: []int{2, 2},
		},
		{
			name:          "3D array",
			code:          "arr = [[[1,2],[3,4]], [[5,6],[7,8]]]",
			expectedDepth: 3,
			expectedSizes: []int{2, 2},
		},
		{
			name:          "irregular nested",
			code:          "arr = [[1], [2,3], [4,5,6]]",
			expectedDepth: 2,
			expectedSizes: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[1].(*ast.VariableDeclaration)
			literal := varDecl.Declarations[0].Init.(*ast.Literal)
			elements := literal.Value.([]ast.Expression)

			if len(elements) != len(tt.expectedSizes) {
				t.Errorf("Expected %d top-level elements, got %d", len(tt.expectedSizes), len(elements))
			}
		})
	}
}

func TestArrayLiteral_InFunctionArguments(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "single array arg",
			code: "result = someFunc([1,2,3])",
		},
		{
			name: "multiple array args",
			code: "result = someFunc([1,2], [3,4])",
		},
		{
			name: "mixed args",
			code: `result = someFunc([1,2], "str", 42, [3,4])`,
		},
		{
			name: "named array arg",
			code: "result = someFunc(data=[1,2,3])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestArrayLiteral_InTernaryExpression(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "both branches arrays",
			code: "result = condition ? [1,2] : [3,4]",
		},
		{
			name: "true branch array",
			code: "result = condition ? [1,2,3] : 0",
		},
		{
			name: "false branch array",
			code: "result = condition ? 0 : [1,2,3]",
		},
		{
			name: "nested ternary with arrays",
			code: "result = cond1 ? (cond2 ? [1,2] : [3,4]) : [5,6]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestArrayLiteral_WithSubscript(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "direct subscript",
			code: "val = ([1,2,3])[0]",
		},
		{
			name: "subscript with expression",
			code: "val = ([10,20,30])[i]",
		},
		{
			name: "nested array subscript",
			code: "val = ([[1,2],[3,4]])[0][1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			varDecl := program.Body[1].(*ast.VariableDeclaration)
			memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("Expected MemberExpression for subscript, got %T", varDecl.Declarations[0].Init)
			}

			if !memberExpr.Computed {
				t.Error("Expected computed property (subscript)")
			}
		})
	}
}

func TestArrayLiteral_ParenthesizedWithSubscript(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "simple parenthesized subscript",
			code: "val = ([1,2,3])[0]",
		},
		{
			name: "nested parenthesized",
			code: "val = ([[1,2],[3,4]])[0][1]",
		},
		{
			name: "in function argument",
			code: "result = func(([1,2,3])[0])",
		},
		{
			name: "multiple in args",
			code: "result = func(([1,2])[0], ([3,4])[1])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
			}
		})
	}
}

func TestArrayLiteral_InArithmeticContext(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{
			name:      "parenthesized array subscript in arithmetic",
			code:      "val = ([1,2,3])[0] + ([4,5,6])[1]",
			shouldErr: false,
		},
		{
			name:      "array element in expression",
			code:      "val = [10,20,30][0] * 2",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if tt.shouldErr && err == nil {
				t.Fatal("Expected parse error, got none")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("Parse failed unexpectedly: %v", err)
			}

			if !tt.shouldErr {
				converter := NewConverter()
				_, err := converter.ToESTree(script)
				if err != nil {
					t.Fatalf("Conversion failed: %v", err)
				}
			}
		})
	}
}

func TestArrayLiteral_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{
			name:      "single element",
			code:      "arr = [42]",
			shouldErr: false,
		},
		{
			name:      "trailing comma",
			code:      "arr = [1,2,3,]",
			shouldErr: true,
		},
		{
			name:      "mixed literal types",
			code:      `arr = [1, "str", true, 3.14]`,
			shouldErr: false,
		},
		{
			name:      "very long array",
			code:      "arr = [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pineScript := "//@version=5\nindicator(\"Test\")\n" + tt.code + "\n"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(pineScript))
			if tt.shouldErr && err == nil {
				t.Fatal("Expected parse error, got none")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("Parse failed unexpectedly: %v", err)
			}

			if !tt.shouldErr {
				converter := NewConverter()
				_, err := converter.ToESTree(script)
				if err != nil {
					t.Fatalf("Conversion failed: %v", err)
				}
			}
		})
	}
}

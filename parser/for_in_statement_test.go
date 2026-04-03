package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestForInStatement_BasicSyntax(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
		isTuple       bool
	}{
		{
			name: "single element form",
			source: `for val in myArray
    x = val`,
			expectedStmts: 1,
			isTuple:       false,
		},
		{
			name: "tuple destructuring form",
			source: `for [i, val] in myArray
    x = val`,
			expectedStmts: 1,
			isTuple:       true,
		},
		{
			name: "multiple body statements single form",
			source: `for val in myArray
    a = val
    b = val
    c = val`,
			expectedStmts: 3,
			isTuple:       false,
		},
		{
			name: "multiple body statements tuple form",
			source: `for [idx, elem] in prices
    sum := sum + elem
    count := count + 1`,
			expectedStmts: 2,
			isTuple:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != 1 {
				t.Fatalf("Expected 1 top-level statement, got %d", len(script.Statements))
			}

			forIn := script.Statements[0].Core.ForIn
			if forIn == nil {
				t.Fatal("Expected ForInStatement, got nil")
			}

			if forIn.Collection == nil {
				t.Fatal("Expected collection expression, got nil")
			}

			if len(forIn.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(forIn.Body))
			}

			if tt.isTuple {
				if forIn.Vars.TupleIndex == nil || forIn.Vars.TupleElement == nil {
					t.Error("Tuple form should have both index and element set")
				}
				if forIn.Vars.SingleElement != nil {
					t.Error("Tuple form should not have single element set")
				}
			} else {
				if forIn.Vars.SingleElement == nil {
					t.Error("Single form should have single element set")
				}
				if forIn.Vars.TupleIndex != nil || forIn.Vars.TupleElement != nil {
					t.Error("Single form should not have tuple fields set")
				}
			}
		})
	}
}

func TestForInStatement_VariableBindings(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectIndex   string
		expectElement string
	}{
		{
			name: "single short name",
			source: `for v in arr
    x = v`,
			expectElement: "v",
		},
		{
			name: "single descriptive name",
			source: `for currentPrice in historicalPrices
    x = currentPrice`,
			expectElement: "currentPrice",
		},
		{
			name: "single underscore-prefixed name",
			source: `for _val in items
    x = _val`,
			expectElement: "_val",
		},
		{
			name: "tuple short names",
			source: `for [i, v] in arr
    x = v`,
			expectIndex:   "i",
			expectElement: "v",
		},
		{
			name: "tuple descriptive names",
			source: `for [barIndex, closePrice] in closePrices
    x = closePrice`,
			expectIndex:   "barIndex",
			expectElement: "closePrice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			vars := script.Statements[0].Core.ForIn.Vars

			if tt.expectIndex != "" {
				if vars.TupleIndex == nil || *vars.TupleIndex != tt.expectIndex {
					got := "<nil>"
					if vars.TupleIndex != nil {
						got = *vars.TupleIndex
					}
					t.Errorf("Expected index var %q, got %s", tt.expectIndex, got)
				}
			}

			if vars.SingleElement != nil {
				if *vars.SingleElement != tt.expectElement {
					t.Errorf("Expected element var %q, got %q", tt.expectElement, *vars.SingleElement)
				}
			} else if vars.TupleElement != nil {
				if *vars.TupleElement != tt.expectElement {
					t.Errorf("Expected element var %q, got %q", tt.expectElement, *vars.TupleElement)
				}
			}
		})
	}
}

func TestForInStatement_NestedLoops(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		outerStmts  int
		innerType   string
		description string
	}{
		{
			name: "for-in nested in for-in",
			source: `for row in matrix
    for val in row
        x = val`,
			outerStmts:  1,
			innerType:   "forin",
			description: "for-in containing for-in",
		},
		{
			name: "for-in nested in traditional for",
			source: `for i = 0 to 5
    for val in myArray
        x = val`,
			outerStmts:  1,
			innerType:   "forin",
			description: "traditional for containing for-in",
		},
		{
			name: "traditional for nested in for-in",
			source: `for val in myArray
    for i = 0 to 5
        x = i`,
			outerStmts:  1,
			innerType:   "for",
			description: "for-in containing traditional for",
		},
		{
			name: "for-in with sibling statements",
			source: `for val in myArray
    a = val
    for elem in nested
        b = elem
    c = val`,
			outerStmts:  3,
			innerType:   "forin",
			description: "for-in body with mixed statements and nested for-in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			stmt := script.Statements[0]

			var bodyStmts []*Statement
			if stmt.Core.ForIn != nil {
				bodyStmts = stmt.Core.ForIn.Body
			} else if stmt.Core.For != nil {
				bodyStmts = stmt.Core.For.Body
			} else {
				t.Fatal("Expected for or for-in as outer statement")
			}

			if len(bodyStmts) != tt.outerStmts {
				t.Errorf("Expected %d outer body statements, got %d", tt.outerStmts, len(bodyStmts))
			}

			foundInner := false
			for _, s := range bodyStmts {
				switch tt.innerType {
				case "forin":
					if s.Core.ForIn != nil {
						foundInner = true
					}
				case "for":
					if s.Core.For != nil {
						foundInner = true
					}
				}
			}
			if !foundInner {
				t.Errorf("Expected inner %s loop but found none", tt.innerType)
			}
		})
	}
}

func TestForInStatement_WithVariableOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "local variable declaration",
			source: `for val in arr
    temp = val * 2`,
		},
		{
			name: "outer variable reassignment",
			source: `for val in arr
    total := total + val`,
		},
		{
			name: "multiple operations",
			source: `for val in arr
    temp = val * 2
    total := total + temp`,
		},
		{
			name: "tuple index used in body",
			source: `for [i, val] in arr
    weight = i * val`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			forIn := script.Statements[0].Core.ForIn
			if forIn == nil {
				t.Fatal("Expected ForInStatement, got nil")
			}

			if len(forIn.Body) == 0 {
				t.Error("Expected at least one statement in for-in body")
			}
		})
	}
}

func TestForInStatement_MultipleSequential(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedForIns int
	}{
		{
			name: "two sequential for-in",
			source: `for val in prices
    x = val
for elem in volumes
    y = elem`,
			expectedForIns: 2,
		},
		{
			name: "three sequential for-in",
			source: `for a in arr1
    x = a
for b in arr2
    y = b
for c in arr3
    z = c`,
			expectedForIns: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != tt.expectedForIns {
				t.Errorf("Expected %d for-in statements, got %d", tt.expectedForIns, len(script.Statements))
			}

			for i, stmt := range script.Statements {
				if stmt.Core.ForIn == nil {
					t.Errorf("Statement %d is not a ForInStatement", i)
				}
			}
		})
	}
}

func TestForInStatement_MixedWithOtherStatements(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedPattern []string
	}{
		{
			name: "assignment then for-in",
			source: `total = 0
for val in myArray
    total := total + val`,
			expectedPattern: []string{"assign", "forin"},
		},
		{
			name: "for-in then assignment",
			source: `for val in myArray
    temp = val
result = temp * 2`,
			expectedPattern: []string{"forin", "assign"},
		},
		{
			name: "interleaved with traditional for",
			source: `for val in myArray
    x = val
for i = 0 to 10
    y = i
for elem in other
    z = elem`,
			expectedPattern: []string{"forin", "for", "forin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != len(tt.expectedPattern) {
				t.Fatalf("Expected %d statements, got %d", len(tt.expectedPattern), len(script.Statements))
			}

			for i, expected := range tt.expectedPattern {
				stmt := script.Statements[i]
				switch expected {
				case "forin":
					if stmt.Core.ForIn == nil {
						t.Errorf("Statement %d: expected for-in, got other type", i)
					}
				case "for":
					if stmt.Core.For == nil {
						t.Errorf("Statement %d: expected traditional for, got other type", i)
					}
				case "assign":
					if stmt.Core.Assignment == nil && stmt.Core.Reassignment == nil {
						t.Errorf("Statement %d: expected assignment, got other type", i)
					}
				}
			}
		})
	}
}

func TestForInStatement_Converter(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectIndex bool
		indexVar    string
		elementVar  string
		bodyCount   int
	}{
		{
			name: "single element to AST",
			source: `for val in prices
    x = val`,
			expectIndex: false,
			elementVar:  "val",
			bodyCount:   1,
		},
		{
			name: "tuple to AST",
			source: `for [i, val] in prices
    x = val`,
			expectIndex: true,
			indexVar:    "i",
			elementVar:  "val",
			bodyCount:   1,
		},
		{
			name: "multiple body statements to AST",
			source: `for [idx, elem] in data
    temp = elem * 2
    total := total + temp`,
			expectIndex: true,
			indexVar:    "idx",
			elementVar:  "elem",
			bodyCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one AST node")
			}

			forIn, ok := program.Body[0].(*ast.ForInStatement)
			if !ok {
				t.Fatalf("Expected *ast.ForInStatement, got %T", program.Body[0])
			}

			if forIn.ElementVar != tt.elementVar {
				t.Errorf("Expected element var %q, got %q", tt.elementVar, forIn.ElementVar)
			}

			if tt.expectIndex {
				if forIn.IndexVar != tt.indexVar {
					t.Errorf("Expected index var %q, got %q", tt.indexVar, forIn.IndexVar)
				}
			} else {
				if forIn.IndexVar != "" {
					t.Errorf("Expected empty index var, got %q", forIn.IndexVar)
				}
			}

			if forIn.Collection == nil {
				t.Fatal("Expected collection expression, got nil")
			}

			if len(forIn.Body) != tt.bodyCount {
				t.Errorf("Expected %d body nodes, got %d", tt.bodyCount, len(forIn.Body))
			}
		})
	}
}

func TestForInStatement_EmptyLinesAndComments(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedStmts int
	}{
		{
			name: "comment in body",
			source: `for val in arr
    // accumulate
    x = val`,
			expectedStmts: 1,
		},
		{
			name: "empty line in body",
			source: `for val in arr
    x = val

    y = val`,
			expectedStmts: 2,
		},
		{
			name: "multiple empty lines",
			source: `for val in arr
    x = val


    y = val`,
			expectedStmts: 2,
		},
		{
			name: "comment before for-in",
			source: `// iterate collection
for val in arr
    x = val`,
			expectedStmts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			var forIn *ForInStatement
			for _, stmt := range script.Statements {
				if stmt.Core.ForIn != nil {
					forIn = stmt.Core.ForIn
					break
				}
			}
			if forIn == nil {
				t.Fatal("Expected ForInStatement")
			}

			if len(forIn.Body) != tt.expectedStmts {
				t.Errorf("Expected %d body statements, got %d", tt.expectedStmts, len(forIn.Body))
			}
		})
	}
}

package security

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/runtime/context"
)

func TestPrefetcher_WithMockFetcher(t *testing.T) {
	/* Test complete prefetch workflow with mock fetcher */
	mockFetcher := &mockDataFetcher{}
	prefetcher := NewSecurityPrefetcher(mockFetcher)

	/* Create mock program with security() call - matches actual AST structure */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID: ast.Identifier{
							NodeType: ast.TypeIdentifier,
							Name:     "dailyClose",
						},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object: &ast.Identifier{
									NodeType: ast.TypeIdentifier,
									Name:     "request",
								},
								Property: &ast.Identifier{
									NodeType: ast.TypeIdentifier,
									Name:     "security",
								},
							},
							Arguments: []ast.Expression{
								&ast.Literal{
									NodeType: ast.TypeLiteral,
									Value:    "TEST",
								},
								&ast.Literal{
									NodeType: ast.TypeLiteral,
									Value:    "1D",
								},
								&ast.Identifier{
									NodeType: ast.TypeIdentifier,
									Name:     "close",
								},
							},
						},
					},
				},
			},
		},
	}

	err := prefetcher.Prefetch(program, 5)
	if err != nil {
		t.Fatalf("Prefetch failed: %v", err)
	}

	cache := prefetcher.GetCache()
	entry, found := cache.Get("TEST", "1D")
	if !found {
		t.Fatal("Expected TEST:1D entry in cache")
	}

	if len(entry.Context.Data) != 5 {
		t.Errorf("Expected 5 bars from mock, got %d", len(entry.Context.Data))
	}

	/* Verify context data (synthetic values 102-106) */
	expected := []float64{102, 103, 104, 105, 106}
	for i, exp := range expected {
		if entry.Context.Data[i].Close != exp {
			t.Errorf("Close[%d]: expected %.0f, got %.0f", i, exp, entry.Context.Data[i].Close)
		}
	}
}

func TestPrefetcher_NoSecurityCalls(t *testing.T) {
	mockFetcher := &mockDataFetcher{}
	prefetcher := NewSecurityPrefetcher(mockFetcher)

	/* Program without security() calls */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body:     []ast.Node{},
	}

	err := prefetcher.Prefetch(program, 100)
	if err != nil {
		t.Fatalf("Prefetch failed: %v", err)
	}

	cache := prefetcher.GetCache()
	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got %d entries", cache.Size())
	}
}

func TestPrefetcher_Deduplication(t *testing.T) {
	/* Test that multiple security() calls to same symbol+timeframe are deduplicated */
	mockFetcher := &mockDataFetcher{}
	prefetcher := NewSecurityPrefetcher(mockFetcher)

	/* Create program with 2 security() calls to TEST:1D */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			createSecurityDeclaration("sma", "TEST", "1D", createTACall("ta", "sma", "close", 20.0)),
			createSecurityDeclaration("ema", "TEST", "1D", createTACall("ta", "ema", "close", 10.0)),
		},
	}

	err := prefetcher.Prefetch(program, 50)
	if err != nil {
		t.Fatalf("Prefetch failed: %v", err)
	}

	cache := prefetcher.GetCache()

	/* Should only have 1 cache entry (deduplicated) */
	if cache.Size() != 1 {
		t.Errorf("Expected 1 cache entry (deduplicated), got %d", cache.Size())
	}

	/* Verify context exists */
	ctx, err := cache.GetContext("TEST", "1D")
	if err != nil {
		t.Errorf("Expected context cached: %v", err)
	}

	if len(ctx.Data) == 0 {
		t.Error("Expected context to have data bars")
	}
}

/* Helper: create VariableDeclaration with request.security() call */
func createSecurityDeclaration(varName, symbol, timeframe string, expr ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		NodeType: ast.TypeVariableDeclaration,
		Kind:     "var",
		Declarations: []ast.VariableDeclarator{
			{
				NodeType: ast.TypeVariableDeclarator,
				ID: ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     varName,
				},
				Init: &ast.CallExpression{
					NodeType: ast.TypeCallExpression,
					Callee: &ast.MemberExpression{
						NodeType: ast.TypeMemberExpression,
						Object: &ast.Identifier{
							NodeType: ast.TypeIdentifier,
							Name:     "request",
						},
						Property: &ast.Identifier{
							NodeType: ast.TypeIdentifier,
							Name:     "security",
						},
					},
					Arguments: []ast.Expression{
						&ast.Literal{NodeType: ast.TypeLiteral, Value: symbol},
						&ast.Literal{NodeType: ast.TypeLiteral, Value: timeframe},
						expr,
					},
				},
			},
		},
	}
}

/* Helper: create ta.function(source, period) call expression */
func createTACall(taObj, taFunc, source string, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     taObj,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     taFunc,
			},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: source},
			&ast.Literal{NodeType: ast.TypeLiteral, Value: period},
		},
	}
}

/* mockDataFetcher returns synthetic test data */
type mockDataFetcher struct{}

func (m *mockDataFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
	data := make([]context.OHLCV, limit)
	for i := 0; i < limit; i++ {
		data[i] = context.OHLCV{
			Time:   int64(1700000000 + i*86400),
			Open:   100.0 + float64(i),
			High:   105.0 + float64(i),
			Low:    95.0 + float64(i),
			Close:  102.0 + float64(i),
			Volume: 1000.0 + float64(i*10),
		}
	}
	return data, nil
}

package security

import (
	"fmt"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/* BenchmarkEvaluateIdentifier measures array allocation cost */
func BenchmarkEvaluateIdentifier(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("bars_%d", size), func(b *testing.B) {
			/* Setup context with N bars */
			secCtx := context.New("BTCUSDT", "1D", size)
			for i := 0; i < size; i++ {
				secCtx.AddBar(context.OHLCV{
					Time:   int64(i * 86400),
					Open:   100.0 + float64(i),
					High:   105.0 + float64(i),
					Low:    95.0 + float64(i),
					Close:  100.0 + float64(i),
					Volume: 1000.0,
				})
			}

			id := &ast.Identifier{Name: "close"}

			b.ResetTimer()
			b.ReportAllocs()

			/* Measure: evaluateIdentifier allocates []float64 every call */
			for i := 0; i < b.N; i++ {
				_, err := evaluateIdentifier(id, secCtx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

/* BenchmarkTASma measures TA function allocation overhead */
func BenchmarkTASma(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("bars_%d", size), func(b *testing.B) {
			/* Setup context */
			secCtx := context.New("BTCUSDT", "1D", size)
			for i := 0; i < size; i++ {
				secCtx.AddBar(context.OHLCV{
					Time:  int64(i * 86400),
					Close: 100.0 + float64(i%10),
				})
			}

			/* Parse ta.sma(close, 20) */
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20},
				},
			}

			b.ResetTimer()
			b.ReportAllocs()

			/* Measure: evaluateTASma calls ta.Sma() which allocates result array */
			for i := 0; i < b.N; i++ {
				_, err := evaluateCallExpression(call, secCtx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

/* BenchmarkPrefetchWorkflow measures full prefetch overhead */
func BenchmarkPrefetchWorkflow(b *testing.B) {
	/* Create AST with 3 security() calls */
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: ast.Identifier{Name: "dailyMA"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "request"},
								Property: &ast.Identifier{Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "BTCUSDT"},
								&ast.Literal{Value: "1D"},
								&ast.CallExpression{
									Callee: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "ta"},
										Property: &ast.Identifier{Name: "sma"},
									},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "close"},
										&ast.Literal{Value: 20},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	/* Mock fetcher */
	fetcher := &mockFetcher{barCount: 1000}

	b.ResetTimer()
	b.ReportAllocs()

	/* Measure: full prefetch allocates contexts + evaluates expressions */
	for i := 0; i < b.N; i++ {
		prefetcher := NewSecurityPrefetcher(fetcher)
		err := prefetcher.Prefetch(program, 500)
		if err != nil {
			b.Fatal(err)
		}
	}
}

/* mockFetcher generates bars without file I/O */
type mockFetcher struct {
	barCount int
}

func (m *mockFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
	count := m.barCount
	if limit > 0 && limit < count {
		count = limit
	}

	bars := make([]context.OHLCV, count)
	for i := 0; i < count; i++ {
		bars[i] = context.OHLCV{
			Time:   int64(i * 86400),
			Open:   100.0 + float64(i%10),
			High:   105.0 + float64(i%10),
			Low:    95.0 + float64(i%10),
			Close:  100.0 + float64(i%10),
			Volume: 1000.0,
		}
	}

	return bars, nil
}

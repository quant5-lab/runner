package security

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/parser"
)

func TestAnalyzeAST_SimpleSecurityCall(t *testing.T) {
	code := `
indicator("Test")
ma20 = request.security(syminfo.tickerid, '1D', close)
`
	program := parseCode(t, code)
	calls := AnalyzeAST(program)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 security call, got %d", len(calls))
	}

	call := calls[0]
	if call.Symbol != "syminfo.tickerid" {
		t.Errorf("Expected symbol 'syminfo.tickerid', got '%s'", call.Symbol)
	}
	if call.Timeframe != "1D" {
		t.Errorf("Expected timeframe '1D', got '%s'", call.Timeframe)
	}
	if call.Expression == nil {
		t.Error("Expected non-nil expression")
	}
}

func TestAnalyzeAST_MultipleSecurityCalls(t *testing.T) {
	code := `
indicator("Test")
daily_close = request.security("BTCUSDT", "1D", close)
hourly_high = security("ETHUSDT", "1h", high)
weekly_vol = request.security("BNBUSDT", "1W", volume)
`
	program := parseCode(t, code)
	calls := AnalyzeAST(program)

	if len(calls) != 3 {
		t.Fatalf("Expected 3 security calls, got %d", len(calls))
	}

	expected := []struct {
		symbol    string
		timeframe string
	}{
		{"BTCUSDT", "1D"},
		{"ETHUSDT", "1h"},
		{"BNBUSDT", "1W"},
	}

	for i, exp := range expected {
		if calls[i].Symbol != exp.symbol {
			t.Errorf("Call %d: expected symbol '%s', got '%s'", i, exp.symbol, calls[i].Symbol)
		}
		if calls[i].Timeframe != exp.timeframe {
			t.Errorf("Call %d: expected timeframe '%s', got '%s'", i, exp.timeframe, calls[i].Timeframe)
		}
	}
}

func TestAnalyzeAST_NestedFunctionExpression(t *testing.T) {
	code := `
indicator("Test")
daily_sma = request.security(syminfo.tickerid, '1D', ta.sma(close, 20))
`
	program := parseCode(t, code)
	calls := AnalyzeAST(program)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 security call, got %d", len(calls))
	}

	/* Expression should be CallExpression for ta.sma() */
	_, ok := calls[0].Expression.(*ast.CallExpression)
	if !ok {
		t.Errorf("Expected expression to be CallExpression, got %T", calls[0].Expression)
	}
}

func TestAnalyzeAST_NoSecurityCalls(t *testing.T) {
	code := `
indicator("Test")
sma20 = ta.sma(close, 20)
plot(sma20)
`
	program := parseCode(t, code)
	calls := AnalyzeAST(program)

	if len(calls) != 0 {
		t.Errorf("Expected 0 security calls, got %d", len(calls))
	}
}

func TestAnalyzeAST_SecurityWithInsufficientArgs(t *testing.T) {
	code := `
indicator("Test")
val = request.security("BTC")
`
	program := parseCode(t, code)
	calls := AnalyzeAST(program)

	/* Should not detect calls with insufficient arguments */
	if len(calls) != 0 {
		t.Errorf("Expected 0 security calls for invalid args, got %d", len(calls))
	}
}

/* Helper: parse code into AST */
func parseCode(t *testing.T, code string) *ast.Program {
	t.Helper()

	/* Create parser */
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	/* Parse to participle AST */
	script, err := p.ParseString("", code)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	/* Convert to ESTree AST */
	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	return program
}

/* TestExtractMaxPeriod_SimpleSMA tests basic SMA period extraction */
func TestExtractMaxPeriod_SimpleSMA(t *testing.T) {
	/* ta.sma(close, 20) */
	expr := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "ta",
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "sma",
			},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			&ast.Literal{NodeType: ast.TypeLiteral, Value: float64(20)},
		},
	}

	period := ExtractMaxPeriod(expr)
	if period != 20 {
		t.Errorf("Expected period 20, got %d", period)
	}
}

/* TestExtractMaxPeriod_SMA200 tests SMA200 extraction */
func TestExtractMaxPeriod_SMA200(t *testing.T) {
	/* ta.sma(close, 200) */
	expr := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "ta",
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "sma",
			},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			&ast.Literal{NodeType: ast.TypeLiteral, Value: float64(200)},
		},
	}

	period := ExtractMaxPeriod(expr)
	if period != 200 {
		t.Errorf("Expected period 200, got %d", period)
	}
}

/* TestExtractMaxPeriod_NestedTA tests nested TA functions */
func TestExtractMaxPeriod_NestedTA(t *testing.T) {
	/* ta.sma(ta.ema(close, 50), 200) → should return 200 (max of 50 and 200) */
	innerCall := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "ta",
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "ema",
			},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
			&ast.Literal{NodeType: ast.TypeLiteral, Value: float64(50)},
		},
	}

	outerCall := &ast.CallExpression{
		NodeType: ast.TypeCallExpression,
		Callee: &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "ta",
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "sma",
			},
		},
		Arguments: []ast.Expression{
			innerCall,
			&ast.Literal{NodeType: ast.TypeLiteral, Value: float64(200)},
		},
	}

	period := ExtractMaxPeriod(outerCall)
	if period != 200 {
		t.Errorf("Expected period 200 (max), got %d", period)
	}
}

/* TestExtractMaxPeriod_DirectClose tests direct close access (no TA) */
func TestExtractMaxPeriod_DirectClose(t *testing.T) {
	/* Just "close" identifier - no period needed */
	expr := &ast.Identifier{
		NodeType: ast.TypeIdentifier,
		Name:     "close",
	}

	period := ExtractMaxPeriod(expr)
	if period != 0 {
		t.Errorf("Expected period 0 for direct close, got %d", period)
	}
}

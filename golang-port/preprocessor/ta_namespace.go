package preprocessor

import (
	"github.com/borisquantlab/pinescript-go/parser"
)

// TANamespaceTransformer adds ta. prefix to technical analysis functions
// Examples: sma() → ta.sma(), ema() → ta.ema(), crossover() → ta.crossover()
type TANamespaceTransformer struct {
	mappings map[string]string
}

// NewTANamespaceTransformer creates a transformer with Pine v5 ta.* mappings
func NewTANamespaceTransformer() *TANamespaceTransformer {
	return &TANamespaceTransformer{
		mappings: map[string]string{
			// Moving averages
			"sma":    "ta.sma",
			"ema":    "ta.ema",
			"rma":    "ta.rma",
			"wma":    "ta.wma",
			"vwma":   "ta.vwma",
			"swma":   "ta.swma",
			"alma":   "ta.alma",
			"hma":    "ta.hma",
			"linreg": "ta.linreg",

			// Oscillators
			"rsi":   "ta.rsi",
			"macd":  "ta.macd",
			"stoch": "ta.stoch",
			"cci":   "ta.cci",
			"cmo":   "ta.cmo",
			"mfi":   "ta.mfi",
			"mom":   "ta.mom",
			"roc":   "ta.roc",
			"tsi":   "ta.tsi",
			"wpr":   "ta.wpr",

			// Bands & channels
			"bb":  "ta.bb",
			"bbw": "ta.bbw",
			"kc":  "ta.kc",
			"kcw": "ta.kcw",

			// Volatility
			"atr":      "ta.atr",
			"tr":       "ta.tr",
			"stdev":    "ta.stdev",
			"dev":      "ta.dev",
			"variance": "ta.variance",

			// Volume
			"obv":     "ta.obv",
			"pvt":     "ta.pvt",
			"nvi":     "ta.nvi",
			"pvi":     "ta.pvi",
			"wad":     "ta.wad",
			"wvad":    "ta.wvad",
			"accdist": "ta.accdist",
			"iii":     "ta.iii",

			// Trend
			"sar":        "ta.sar",
			"supertrend": "ta.supertrend",
			"dmi":        "ta.dmi",
			"cog":        "ta.cog",

			// Crossovers & comparisons
			"cross":      "ta.cross",
			"crossover":  "ta.crossover",
			"crossunder": "ta.crossunder",

			// Statistical
			"change":    "ta.change",
			"cum":       "ta.cum",
			"falling":   "ta.falling",
			"rising":    "ta.rising",
			"barsince":  "ta.barsince",
			"valuewhen": "ta.valuewhen",

			// High/Low
			"highest":     "ta.highest",
			"highestbars": "ta.highestbars",
			"lowest":      "ta.lowest",
			"lowestbars":  "ta.lowestbars",
			"pivothigh":   "ta.pivothigh",
			"pivotlow":    "ta.pivotlow",

			// Other
			"correlation":                     "ta.correlation",
			"median":                          "ta.median",
			"mode":                            "ta.mode",
			"percentile_linear_interpolation": "ta.percentile_linear_interpolation",
			"percentile_nearest_rank":         "ta.percentile_nearest_rank",
			"percentrank":                     "ta.percentrank",
			"range":                           "ta.range",
		},
	}
}

// Transform walks the AST and renames function calls
func (t *TANamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	for _, stmt := range script.Statements {
		t.transformStatement(stmt)
	}
	return script, nil
}

func (t *TANamespaceTransformer) transformStatement(stmt *parser.Statement) {
	if stmt.Assignment != nil {
		t.transformExpression(stmt.Assignment.Value)
	}
	if stmt.If != nil {
		t.transformComparison(stmt.If.Condition)
		if stmt.If.Body != nil {
			t.transformStatement(stmt.If.Body)
		}
	}
	if stmt.Expression != nil {
		t.transformExpression(stmt.Expression.Expr)
	}
}

func (t *TANamespaceTransformer) transformExpression(expr *parser.Expression) {
	if expr == nil {
		return
	}

	if expr.Ternary != nil {
		t.transformTernaryExpr(expr.Ternary)
	}
	if expr.Call != nil {
		t.transformCallExpr(expr.Call)
	}
	if expr.MemberAccess != nil {
		// Member accesses like bar_index don't need transformation
	}
}

func (t *TANamespaceTransformer) transformCallExpr(call *parser.CallExpr) {
	// Check if function name needs ta. prefix (only for simple identifiers)
	if call.Callee != nil && call.Callee.Ident != nil {
		if newName, ok := t.mappings[*call.Callee.Ident]; ok {
			call.Callee.Ident = &newName
		}
	}

	// Recursively transform arguments
	for _, arg := range call.Args {
		if arg.Value != nil {
			t.transformTernaryExpr(arg.Value)
		}
	}
}

func (t *TANamespaceTransformer) transformTernaryExpr(ternary *parser.TernaryExpr) {
	if ternary.Condition != nil {
		t.transformOrExpr(ternary.Condition)
	}
	if ternary.TrueVal != nil {
		t.transformExpression(ternary.TrueVal)
	}
	if ternary.FalseVal != nil {
		t.transformExpression(ternary.FalseVal)
	}
}

func (t *TANamespaceTransformer) transformOrExpr(or *parser.OrExpr) {
	if or.Left != nil {
		t.transformAndExpr(or.Left)
	}
	if or.Right != nil {
		t.transformOrExpr(or.Right)
	}
}

func (t *TANamespaceTransformer) transformAndExpr(and *parser.AndExpr) {
	if and.Left != nil {
		t.transformCompExpr(and.Left)
	}
	if and.Right != nil {
		t.transformAndExpr(and.Right)
	}
}

func (t *TANamespaceTransformer) transformCompExpr(comp *parser.CompExpr) {
	if comp.Left != nil {
		t.transformArithExpr(comp.Left)
	}
	if comp.Right != nil {
		t.transformCompExpr(comp.Right)
	}
}

func (t *TANamespaceTransformer) transformArithExpr(arith *parser.ArithExpr) {
	if arith.Left != nil {
		t.transformTerm(arith.Left)
	}
	if arith.Right != nil {
		t.transformArithExpr(arith.Right)
	}
}

func (t *TANamespaceTransformer) transformTerm(term *parser.Term) {
	if term.Left != nil {
		t.transformFactor(term.Left)
	}
	if term.Right != nil {
		t.transformTerm(term.Right)
	}
}

func (t *TANamespaceTransformer) transformFactor(factor *parser.Factor) {
	if factor.Call != nil {
		t.transformCallExpr(factor.Call)
	}
	if factor.MemberAccess != nil {
		// Member accesses don't need transformation
	}
	if factor.Subscript != nil {
		if factor.Subscript.Index != nil {
			t.transformArithExpr(factor.Subscript.Index)
		}
	}
}

func (t *TANamespaceTransformer) transformComparison(comp *parser.Comparison) {
	if comp.Left != nil {
		t.transformComparisonTerm(comp.Left)
	}
	if comp.Right != nil {
		t.transformComparisonTerm(comp.Right)
	}
}

func (t *TANamespaceTransformer) transformComparisonTerm(term *parser.ComparisonTerm) {
	if term.Call != nil {
		t.transformCallExpr(term.Call)
	}
	if term.Subscript != nil && term.Subscript.Index != nil {
		t.transformArithExpr(term.Subscript.Index)
	}
}

func (t *TANamespaceTransformer) transformValue(val *parser.Value) {
	if val == nil {
		return
	}

	if val.Subscript != nil && val.Subscript.Index != nil {
		t.transformArithExpr(val.Subscript.Index)
	}
}

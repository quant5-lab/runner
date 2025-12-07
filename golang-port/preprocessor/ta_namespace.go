package preprocessor

import (
	"github.com/quant5-lab/runner/parser"
)

/* Pine v4→v5: sma() → ta.sma(), ema() → ta.ema() */
type TANamespaceTransformer struct {
	base *NamespaceTransformer
}

// NewTANamespaceTransformer creates a transformer with Pine v5 ta.* mappings
func NewTANamespaceTransformer() *TANamespaceTransformer {
	mappings := map[string]string{
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
	}

	return &TANamespaceTransformer{
		base: NewNamespaceTransformer(mappings),
	}
}

// Transform walks the AST and renames function calls
func (t *TANamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	return t.base.Transform(script)
}

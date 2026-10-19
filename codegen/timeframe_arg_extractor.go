package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

const primaryTimeframeGoExpr = "ctx.Timeframe"

// TimeframeArgExtractor maps a Pine AST timeframe argument to the Go expression
// that evaluates to the timeframe string at runtime:
//   - string literal → Go quoted literal (e.g. "240")
//   - timeframe.period → ctx.Timeframe
//   - any other identifier → the identifier name (resolved at runtime)
//   - nil or unrecognised → ctx.Timeframe
type TimeframeArgExtractor struct {
	argParser *ArgumentParser
}

func NewTimeframeArgExtractor() *TimeframeArgExtractor {
	return &TimeframeArgExtractor{argParser: NewArgumentParser()}
}

// GoExpr returns the Go source expression for the given Pine timeframe argument.
func (e *TimeframeArgExtractor) GoExpr(arg ast.Expression) string {
	if arg == nil {
		return primaryTimeframeGoExpr
	}

	if s := e.argParser.ParseString(arg); s.IsValid {
		return fmt.Sprintf("%q", s.MustBeString())
	}

	if id := e.argParser.ParseIdentifier(arg); id.IsValid {
		if id.Identifier == "timeframe.period" {
			return primaryTimeframeGoExpr
		}
		return id.Identifier
	}

	return primaryTimeframeGoExpr
}

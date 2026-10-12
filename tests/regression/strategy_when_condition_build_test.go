package regression

import (
	"fmt"
	"testing"
)

/*
TestStrategyWhenConditionConstructsBuild guards the boolean-coercion decision by compiling the
generated Go, which is the only observable that reddens when a condition operand is wrapped in
value.IsTrue although it already is a Go bool.
*/
func TestStrategyWhenConditionConstructsBuild(t *testing.T) {
	root := projectRootFromCwd()

	constructs := []struct {
		name      string
		condition string
	}{
		{
			name:      "inline comparison",
			condition: `strategy.entry("L", strategy.long, when=close > 100)`,
		},
		{
			name:      "logical combination of comparisons",
			condition: `strategy.entry("L", strategy.long, when=close > 100 and close < 200)`,
		},
		{
			name:      "conditional expression",
			condition: `strategy.entry("L", strategy.long, when=close > 100 ? true : false)`,
		},
		{
			name:      "session namespace member",
			condition: `strategy.entry("L", strategy.long, when=session.isfirstbar)`,
		},
		{
			name:      "strategy runtime member",
			condition: `strategy.entry("L", strategy.long, when=strategy.position_size > 0)`,
		},
		{
			name:      "logical member in a variable declaration",
			condition: "signal = barstate.islast and close > 100\nstrategy.entry(\"L\", strategy.long, when=signal)",
		},
		{
			name:      "runtime member under unary minus",
			condition: `strategy.entry("L", strategy.long, when=-strategy.position_size < 0)`,
		},
		{
			name:      "numeric truthiness stays coerced",
			condition: `strategy.entry("L", strategy.long, when=-close)`,
		},
	}

	for _, tt := range constructs {
		t.Run(tt.name, func(t *testing.T) {
			source := fmt.Sprintf(`//@version=5
strategy("When condition", overlay=true)
%s
`, tt.condition)

			if _, ok := codegenAndBuild(t, t.TempDir(), "when_condition", source, root); !ok {
				t.Errorf("generated Go does not build for %s", tt.name)
			}
		})
	}
}

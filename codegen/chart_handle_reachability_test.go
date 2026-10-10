package codegen

import (
	"testing"
	"time"

	"github.com/quant5-lab/runner/ast"
)

func TestChartHandleReachability_CalculationBearingBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		body      []ast.Node
		reachable map[string]bool
	}{
		{
			name: "display-only constructor is silent",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("close"))},
			},
			reachable: map[string]bool{"line.new": false},
		},
		{
			name: "getter consumption marks constructor handle",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("price", chartCall("line", "get_price", identifier("line_handle"), identifier("bar_index"))),
				&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("price"))},
			},
			reachable: map[string]bool{"line.new": true},
		},
		{
			name: "alias chain preserves originating handle",
			body: []ast.Node{
				compatVarDecl("source_handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("alias_handle", identifier("source_handle")),
				compatVarDecl("top", chartCall("box", "get_top", identifier("alias_handle"))),
				&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("top"))},
			},
			reachable: map[string]bool{"box.new": true},
		},
		{
			name: "critical strategy argument marks handle",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(member("strategy", "entry"),
					literal("L"),
					member("strategy", "long"),
					objectArg("qty", identifier("line_handle")),
				)},
			},
			reachable: map[string]bool{"line.new": true},
		},
		{
			name: "presentation-only strategy comment leaves handle silent",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(member("strategy", "entry"),
					literal("L"),
					member("strategy", "long"),
					objectArg("comment", identifier("line_handle")),
				)},
			},
			reachable: map[string]bool{"line.new": false},
		},
		{
			name: "inline branch gate marks handle",
			body: []ast.Node{
				&ast.IfStatement{
					Test:       chartCall("line", "new", identifier("bar_index"), identifier("close")),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			reachable: map[string]bool{"line.new": true},
		},
		{
			name: "handle rebound by reassignment marks every producer",
			body: []ast.Node{
				compatVarDecl("handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
				&ast.IfStatement{
					Test:       identifier("handle"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			reachable: map[string]bool{"line.new": true, "box.new": true},
		},
		{
			name: "handle declared in both branches marks every producer",
			body: []ast.Node{
				&ast.IfStatement{
					Test: identifier("close"),
					Consequent: []ast.Node{
						compatVarDecl("handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
					},
					Alternate: []ast.Node{
						compatVarDecl("handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
					},
				},
				&ast.IfStatement{
					Test:       identifier("handle"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			reachable: map[string]bool{"line.new": true, "box.new": true},
		},
		{
			name: "rebound handle never consumed stays silent",
			body: []ast.Node{
				compatVarDecl("handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("close"))},
			},
			reachable: map[string]bool{"line.new": false, "box.new": false},
		},
		{
			name: "mutually aliased handles terminate and keep the shared producer",
			body: []ast.Node{
				compatVarDecl("first", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("second", identifier("first")),
				compatVarDecl("first", identifier("second")),
				&ast.IfStatement{
					Test:       identifier("first"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			reachable: map[string]bool{"line.new": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reachability := newChartHandleReachability(&ast.Program{Body: tt.body})
			for feature, want := range tt.reachable {
				call := firstChartHandleProducer(tt.body, feature)
				if call == nil {
					t.Fatalf("test fixture has no %s producer", feature)
				}
				if got := reachability.IsCalculationBearing(call); got != want {
					t.Fatalf("%s reachability = %v, want %v", feature, got, want)
				}
			}
		})
	}
}

func TestChartHandleReachability_NilProgramIsEmpty(t *testing.T) {
	if newChartHandleReachability(nil).IsCalculationBearing(nil) {
		t.Fatal("nil program reachability must not mark nil calls calculation-bearing")
	}
}

func TestChartHandleReachability_RebindingTerminates(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
		compatVarDecl("handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
	}}

	done := make(chan struct{})
	go func() {
		newChartHandleReachability(program)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("newChartHandleReachability did not terminate for a rebound handle")
	}
}

func firstChartHandleProducer(nodes []ast.Node, feature string) *ast.CallExpression {
	var found *ast.CallExpression
	walkCallExpressions(nodes, func(call *ast.CallExpression) {
		if found == nil && extractCallFunctionName(call) == feature {
			found = call
		}
	})
	return found
}

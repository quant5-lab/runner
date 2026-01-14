package main

import (
	"encoding/json"
	"fmt"
	"github.com/quant5-lab/runner/parser"
	"os"
)

func main() {
	src := `//@version=4
strategy("Test", overlay=true)
stop_level = 48000.0
smart_take_level = 58000.0
strategy.exit("BB exit", "BB entry", stop=stop_level, limit=smart_take_level)
`

	p, err := parser.NewParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewParser error: %v\n", err)
		os.Exit(1)
	}

	ast, err := p.ParseString("", src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Find strategy.exit call and dump arguments
	for _, stmt := range ast.Statements {
		if stmt.Expression != nil && stmt.Expression.Expr != nil {
			if stmt.Expression.Expr.Call != nil {
				call := stmt.Expression.Expr.Call
				if call.Callee != nil && call.Callee.MemberAccess != nil {
					sel := call.Callee.MemberAccess
					if sel.Object == "strategy" && len(sel.Properties) > 0 && sel.Properties[0] == "exit" {
						fmt.Println("Found strategy.exit call:")
						for i, arg := range call.Args {
							data, _ := json.MarshalIndent(arg, "", "  ")
							fmt.Printf("Arg[%d]: %s\n", i, string(data))
						}
					}
				}
			}
		}
	}
}

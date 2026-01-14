package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/quant5-lab/runner/parser"
)

func main() {
	src := `//@version=4
strategy("Test Exit Debug", overlay=true)
if (close > open)
    strategy.entry("Long", strategy.long)
strategy.exit("Exit", "Long", stop=48000, limit=58000)
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

	// Dump full AST structure
	data, _ := json.MarshalIndent(ast, "", "  ")
	fmt.Printf("Full AST:\n%s\n\n", string(data))

	// Find strategy.exit call
	for _, stmt := range ast.Statements {
		if stmt.Expression != nil && stmt.Expression.Expr != nil {
			if stmt.Expression.Expr.Call != nil {
				call := stmt.Expression.Expr.Call
				if call.Callee != nil && call.Callee.MemberAccess != nil {
					sel := call.Callee.MemberAccess
					if sel.Object == "strategy" && len(sel.Properties) > 0 && sel.Properties[0] == "exit" {
						fmt.Println("Found strategy.exit call:")
						fmt.Printf("Arguments count: %d\n", len(call.Args))
						for i, arg := range call.Args {
							data, _ := json.MarshalIndent(arg, "", "  ")
							argName := "positional"
							if arg.Name != nil {
								argName = *arg.Name
							}
							fmt.Printf("Arg[%d] name=%s type=%T: %s\n", i, argName, arg, string(data))
						}
					}
				}
			}
		}
	}
}

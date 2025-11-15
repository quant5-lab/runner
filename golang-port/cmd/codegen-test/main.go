package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/codegen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ast.json>\n", os.Args[0])
		os.Exit(1)
	}

	astBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read AST: %v\n", err)
		os.Exit(1)
	}

	var program ast.Program
	err = json.Unmarshal(astBytes, &program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse AST: %v\n", err)
		os.Exit(1)
	}

	code, err := codegen.GenerateStrategyCodeFromAST(&program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Codegen error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("// Generated Go code from Pine Script strategy")
	fmt.Println("// This demonstrates Series usage with historical value access")
	fmt.Println()
	fmt.Println("package main")
	fmt.Println()
	fmt.Println("import (")
	fmt.Println("\t\"github.com/borisquantlab/pinescript-go/runtime/context\"")
	fmt.Println("\t\"github.com/borisquantlab/pinescript-go/runtime/series\"")
	fmt.Println("\t\"github.com/borisquantlab/pinescript-go/runtime/strategy\"")
	fmt.Println(")")
	fmt.Println()
	fmt.Println("func executeStrategy(ctx *context.Context) (*strategy.Strategy) {")
	fmt.Println("\tstrat := strategy.NewStrategy()")
	fmt.Println()
	fmt.Print(code.FunctionBody)
	fmt.Println()
	fmt.Println("\treturn strat")
	fmt.Println("}")
}

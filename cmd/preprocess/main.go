package main

import (
	"flag"
	"fmt"
	"github.com/quant5-lab/runner/preprocessor"
	"os"
)

func main() {
	input := flag.String("input", "", "Input file")
	output := flag.String("output", "", "Output file")
	flag.Parse()

	content, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	processed := preprocessor.NormalizeIfBlocks(string(content))

	if err := os.WriteFile(*output, []byte(processed), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Preprocessed %s -> %s\n", *input, *output)
}

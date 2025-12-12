package parser

import (
	"encoding/json"
	"testing"
)

func TestParseSimpleIndicator(t *testing.T) {
	input := `//@version=5
indicator("Simple SMA", overlay=true)
sma20 = ta.sma(close, 20)
plot(sma20, color=color.blue, title="SMA20")
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(script.Statements) != 3 {
		t.Fatalf("Statements count = %d, want 3", len(script.Statements))
	}
}

func TestConvertToESTree(t *testing.T) {
	input := `//@version=5
indicator("Test", overlay=true)
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if program.NodeType != "Program" {
		t.Errorf("NodeType = %s, want Program", program.NodeType)
	}

	jsonBytes, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if result["type"] != "Program" {
		t.Errorf("JSON type = %s, want Program", result["type"])
	}
}

func TestParseBooleanLiterals(t *testing.T) {
	input := `indicator("Test", overlay=true)`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}

	if len(program.Body) == 0 {
		t.Fatal("Empty program body")
	}
}

func TestParseNamedArguments(t *testing.T) {
	input := `plot(close, color=color.blue, title="Test")`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	script, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(script.Statements) != 1 {
		t.Fatalf("Statements count = %d, want 1", len(script.Statements))
	}
}

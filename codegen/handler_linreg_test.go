package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestLinregHandler(t *testing.T) {
	handler := &LinregHandler{}

	t.Run("CanHandle ta.linreg", func(t *testing.T) {
		if !handler.CanHandle("ta.linreg") {
			t.Error("should handle ta.linreg")
		}
	})

	t.Run("CanHandle linreg", func(t *testing.T) {
		if !handler.CanHandle("linreg") {
			t.Error("should handle linreg")
		}
	})

	t.Run("Cannot handle other functions", func(t *testing.T) {
		if handler.CanHandle("ta.sma") {
			t.Error("should not handle ta.sma")
		}
		if handler.CanHandle("ta.ema") {
			t.Error("should not handle ta.ema")
		}
	})
}

func TestLinregHandlerInRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("ta.linreg is supported", func(t *testing.T) {
		if !registry.IsSupported("ta.linreg") {
			t.Error("ta.linreg should be supported by registry")
		}
	})

	t.Run("linreg is supported", func(t *testing.T) {
		if !registry.IsSupported("linreg") {
			t.Error("linreg should be supported by registry")
		}
	})

	t.Run("handler can be found", func(t *testing.T) {
		handler := registry.FindHandler("ta.linreg")
		if handler == nil {
			t.Fatal("handler for ta.linreg should be found")
		}
		if _, ok := handler.(*LinregHandler); !ok {
			t.Error("handler should be LinregHandler type")
		}
	})
}

func TestLinregHandlerCodeGeneration(t *testing.T) {
	handler := &LinregHandler{}

	// Create a minimal mock generator
	mockGen := &generator{}

	t.Run("missing arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{},
		}
		_, err := handler.GenerateCode(mockGen, "osc", call)
		if err == nil {
			t.Error("should error on missing arguments")
		}
	})

	t.Run("insufficient arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
		}
		_, err := handler.GenerateCode(mockGen, "osc", call)
		if err == nil {
			t.Error("should error on insufficient arguments (need at least 2)")
		}
	})
}

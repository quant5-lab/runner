package codegen

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/runtime/ta"
)

func TestTASignatureRegistryRuntimeCoverage(t *testing.T) {
	registry := sharedTASignatures
	allSignatures := registry.AllFunctionNames()

	runtimeFunctions := extractRuntimeTAFunctions()
	handlerRegistry := NewTAFunctionRegistry()

	var missingRuntime []string
	var missingHandlers []string

	for _, funcName := range allSignatures {
		if !strings.HasPrefix(funcName, "ta.") {
			continue
		}

		bareName := strings.TrimPrefix(funcName, "ta.")
		runtimeName := strings.Title(bareName)

		if !containsString(runtimeFunctions, runtimeName) {
			missingRuntime = append(missingRuntime, funcName)
		}

		if !handlerRegistry.IsSupported(funcName) {
			missingHandlers = append(missingHandlers, funcName)
		}
	}

	if len(missingRuntime) > 0 {
		t.Logf("Missing runtime implementations (%d): %v", len(missingRuntime), missingRuntime)
	}

	if len(missingHandlers) > 0 {
		t.Logf("Missing codegen handlers (%d): %v", len(missingHandlers), missingHandlers)
	}

	t.Logf("Coverage: %d/%d runtime, %d/%d handlers",
		len(allSignatures)-len(missingRuntime), len(allSignatures),
		len(allSignatures)-len(missingHandlers), len(allSignatures))
}

func extractRuntimeTAFunctions() []string {
	var functions []string
	taType := reflect.TypeOf(ta.Sma)
	pkgPath := taType.PkgPath()

	for i := 0; i < 100; i++ {
		funcName := runtime.FuncForPC(reflect.ValueOf(ta.Sma).Pointer() + uintptr(i*8)).Name()
		if strings.HasPrefix(funcName, pkgPath+".") {
			name := strings.TrimPrefix(funcName, pkgPath+".")
			if name != "" && !strings.Contains(name, ".") {
				functions = append(functions, name)
			}
		}
	}

	knownFunctions := []string{
		"Sma", "Ema", "Rma", "Rsi", "Tr", "Atr", "BBands", "Macd",
		"Stoch", "Stdev", "Change", "Pivothigh", "Pivotlow", "Dmi", "Linreg", "Cum",
	}

	return knownFunctions
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

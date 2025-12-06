package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// TempVariableManager manages lifecycle of temporary Series variables for inline TA calls.
//
// Purpose: Single Responsibility - generate unique temp var names, track mappings, manage registry
// Alignment: ForwardSeriesBuffer paradigm - ALL temp vars use Series storage
//
// Usage:
//
//	mgr := NewTempVariableManager(g)
//	varName := mgr.GetOrCreate(callInfo)  // "ta_sma_50_a1b2c3d4"
//	code := mgr.GenerateDeclaration()     // Declare all temp Series
//	code += mgr.GenerateInitialization()  // Generate TA calculation code
//
// Design:
//   - Deduplication: Same call expression → same temp var
//   - Unique naming: funcName + period + argHash
//   - Series lifecycle: Declaration, initialization, .Next() calls
type TempVariableManager struct {
	gen           *generator                     // Generator context
	callToVar     map[*ast.CallExpression]string // Deduplication map
	varToCallInfo map[string]CallInfo            // Reverse mapping for code generation
	declaredVars  map[string]bool                // Track which vars need declaration
}

// NewTempVariableManager creates manager with generator context
func NewTempVariableManager(g *generator) *TempVariableManager {
	return &TempVariableManager{
		gen:           g,
		callToVar:     make(map[*ast.CallExpression]string),
		varToCallInfo: make(map[string]CallInfo),
		declaredVars:  make(map[string]bool),
	}
}

// GetOrCreate returns existing temp var name or creates new unique name for call.
//
// Ensures: sma(close,50) and sma(close,200) get different names
// Format: {funcName}_{period}_{hash}
//
// Example:
//
//	sma(close, 50)  → ta_sma_50_a1b2c3d4
//	sma(close, 200) → ta_sma_200_e5f6g7h8
func (m *TempVariableManager) GetOrCreate(info CallInfo) string {
	// Check if already created (deduplication)
	if varName, exists := m.callToVar[info.Call]; exists {
		return varName
	}

	// Generate unique name: funcName + extracted params + hash
	varName := m.generateUniqueName(info)

	// Store mappings
	m.callToVar[info.Call] = varName
	m.varToCallInfo[varName] = info
	m.declaredVars[varName] = true

	// Temp vars managed exclusively by TempVariableManager (not g.variables)
	// Prevents double declaration: g.variables loop + GenerateDeclarations()

	return varName
}

// generateUniqueName creates descriptive unique variable name
//
// Strategy:
//  1. Extract period from first literal argument (if exists)
//  2. Combine: funcName + period + argHash
//  3. Sanitize for Go identifier rules
func (m *TempVariableManager) generateUniqueName(info CallInfo) string {
	// Base name from function
	baseName := strings.ReplaceAll(info.FuncName, ".", "_")

	// Try to extract period from arguments for readability
	period := m.extractPeriodFromCall(info.Call)

	// Build unique name
	if period > 0 {
		return fmt.Sprintf("%s_%d_%s", baseName, period, info.ArgHash)
	}
	return fmt.Sprintf("%s_%s", baseName, info.ArgHash)
}

// extractPeriodFromCall attempts to extract numeric period from call arguments
func (m *TempVariableManager) extractPeriodFromCall(call *ast.CallExpression) int {
	// Common pattern: ta.sma(source, period) - period is 2nd arg
	if len(call.Arguments) < 2 {
		return 0
	}

	// Check if second argument is literal number
	if lit, ok := call.Arguments[1].(*ast.Literal); ok {
		switch v := lit.Value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}

	return 0
}

// GenerateDeclarations outputs Series variable declarations for all temp vars
//
// Returns: var declarations block for top of strategy function
// Example:
//
//	var ta_sma_50_a1b2c3d4Series *series.Series
//	var ta_sma_200_e5f6g7h8Series *series.Series
func (m *TempVariableManager) GenerateDeclarations() string {
	if len(m.declaredVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""
	code += indent + "// Temp variables for inline TA calls in expressions\n"

	for varName := range m.declaredVars {
		code += indent + fmt.Sprintf("var %sSeries *series.Series\n", varName)
	}

	return code
}

// GenerateInitializations outputs Series.NewSeries() calls in initialization block
//
// Returns: Series initialization code
// Example:
//
//	ta_sma_50_a1b2c3d4Series = series.NewSeries(len(ctx.Data))
//	ta_sma_200_e5f6g7h8Series = series.NewSeries(len(ctx.Data))
func (m *TempVariableManager) GenerateInitializations() string {
	if len(m.declaredVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""

	for varName := range m.declaredVars {
		code += indent + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", varName)
	}

	return code
}

// GenerateCalculations outputs TA calculation code for all temp vars
//
// Returns: Inline TA calculation code using TAFunctionRegistry
// Example:
//
//	/* Inline ta.sma(50) */
//	if i >= 49 {
//	  sum := 0.0
//	  for j := 0; j < 50; j++ { ... }
//	  ta_sma_50_a1b2c3d4Series.Set(sum/50)
//	} else {
//	  ta_sma_50_a1b2c3d4Series.Set(math.NaN())
//	}
func (m *TempVariableManager) GenerateCalculations() (string, error) {
	if len(m.varToCallInfo) == 0 {
		return "", nil
	}

	if m.gen == nil {
		return "", fmt.Errorf("generator context required for calculations")
	}

	code := ""

	for varName, info := range m.varToCallInfo {
		// Use TAFunctionRegistry to generate inline calculation
		calcCode, err := m.gen.generateVariableFromCall(varName, info.Call)
		if err != nil {
			return "", fmt.Errorf("failed to generate temp var %s: %w", varName, err)
		}
		code += calcCode
	}

	return code, nil
}

// GenerateNextCalls outputs .Next() calls for bar advancement (ForwardSeriesBuffer paradigm)
//
// Returns: Series.Next() calls for end of bar loop
// Example:
//
//	if i < barCount-1 { ta_sma_50_a1b2c3d4Series.Next() }
//	if i < barCount-1 { ta_sma_200_e5f6g7h8Series.Next() }
func (m *TempVariableManager) GenerateNextCalls() string {
	if len(m.declaredVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""

	for varName := range m.declaredVars {
		code += indent + fmt.Sprintf("if i < barCount-1 { %sSeries.Next() }\n", varName)
	}

	return code
}

// GetVarNameForCall returns temp var name for call expression (for expression rewriting)
//
// Returns: Variable name if exists, empty string if not found
func (m *TempVariableManager) GetVarNameForCall(call *ast.CallExpression) string {
	return m.callToVar[call]
}

// Reset clears all state (for testing or multiple strategy generation)
func (m *TempVariableManager) Reset() {
	m.callToVar = make(map[*ast.CallExpression]string)
	m.varToCallInfo = make(map[string]CallInfo)
	m.declaredVars = make(map[string]bool)
}

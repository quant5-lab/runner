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
//   - Ordered generation: insertion order preserved for dependency correctness
type TempVariableManager struct {
	gen             *generator
	callToVar       map[*ast.CallExpression]string
	varToCallInfo   map[string]CallInfo
	conditionalVars map[string]*ast.ConditionalExpression
	orderedVars     []string
	emissionTracker *TempVarEmissionTracker
}

// NewTempVariableManager creates manager with generator context
func NewTempVariableManager(g *generator) *TempVariableManager {
	return &TempVariableManager{
		gen:             g,
		callToVar:       make(map[*ast.CallExpression]string),
		varToCallInfo:   make(map[string]CallInfo),
		conditionalVars: make(map[string]*ast.ConditionalExpression),
		emissionTracker: NewTempVarEmissionTracker(),
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
	// Check if already created (deduplication by AST pointer)
	if varName, exists := m.callToVar[info.Call]; exists {
		return varName
	}

	// Generate unique name: funcName + extracted params + hash
	varName := m.generateUniqueName(info)

	// Deduplicate by generated name (different AST nodes, same content)
	if _, exists := m.varToCallInfo[varName]; exists {
		m.callToVar[info.Call] = varName
		return varName
	}

	// Store mappings
	m.callToVar[info.Call] = varName
	m.varToCallInfo[varName] = info
	m.orderedVars = append(m.orderedVars, varName)

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
	if len(m.orderedVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""
	code += indent + "// Temp variables for inline TA calls in expressions\n"

	hasFixnan := false
	for _, varName := range m.orderedVars {
		code += indent + fmt.Sprintf("var %sSeries *series.Series\n", varName)

		/* Generate internal Series for composite indicators (RSI needs gains/losses Series) */
		if m.gen != nil && m.gen.compositeIndicatorRegistry != nil {
			info, exists := m.varToCallInfo[varName]
			if exists {
				internalNames := m.gen.compositeIndicatorRegistry.GetInternalSeriesNames(info.FuncName, varName, info.Call)
				for _, internalName := range internalNames {
					code += indent + fmt.Sprintf("var %sSeries *series.Series\n", internalName)
				}
				if info.FuncName == "fixnan" {
					hasFixnan = true
				}
			}
		}
	}

	/* fixnan requires cross-bar state variable for forward-fill */
	if hasFixnan {
		code += indent + "// State variables for fixnan forward-fill (temp vars)\n"
		for _, varName := range m.orderedVars {
			info, exists := m.varToCallInfo[varName]
			if exists && info.FuncName == "fixnan" {
				code += indent + fmt.Sprintf("var fixnanState_%s = math.NaN()\n", varName)
			}
		}
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
	if len(m.orderedVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""

	for _, varName := range m.orderedVars {
		code += indent + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", varName)

		/* Initialize internal Series for composite indicators */
		if m.gen != nil && m.gen.compositeIndicatorRegistry != nil {
			info, exists := m.varToCallInfo[varName]
			if exists {
				internalNames := m.gen.compositeIndicatorRegistry.GetInternalSeriesNames(info.FuncName, varName, info.Call)
				for _, internalName := range internalNames {
					code += indent + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", internalName)
				}
			}
		}
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
	if len(m.orderedVars) == 0 {
		return "", nil
	}

	if m.gen == nil {
		return "", fmt.Errorf("generator context required for calculations")
	}

	code := ""

	for _, varName := range m.orderedVars {
		calcCode, err := m.generateCalculationForVar(varName)
		if err != nil {
			return "", err
		}
		code += calcCode
	}

	return code, nil
}

func (m *TempVariableManager) GenerateCalculationsForStatement(stmtIdx int) (string, error) {
	if len(m.orderedVars) == 0 {
		return "", nil
	}

	if m.gen == nil {
		return "", fmt.Errorf("generator context required for calculations")
	}

	code := ""

	for _, varName := range m.orderedVars {
		info, exists := m.varToCallInfo[varName]
		if !exists || info.StmtIndex != stmtIdx {
			continue
		}
		calcCode, err := m.generateCalculationForVar(varName)
		if err != nil {
			return "", err
		}
		code += calcCode
		m.emissionTracker.MarkAsEmitted(varName)
	}

	return code, nil
}

func (m *TempVariableManager) generateCalculationForVar(varName string) (string, error) {
	info, exists := m.varToCallInfo[varName]
	if !exists {
		return "", nil
	}
	calcCode, err := m.gen.generateVariableFromCall(varName, info.Call)
	if err != nil {
		return "", fmt.Errorf("failed to generate temp var %s: %w", varName, err)
	}
	return calcCode, nil
}

// GenerateNextCalls outputs .Next() calls for bar advancement (ForwardSeriesBuffer paradigm)
//
// Returns: Series.Next() calls for end of bar loop
// Example:
//
//	if i < barCount-1 { ta_sma_50_a1b2c3d4Series.Next() }
//	if i < barCount-1 { ta_sma_200_e5f6g7h8Series.Next() }
func (m *TempVariableManager) GenerateNextCalls() string {
	if len(m.orderedVars) == 0 {
		return ""
	}

	indent := ""
	if m.gen != nil {
		indent = m.gen.ind()
	}

	code := ""

	for _, varName := range m.orderedVars {
		code += indent + fmt.Sprintf("if i < barCount-1 { %sSeries.Next() }\n", varName)

		if m.gen != nil && m.gen.compositeIndicatorRegistry != nil {
			info, exists := m.varToCallInfo[varName]
			if exists {
				internalNames := m.gen.compositeIndicatorRegistry.GetInternalSeriesNames(info.FuncName, varName, info.Call)
				for _, internalName := range internalNames {
					code += indent + fmt.Sprintf("if i < barCount-1 { %sSeries.Next() }\n", internalName)
				}
			}
		}
	}

	return code
}

// GetVarNameForCall returns temp var name for call expression (for expression rewriting)
//
// Returns: Variable name if exists, empty string if not found
func (m *TempVariableManager) GetVarNameForCall(call *ast.CallExpression) string {
	return m.callToVar[call]
}

func (m *TempVariableManager) WasAlreadyEmitted(varName string) bool {
	return m.emissionTracker.WasEmitted(varName)
}

// Reset clears all state (for testing or multiple strategy generation)
func (m *TempVariableManager) Reset() {
	m.callToVar = make(map[*ast.CallExpression]string)
	m.varToCallInfo = make(map[string]CallInfo)
	m.conditionalVars = make(map[string]*ast.ConditionalExpression)
	m.orderedVars = nil
	m.emissionTracker.Reset()
}

func (m *TempVariableManager) RegisterConditional(hash string, cond *ast.ConditionalExpression) string {
	if existingVar, exists := m.conditionalVars[hash]; exists {
		for varName, storedCond := range m.conditionalVars {
			if storedCond == existingVar {
				return varName
			}
		}
	}

	varName := fmt.Sprintf("conditional_%s", hash)
	m.conditionalVars[varName] = cond
	return varName
}

func (m *TempVariableManager) GetConditionalByHash(hash string) *ast.ConditionalExpression {
	varName := fmt.Sprintf("conditional_%s", hash)
	return m.conditionalVars[varName]
}

func (m *TempVariableManager) GetConditionalVarName(hash string) string {
	varName := fmt.Sprintf("conditional_%s", hash)
	if _, exists := m.conditionalVars[varName]; exists {
		return varName
	}
	return ""
}

func (m *TempVariableManager) GetAllConditionals() map[string]*ast.ConditionalExpression {
	return m.conditionalVars
}

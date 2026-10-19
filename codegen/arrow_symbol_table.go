package codegen

import "github.com/quant5-lab/runner/ast"

// buildArrowSymbolTable constructs a scope-aware SymbolTable for an arrow function body.
//
// All arrow-local variables and TA-source parameters are series-backed
// (arrowCtx.GetOrCreateSeries), so they must resolve as VariableTypeSeries
// inside TA init loops. Scalar parameters carry a single float64 per bar and
// must NOT be treated as series.
//
// The returned table inherits all main-scope entries (via clone) so that
// outer-scope series variables referenced inside the arrow body continue to
// resolve correctly.
func buildArrowSymbolTable(
	mainScope SymbolTable,
	localVarNames []string,
	params []ast.Identifier,
	paramUsage map[string]ParameterUsageType,
) SymbolTable {
	var table SymbolTable
	if mainScope != nil {
		table = mainScope.Clone()
	} else {
		table = NewSymbolTable()
	}

	for _, name := range localVarNames {
		table.Register(name, VariableTypeSeries)
	}

	for _, p := range params {
		switch paramUsage[p.Name] {
		case ParameterUsageTASource, ParameterUsageSeries:
			table.Register(p.Name, VariableTypeSeries)
		default:
			table.Register(p.Name, VariableTypeScalar)
		}
	}

	return table
}

// activeSymbolTable ensures TA init loops inside arrow functions resolve arrow-local
// variable names to their backing series, falling back to main scope elsewhere.
func activeSymbolTable(gen *generator) SymbolTable {
	if gen.arrowSymbolTable != nil {
		return gen.arrowSymbolTable
	}
	return gen.symbolTable
}

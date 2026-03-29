package codegen

import (
	"fmt"
)

const barLoopEvaluatorVar = "secBarEvaluator"
const ArrowEvalMapVar = "arrowSecEvals"
const arrowEvaluatorExpr = ArrowEvalMapVar + "[secKey]"

/*
SecurityEvaluatorInitializer centralises evaluator wiring so bar-loop, inline, and
arrow call-sites emit identical BarMapper, VarLookup, and InputConstants setup code.

evaluatorVar controls which Go variable/expression holds the persistent evaluator:
  - bar-loop / inline:  barLoopEvaluatorVar  ("secBarEvaluator")
  - arrow IIFE:         arrowEvaluatorExpr   ("arrowSecEvals[secKey]")

Used by:
  - SecurityExpressionHandler (bar-loop scope)
  - SecurityInlineHandler (ternary/conditional inline scope)
  - ArrowSecurityCallGenerator (arrow function scope)
*/
type SecurityEvaluatorInitializer struct {
	symbolTable  SymbolTable
	gen          *generator
	evaluatorVar string
}

func NewSecurityEvaluatorInitializer(symbolTable SymbolTable, gen *generator) *SecurityEvaluatorInitializer {
	return &SecurityEvaluatorInitializer{
		symbolTable:  symbolTable,
		gen:          gen,
		evaluatorVar: barLoopEvaluatorVar,
	}
}

func NewArrowSecurityEvaluatorInitializer(symbolTable SymbolTable, gen *generator) *SecurityEvaluatorInitializer {
	return &SecurityEvaluatorInitializer{
		symbolTable:  symbolTable,
		gen:          gen,
		evaluatorVar: arrowEvaluatorExpr,
	}
}

func (init *SecurityEvaluatorInitializer) EmitInitialization(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := ""

	code += indentFunc() + "if " + init.evaluatorVar + " == nil {\n"
	incrementIndent()

	code += init.EmitInitializationBody(indentFunc, incrementIndent, decrementIndent)
	code += indentFunc() + init.evaluatorVar + " = security.NewSeriesCachingEvaluator(baseEvaluator)\n"

	decrementIndent()
	code += indentFunc() + "}\n"

	return code
}

/*
EmitInitializationBody emits the evaluator wiring (NewStreamingBarEvaluator, registry, bar mapper,
var lookup, input constants) without the outer nil guard or final assignment. Used by
ArrowSecurityCallGenerator which manages its own guard and assignment to an interface{} map.
*/
func (init *SecurityEvaluatorInitializer) EmitInitializationBody(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := ""
	code += indentFunc() + "baseEvaluator := security.NewStreamingBarEvaluator()\n"
	code += indentFunc() + "varRegistry := security.NewVariableRegistry()\n"
	code += indentFunc() + "baseEvaluator.SetVariableRegistry(varRegistry)\n"
	code += init.emitBarMapperSetup(indentFunc, incrementIndent, decrementIndent)
	code += init.emitVarLookupSetup(indentFunc, incrementIndent, decrementIndent)
	code += init.emitInputConstantsSetup(indentFunc)
	return code
}

func (init *SecurityEvaluatorInitializer) emitBarMapperSetup(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := ""
	code += indentFunc() + "barMapper := security.NewBarIndexMapper()\n"
	code += indentFunc() + "switch securityBarMapper.Mode() {\n"
	code += init.emitIdentityBarMappingBranch(indentFunc, incrementIndent, decrementIndent)
	code += init.emitTransformedBarMappingBranch(indentFunc, incrementIndent, decrementIndent)
	code += init.emitRangeBarMappingBranch(indentFunc, incrementIndent, decrementIndent)
	code += indentFunc() + "}\n"
	code += indentFunc() + "baseEvaluator.SetBarIndexMapper(barMapper)\n"
	return code
}

func (init *SecurityEvaluatorInitializer) emitIdentityBarMappingBranch(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := indentFunc() + "case request.ModeIdentity:\n"
	incrementIndent()
	code += indentFunc() + "for i := range secCtx.Data {\n"
	incrementIndent()
	code += indentFunc() + "barMapper.SetMapping(i, i)\n"
	decrementIndent()
	code += indentFunc() + "}\n"
	decrementIndent()
	return code
}

func (init *SecurityEvaluatorInitializer) emitTransformedBarMappingBranch(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := indentFunc() + "case request.ModeTransformed:\n"
	incrementIndent()
	code += indentFunc() + "for mainIdx, secIdx := range securityBarMapper.MainToSynthetic() {\n"
	incrementIndent()
	code += indentFunc() + "if secIdx >= 0 && barMapper.GetMainBarIndexForSecurityBar(secIdx) < 0 {\n"
	incrementIndent()
	code += indentFunc() + "barMapper.SetMapping(secIdx, mainIdx)\n"
	decrementIndent()
	code += indentFunc() + "}\n"
	decrementIndent()
	code += indentFunc() + "}\n"
	decrementIndent()
	return code
}

func (init *SecurityEvaluatorInitializer) emitRangeBarMappingBranch(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := indentFunc() + "default:\n"
	incrementIndent()
	code += indentFunc() + "for _, rr := range securityBarMapper.GetRanges() {\n"
	incrementIndent()
	code += indentFunc() + "if rr.StartHourlyIndex >= 0 {\n"
	incrementIndent()
	code += indentFunc() + "barMapper.SetMapping(rr.DailyBarIndex, rr.StartHourlyIndex)\n"
	decrementIndent()
	code += indentFunc() + "}\n"
	decrementIndent()
	code += indentFunc() + "}\n"
	decrementIndent()
	return code
}

func (init *SecurityEvaluatorInitializer) emitVarLookupSetup(
	indentFunc func() string,
	incrementIndent func(),
	decrementIndent func(),
) string {
	code := ""

	code += indentFunc() + "baseEvaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {\n"
	incrementIndent()

	code += indentFunc() + "var varSeries *series.Series\n"
	code += indentFunc() + "switch varName {\n"

	taFunctions := map[string]bool{
		"minus": true, "plus": true, "sum": true, "truerange": true,
		"abs": true, "max": true, "min": true, "sign": true,
	}

	if init.symbolTable != nil {
		for _, symbol := range init.symbolTable.AllSymbols() {
			if symbol.Type == VariableTypeSeries {
				varName := symbol.Name
				if taFunctions[varName] {
					continue
				}
				code += indentFunc() + fmt.Sprintf("case %q:\n", varName)
				incrementIndent()
				code += indentFunc() + fmt.Sprintf("varSeries = %sSeries\n", varName)
				decrementIndent()
			}
		}
	}

	code += indentFunc() + "default:\n"
	incrementIndent()
	code += indentFunc() + "return nil, -1, false\n"
	decrementIndent()
	code += indentFunc() + "}\n"

	code += indentFunc() + "if varSeries == nil {\n"
	incrementIndent()
	code += indentFunc() + "return nil, -1, false\n"
	decrementIndent()
	code += indentFunc() + "}\n"

	code += indentFunc() + "mainIdx := barMapper.GetMainBarIndexForSecurityBar(secBarIdx)\n"
	code += indentFunc() + "return varSeries, mainIdx, true\n"

	decrementIndent()
	code += indentFunc() + "})\n"

	return code
}

func (init *SecurityEvaluatorInitializer) emitInputConstantsSetup(indentFunc func() string) string {
	code := ""

	inputConstantsMap := init.generateInputConstantsMap()
	code += indentFunc() + "inputConstantsMap := " + inputConstantsMap + "\n"
	code += indentFunc() + "baseEvaluator.SetInputConstantsMap(inputConstantsMap)\n"

	return code
}

func (init *SecurityEvaluatorInitializer) generateInputConstantsMap() string {
	if init.gen.inputHandler == nil {
		return "map[string]float64(nil)"
	}

	constantsMap := init.gen.inputHandler.GetInputConstantsMap()
	if len(constantsMap) == 0 {
		return "map[string]float64(nil)"
	}

	result := "map[string]float64{"
	first := true
	for varName, value := range constantsMap {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%q: %f", varName, value)
		first = false
	}
	result += "}"
	return result
}

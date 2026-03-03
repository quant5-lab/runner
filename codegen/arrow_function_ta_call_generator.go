package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/codegen/series_naming"
)

type ArrowFunctionTACallGenerator struct {
	gen               *generator
	exprGen           ArrowExpressionGenerator
	iifeRegistry      *InlineTAIIFERegistry
	tupleRegistry     *TupleIndicatorRegistry
	accessorFactory   *ArrowAwareAccessorFactory
	signatureResolver *ArrowTACallSignatureResolver
}

func NewArrowFunctionTACallGenerator(gen *generator, exprGen ArrowExpressionGenerator) *ArrowFunctionTACallGenerator {
	accessResolver := NewArrowSeriesAccessResolver()

	for paramName, paramType := range gen.variables {
		if paramType == "float" {
			accessResolver.RegisterParameter(paramName)
		}
	}

	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	accessorFactory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

	return &ArrowFunctionTACallGenerator{
		gen:               gen,
		exprGen:           exprGen,
		iifeRegistry:      NewInlineTAIIFERegistry(),
		tupleRegistry:     NewTupleIndicatorRegistry(),
		accessorFactory:   accessorFactory,
		signatureResolver: NewArrowTACallSignatureResolver(),
	}
}

func (a *ArrowFunctionTACallGenerator) Generate(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	detector := NewUserDefinedFunctionDetector(a.gen.variables)
	if detector.IsUserDefinedFunction(funcName) {
		handler := &UserDefinedFunctionHandler{}
		return handler.GenerateCode(a.gen, call)
	}

	if funcName == "fixnan" || funcName == "ta.fixnan" {
		return a.generateFixnanIIFE(call)
	}

	if funcName == "valuewhen" || funcName == "ta.valuewhen" {
		return a.generateValuewhenIIFE(call)
	}

	pivotResolver := NewPivotSignatureResolver()
	if pivotResolver.IsPivotFunction(funcName) {
		return a.generatePivotCall(funcName, call)
	}

	if a.iifeRegistry.IsRegisteredDualPeriod(funcName) {
		return a.generateDualPeriodTACall(funcName, call)
	}

	if a.tupleRegistry.IsRegistered(funcName) {
		return a.generateTupleIIFE(funcName, call)
	}

	if funcName == "ta.bbw" || funcName == "bbw" {
		return a.generateBBWCall(call)
	}

	accessor, periodExpr, err := a.extractTAArguments(funcName, call)
	if err != nil {
		return "", fmt.Errorf("failed to extract TA arguments: %w", err)
	}

	sourceHash := ""
	if len(call.Arguments) > 0 {
		hasher := &ExpressionHasher{}
		sourceHash = hasher.Hash(call.Arguments[0])
	}

	code, ok := a.iifeRegistry.Generate(funcName, accessor, periodExpr, sourceHash)
	if !ok {
		return "", fmt.Errorf("TA function %s requires IIFE generator implementation", funcName)
	}

	/* Wrap with preamble when accessor needs per-bar series storage for TA-as-source chaining */
	if preambleAccessor, ok := accessor.(interface{ GetPreamble() string }); ok {
		preamble := preambleAccessor.GetPreamble()
		if preamble != "" {
			return fmt.Sprintf("func() float64 { %s\nreturn %s }()", preamble, code), nil
		}
	}

	return code, nil
}

func (a *ArrowFunctionTACallGenerator) generateFixnanIIFE(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("fixnan requires 1 argument")
	}

	sourceArg := call.Arguments[0]
	sourceCode, err := a.gen.generateArrowFunctionExpression(sourceArg)
	if err != nil {
		return "", fmt.Errorf("failed to generate fixnan source: %w", err)
	}

	return fmt.Sprintf("func() float64 { val := %s; if math.IsNaN(val) { return 0.0 }; return val }()", sourceCode), nil
}

/* Scans backward through bars for the Nth true condition, returns source at that bar */
func (a *ArrowFunctionTACallGenerator) generateValuewhenIIFE(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("valuewhen requires 3 arguments (condition, source, occurrence)")
	}

	condAccessor, err := a.accessorFactory.CreateAccessorForExpression(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for valuewhen condition: %w", err)
	}

	srcAccessor, err := a.accessorFactory.CreateAccessorForExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for valuewhen source: %w", err)
	}

	occurrenceExpr, err := a.extractPeriodExpression(call.Arguments[2])
	if err != nil {
		return "", fmt.Errorf("failed to extract valuewhen occurrence: %w", err)
	}

	preamble := ""
	if pa, ok := condAccessor.(interface{ GetPreamble() string }); ok {
		if p := pa.GetPreamble(); p != "" {
			preamble += p + "; "
		}
	}
	if pa, ok := srcAccessor.(interface{ GetPreamble() string }); ok {
		if p := pa.GetPreamble(); p != "" {
			preamble += p + "; "
		}
	}

	body := fmt.Sprintf("occurrence := %s; count := 0; ", occurrenceExpr.AsIntCast())
	body += "for j := 0; j <= ctx.BarIndex; j++ { "
	body += fmt.Sprintf("condVal := %s; ", condAccessor.GenerateLoopValueAccess("j"))
	body += "if !math.IsNaN(condVal) && condVal != 0 { "
	body += "if count == occurrence { "
	body += fmt.Sprintf("return %s", srcAccessor.GenerateLoopValueAccess("j"))
	body += " }; count++ } }; return math.NaN()"

	if preamble != "" {
		return fmt.Sprintf("func() float64 { %s%s }()", preamble, body), nil
	}
	return fmt.Sprintf("func() float64 { %s }()", body), nil
}

func (a *ArrowFunctionTACallGenerator) generatePivotCall(funcName string, call *ast.CallExpression) (string, error) {
	pivotResolver := NewPivotSignatureResolver()
	resolved, err := pivotResolver.Resolve(funcName, call)
	if err != nil {
		return "", fmt.Errorf("failed to resolve pivot signature: %w", err)
	}

	accessor, err := a.accessorFactory.CreateAccessorForExpression(resolved.SourceExpr)
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for pivot source: %w", err)
	}

	leftPeriod, err := a.extractPeriodExpression(resolved.LeftPeriod)
	if err != nil {
		return "", fmt.Errorf("failed to extract left period: %w", err)
	}

	rightPeriod, err := a.extractPeriodExpression(resolved.RightPeriod)
	if err != nil {
		return "", fmt.Errorf("failed to extract right period: %w", err)
	}

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(resolved.SourceExpr)

	code, ok := a.iifeRegistry.GenerateDualPeriod(funcName, accessor, leftPeriod, rightPeriod, sourceHash)
	if !ok {
		return "", fmt.Errorf("pivot function %s requires dual-period IIFE generator", funcName)
	}

	return code, nil
}

func (a *ArrowFunctionTACallGenerator) generateDualPeriodTACall(funcName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("%s requires 3 arguments (source, shortPeriod, longPeriod)", funcName)
	}

	accessor, err := a.accessorFactory.CreateAccessorForExpression(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("failed to create accessor for %s: %w", funcName, err)
	}

	leftPeriod, err := a.extractPeriodExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("failed to extract first period for %s: %w", funcName, err)
	}

	rightPeriod, err := a.extractPeriodExpression(call.Arguments[2])
	if err != nil {
		return "", fmt.Errorf("failed to extract second period for %s: %w", funcName, err)
	}

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(call.Arguments[0])

	code, ok := a.iifeRegistry.GenerateDualPeriod(funcName, accessor, leftPeriod, rightPeriod, sourceHash)
	if !ok {
		return "", fmt.Errorf("dual-period IIFE generator for %s not found", funcName)
	}

	return code, nil
}

func (a *ArrowFunctionTACallGenerator) generateBBWCall(call *ast.CallExpression) (string, error) {
	accessor, periodExpr, err := a.extractTAArguments("ta.bbw", call)
	if err != nil {
		return "", fmt.Errorf("failed to extract BBW arguments: %w", err)
	}

	multExtractor := NewBBWMultExtractor(a.exprGen)
	multResult, err := multExtractor.Extract(call)
	if err != nil {
		return "", fmt.Errorf("failed to extract mult: %w", err)
	}

	generator := a.createBBWGenerator(multResult)

	sourceHash := ""
	if len(call.Arguments) > 0 {
		hasher := &ExpressionHasher{}
		sourceHash = hasher.Hash(call.Arguments[0])
	}

	code := generator.Generate(accessor, periodExpr, sourceHash)

	if preambleAccessor, ok := accessor.(interface{ GetPreamble() string }); ok {
		preamble := preambleAccessor.GetPreamble()
		if preamble != "" {
			return fmt.Sprintf("func() float64 { %s\nreturn %s }()", preamble, code), nil
		}
	}

	return code, nil
}

func (a *ArrowFunctionTACallGenerator) extractTAArguments(funcName string, call *ast.CallExpression) (AccessGenerator, PeriodExpression, error) {
	if funcName == "ta.change" || funcName == "change" {
		return a.extractChangeArguments(call)
	}

	resolved, err := a.signatureResolver.ResolveCall(funcName, call)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve TA call signature: %w", err)
	}

	var sourceArg ast.Expression
	if resolved.NeedsDefaultSource {
		sourceArg = &ast.Identifier{Name: resolved.DefaultSourceName}
	} else {
		sourceArg = resolved.SourceExpr
	}

	/* OHLC-only functions (atr, tr) have no explicit source — generator handles OHLC directly */
	var accessor AccessGenerator
	if sourceArg != nil {
		accessor, err = a.accessorFactory.CreateAccessorForExpression(sourceArg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create accessor: %w", err)
		}
	}

	periodExpr, err := a.extractPeriodExpression(resolved.LengthExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract period: %w", err)
	}

	return accessor, periodExpr, nil
}

func (a *ArrowFunctionTACallGenerator) extractChangeArguments(call *ast.CallExpression) (AccessGenerator, PeriodExpression, error) {
	if len(call.Arguments) < 1 {
		return nil, nil, fmt.Errorf("change() requires at least 1 argument (source)")
	}

	sourceArg := call.Arguments[0]
	accessor, err := a.accessorFactory.CreateAccessorForExpression(sourceArg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor for change(): %w", err)
	}

	offsetExpr := PeriodExpression(NewConstantPeriod(1))
	if len(call.Arguments) >= 2 {
		extracted, err := a.extractPeriodExpression(call.Arguments[1])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract offset for change(): %w", err)
		}
		offsetExpr = extracted
	}

	return accessor, offsetExpr, nil
}

func (a *ArrowFunctionTACallGenerator) extractPeriodExpression(expr ast.Expression) (PeriodExpression, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		if floatVal, ok := e.Value.(float64); ok {
			return NewConstantPeriod(int(floatVal)), nil
		}
		if intVal, ok := e.Value.(int); ok {
			return NewConstantPeriod(intVal), nil
		}
		if strVal, ok := e.Value.(string); ok {
			periodInt, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, fmt.Errorf("period string is not numeric: %s", strVal)
			}
			return NewConstantPeriod(periodInt), nil
		}
		return nil, fmt.Errorf("period literal is not numeric: %v", e.Value)

	case *ast.Identifier:
		if _, exists := a.gen.variables[e.Name]; !exists {
			return nil, fmt.Errorf("unknown period identifier: %s (not an arrow function parameter)", e.Name)
		}
		return NewRuntimePeriod(e.Name), nil

	default:
		rendered, err := a.exprGen.Generate(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to render period expression: %w", err)
		}
		return NewComputedPeriod(rendered), nil
	}
}

type ArrowFunctionParameterAccessor struct {
	parameterName string
}

func NewArrowFunctionParameterAccessor(parameterName string) *ArrowFunctionParameterAccessor {
	return &ArrowFunctionParameterAccessor{
		parameterName: parameterName,
	}
}

func (a *ArrowFunctionParameterAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", a.parameterName, loopVar)
}

func (a *ArrowFunctionParameterAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%sSeries.Get(%d-1)", a.parameterName, period)
}

func (a *ArrowFunctionParameterAccessor) GenerateCurrentValueAccess() string {
	return fmt.Sprintf("%sSeries.GetCurrent()", a.parameterName)
}

func (a *ArrowFunctionTACallGenerator) generateTupleIIFE(funcName string, call *ast.CallExpression) (string, error) {
	spec := a.tupleRegistry.Lookup(funcName)
	if spec == nil {
		return "", fmt.Errorf("tuple indicator %s not registered", funcName)
	}

	argBuilder := NewTupleArgumentBuilder(a.gen, a.accessorFactory)
	argExpressions, err := argBuilder.BuildArguments(call.Arguments, spec)
	if err != nil {
		return "", fmt.Errorf("failed to build tuple arguments: %w", err)
	}

	argList := ""
	for i, argExpr := range argExpressions {
		if i > 0 {
			argList += ", "
		}
		argList += argExpr
	}

	returnType := "("
	for i := 0; i < spec.OutputCount; i++ {
		if i > 0 {
			returnType += ", "
		}
		returnType += "float64"
	}
	returnType += ")"

	return fmt.Sprintf("func() %s { return %s(%s) }()", returnType, spec.RuntimeFunction, argList), nil
}

type TupleArgumentBuilder struct {
	gen             *generator
	accessorFactory *ArrowAwareAccessorFactory
}

func NewTupleArgumentBuilder(gen *generator, accessorFactory *ArrowAwareAccessorFactory) *TupleArgumentBuilder {
	return &TupleArgumentBuilder{
		gen:             gen,
		accessorFactory: accessorFactory,
	}
}

func (b *TupleArgumentBuilder) BuildArguments(args []ast.Expression, spec *TupleIndicatorSpec) ([]string, error) {
	argExpressions := make([]string, len(args))

	for i, arg := range args {
		if b.isSourceArgument(i, spec) {
			accessor, err := b.accessorFactory.CreateAccessorForExpression(arg)
			if err != nil {
				return nil, fmt.Errorf("failed to create accessor for source arg %d: %w", i, err)
			}
			argExpressions[i] = accessor.GenerateCurrentValueAccess()
		} else {
			argCode, err := b.gen.generateArrowFunctionExpression(arg)
			if err != nil {
				return nil, fmt.Errorf("failed to generate arg %d: %w", i, err)
			}
			argExpressions[i] = argCode
		}
	}

	return argExpressions, nil
}

func (b *TupleArgumentBuilder) isSourceArgument(argIndex int, spec *TupleIndicatorSpec) bool {
	return spec.SourceArgIndex == argIndex
}

func (a *ArrowFunctionParameterAccessor) GetBaseOffset() int {
	return 0
}

func (a *ArrowFunctionTACallGenerator) createBBWGenerator(multResult *BBWMultResult) *BBWIIFEGenerator {
	if multResult.IsLiteral {
		return &BBWIIFEGenerator{
			namingStrategy: series_naming.NewWindowBasedNamer(),
			multLiteral:    multResult.Literal,
			useLiteralMult: true,
		}
	}

	return &BBWIIFEGenerator{
		namingStrategy: series_naming.NewWindowBasedNamer(),
		multExpression: multResult.Expression,
		useLiteralMult: false,
	}
}

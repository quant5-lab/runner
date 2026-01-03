package codegen

import "github.com/quant5-lab/runner/ast"

type ConversionRule interface {
	ShouldConvert(expr ast.Expression, code string) bool
}

type skipComparisonRule struct {
	comparisonMatcher PatternMatcher
}

func (r *skipComparisonRule) ShouldConvert(expr ast.Expression, code string) bool {
	return !r.comparisonMatcher.Matches(code)
}

type convertSeriesAccessRule struct {
	seriesMatcher PatternMatcher
}

func (r *convertSeriesAccessRule) ShouldConvert(expr ast.Expression, code string) bool {
	return r.seriesMatcher.Matches(code)
}

type typeBasedRule struct {
	typeSystem *TypeInferenceEngine
}

func (r *typeBasedRule) ShouldConvert(expr ast.Expression, code string) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		if r.typeSystem.IsBoolConstant(ident.Name) {
			return false
		}
		if r.typeSystem.IsBoolVariableByName(ident.Name) {
			return false
		}
		varType, exists := r.typeSystem.variables[ident.Name]
		if exists && varType != "bool" {
			return true
		}
		return false
	}

	if member, ok := expr.(*ast.MemberExpression); ok {
		if ident, ok := member.Object.(*ast.Identifier); ok {
			if r.typeSystem.IsBoolConstant(ident.Name) {
				return false
			}
			if r.typeSystem.IsBoolVariableByName(ident.Name) {
				return false
			}
			varType, exists := r.typeSystem.variables[ident.Name]
			if exists && varType != "bool" {
				return true
			}
			return false
		}
	}

	return false
}

func NewSkipComparisonRule(comparisonMatcher PatternMatcher) ConversionRule {
	return &skipComparisonRule{comparisonMatcher: comparisonMatcher}
}

func NewConvertSeriesAccessRule(seriesMatcher PatternMatcher) ConversionRule {
	return &convertSeriesAccessRule{seriesMatcher: seriesMatcher}
}

func NewTypeBasedRule(typeSystem *TypeInferenceEngine) ConversionRule {
	return &typeBasedRule{typeSystem: typeSystem}
}

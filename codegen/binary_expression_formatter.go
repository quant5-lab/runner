package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ExpressionGenerator interface {
	Generate(ast.Expression) (string, error)
}

type ExpressionExtractor interface {
	Extract(ast.Expression) string
}

type BinaryExpressionFormatter struct {
	generator ExpressionGenerator
	extractor ExpressionExtractor
}

func NewBinaryExpressionFormatter(generator func(ast.Expression) (string, error)) *BinaryExpressionFormatter {
	return &BinaryExpressionFormatter{
		generator: exprGeneratorFunc(generator),
	}
}

func NewBinaryExpressionFormatterWithExtractor(extractor func(ast.Expression) string) *BinaryExpressionFormatter {
	return &BinaryExpressionFormatter{
		extractor: exprExtractorFunc(extractor),
	}
}

type exprGeneratorFunc func(ast.Expression) (string, error)

func (f exprGeneratorFunc) Generate(expr ast.Expression) (string, error) {
	return f(expr)
}

type exprExtractorFunc func(ast.Expression) string

func (f exprExtractorFunc) Extract(expr ast.Expression) string {
	return f(expr)
}

func (f *BinaryExpressionFormatter) Format(binExpr *ast.BinaryExpression) (string, error) {
	if f.generator != nil {
		return f.formatWithGenerator(binExpr)
	}
	return f.formatWithExtractor(binExpr), nil
}

func (f *BinaryExpressionFormatter) formatWithGenerator(binExpr *ast.BinaryExpression) (string, error) {
	left, err := f.formatOperandWithGenerator(binExpr.Left, binExpr.Operator, false)
	if err != nil {
		return "", err
	}

	right, err := f.formatOperandWithGenerator(binExpr.Right, binExpr.Operator, true)
	if err != nil {
		return "", err
	}

	if binExpr.Operator == "%" {
		return fmt.Sprintf("float64(int(%s) %s int(%s))", left, binExpr.Operator, right), nil
	}

	result := fmt.Sprintf("(%s %s %s)", left, binExpr.Operator, right)
	return result, nil
}

func (f *BinaryExpressionFormatter) formatWithExtractor(binExpr *ast.BinaryExpression) string {
	left := f.formatOperandWithExtractor(binExpr.Left, binExpr.Operator, false)
	right := f.formatOperandWithExtractor(binExpr.Right, binExpr.Operator, true)

	if binExpr.Operator == "%" {
		return fmt.Sprintf("float64(int(%s) %s int(%s))", left, binExpr.Operator, right)
	}

	return fmt.Sprintf("(%s %s %s)", left, binExpr.Operator, right)
}

func (f *BinaryExpressionFormatter) formatOperandWithGenerator(operand ast.Expression, parentOp string, isRightChild bool) (string, error) {
	if childBin, ok := operand.(*ast.BinaryExpression); ok {
		childCode, err := f.formatWithGenerator(childBin)
		if err != nil {
			return "", err
		}

		if NeedsParentheses(childBin.Operator, parentOp, isRightChild) {
			return childCode, nil
		}

		if len(childCode) > 0 && childCode[0] == '(' && childCode[len(childCode)-1] == ')' {
			return childCode[1 : len(childCode)-1], nil
		}

		return childCode, nil
	}

	return f.generator.Generate(operand)
}

func (f *BinaryExpressionFormatter) formatOperandWithExtractor(operand ast.Expression, parentOp string, isRightChild bool) string {
	if childBin, ok := operand.(*ast.BinaryExpression); ok {
		childCode := f.formatWithExtractor(childBin)

		if NeedsParentheses(childBin.Operator, parentOp, isRightChild) {
			return childCode
		}

		if len(childCode) > 0 && childCode[0] == '(' && childCode[len(childCode)-1] == ')' {
			return childCode[1 : len(childCode)-1]
		}

		return childCode
	}

	return f.extractor.Extract(operand)
}

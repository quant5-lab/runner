package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* TradeCollectionCallHandler adapts TradeCollectionMemberHandler to CallExpressionHandler. */
type TradeCollectionCallHandler struct {
	memberHandler *TradeCollectionMemberHandler
}

func NewTradeCollectionCallHandler() *TradeCollectionCallHandler {
	return &TradeCollectionCallHandler{
		memberHandler: NewTradeCollectionMemberHandler(),
	}
}

func (h *TradeCollectionCallHandler) CanHandle(funcName string) bool {
	if !strings.HasPrefix(funcName, "strategy.") {
		return false
	}

	parts := strings.SplitN(funcName, ".", 3)
	if len(parts) != 3 {
		return false
	}

	return h.memberHandler.CanHandle(parts[1], parts[2])
}

func (h *TradeCollectionCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	callee, ok := call.Callee.(*ast.MemberExpression)
	if !ok {
		return "", nil
	}

	property, ok := callee.Property.(*ast.Identifier)
	if !ok {
		return "", nil
	}

	collection := extractCollectionName(callee.Object)
	if collection == "" {
		return "", nil
	}

	if !h.memberHandler.CanHandle(collection, property.Name) {
		return "", nil
	}

	exprGen := &generatorExpressionAdapter{g: g}
	return h.memberHandler.GenerateAccess(collection, property.Name, call.Arguments, exprGen)
}

/* extractCollectionName resolves the collection name from a strategy.{collection} MemberExpression. */
func extractCollectionName(expr ast.Expression) string {
	member, ok := expr.(*ast.MemberExpression)
	if !ok {
		return ""
	}

	obj, ok := member.Object.(*ast.Identifier)
	if !ok || obj.Name != "strategy" {
		return ""
	}

	prop, ok := member.Property.(*ast.Identifier)
	if !ok {
		return ""
	}

	return prop.Name
}

/* generatorExpressionAdapter bridges generator to ExpressionGenerator for TradeCollectionMemberHandler. */
type generatorExpressionAdapter struct {
	g *generator
}

func (a *generatorExpressionAdapter) Generate(expr ast.Expression) (string, error) {
	return a.g.generateArrowFunctionExpression(expr)
}

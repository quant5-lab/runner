package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// TradeCollectionCallHandler adapts TradeCollectionMemberHandler to CallExpressionHandler interface.
//
// Purpose: Routes strategy.closedtrades.profit(N) and strategy.opentrades.size(N) calls
//
//	to the TradeCollectionMemberHandler for code generation.
//
// Pattern: Adapter pattern - bridges two interfaces
type TradeCollectionCallHandler struct {
	memberHandler *TradeCollectionMemberHandler
}

// NewTradeCollectionCallHandler creates adapter instance.
func NewTradeCollectionCallHandler() *TradeCollectionCallHandler {
	return &TradeCollectionCallHandler{
		memberHandler: NewTradeCollectionMemberHandler(),
	}
}

// CanHandle checks if function name matches strategy.{closedtrades|opentrades}.{property}
func (h *TradeCollectionCallHandler) CanHandle(funcName string) bool {
	if !strings.HasPrefix(funcName, "strategy.") {
		return false
	}

	// Extract object and member from "strategy.closedtrades.profit" → "closedtrades", "profit"
	parts := strings.SplitN(funcName, ".", 3)
	if len(parts) != 3 {
		return false
	}

	collection := parts[1] // "closedtrades" or "opentrades"
	member := parts[2]     // "profit"

	return h.memberHandler.CanHandle(collection, member)
}

// GenerateCode delegates to TradeCollectionMemberHandler.GenerateAccess
func (h *TradeCollectionCallHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	// Extract nested member expression parts from CallExpression.Callee
	// Example: strategy.closedtrades.profit(5) has Callee = MemberExpression{
	//   Object: MemberExpression{Object: "strategy", Property: "closedtrades"},
	//   Property: "profit"
	// }

	// Quick check: is this even a member expression?
	callee, ok := call.Callee.(*ast.MemberExpression)
	if !ok {
		return "", nil // Not a member expression, not our job
	}

	// Extract property name (e.g., "profit")
	property, ok := callee.Property.(*ast.Identifier)
	if !ok {
		return "", nil // Property isn't an identifier, not our pattern
	}

	// Extract collection name from nested object (strategy.closedtrades → "closedtrades")
	collection := extractCollectionName(callee.Object)
	if collection == "" {
		return "", nil // Not a strategy.{collection} pattern, not our job
	}

	// Validate this is actually a trade collection call we can handle
	if !h.memberHandler.CanHandle(collection, property.Name) {
		return "", nil // Not a valid trade property, not our job
	}

	// Now we know it's ours - delegate to member handler
	exprGen := &generatorExpressionAdapter{g: g}
	return h.memberHandler.GenerateAccess(collection, property.Name, call.Arguments, exprGen)
}

// extractCollectionName extracts collection identifier from strategy.{collection} MemberExpression.
//
// Example: MemberExpression{Object: "strategy", Property: "closedtrades"} → "closedtrades"
//
// Returns: Empty string if not a valid strategy.{collection} pattern
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

// extractMemberExpressionPath recursively builds dotted path from nested MemberExpressions.
//
// Examples:
//   - MemberExpression{Object: "strategy", Property: "closedtrades"} → "strategy.closedtrades"
//   - MemberExpression{Object: MemberExpression{...}, Property: "profit"} → recursively processes
//
// Returns: Empty string if extraction fails
func extractMemberExpressionPath(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		objPath := extractMemberExpressionPath(e.Object)
		propName := extractIdentifierNameSafe(e.Property)
		if objPath != "" && propName != "" {
			return objPath + "." + propName
		}
	}
	return ""
}

// extractIdentifierNameSafe extracts name from Identifier, returns empty string otherwise.
func extractIdentifierNameSafe(expr ast.Expression) string {
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}

// generatorExpressionAdapter adapts generator to ExpressionGenerator interface.
//
// Purpose: Allows TradeCollectionMemberHandler to call g.generateArrowFunctionExpression
//
//	without depending on the full generator type.
type generatorExpressionAdapter struct {
	g *generator
}

// Generate delegates to generator's arrow function expression handler.
func (a *generatorExpressionAdapter) Generate(expr ast.Expression) (string, error) {
	return a.g.generateArrowFunctionExpression(expr)
}

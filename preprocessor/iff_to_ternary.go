package preprocessor

import (
	"fmt"

	"github.com/quant5-lab/runner/parser"
)

/* IffToTernaryTransformer converts iff(condition, consequent, alternate) → condition ? consequent : alternate
 *
 * Pine v4 iff() is deprecated in v5 - semantically identical to ternary operator
 * Transformation: CallExpr with callee="iff" → TernaryExpr
 */
type IffToTernaryTransformer struct{}

func NewIffToTernaryTransformer() *IffToTernaryTransformer {
	return &IffToTernaryTransformer{}
}

func (t *IffToTernaryTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	for _, stmt := range script.Statements {
		if err := t.visitStatement(stmt); err != nil {
			return nil, err
		}
	}
	return script, nil
}

func (t *IffToTernaryTransformer) visitStatement(stmt *parser.Statement) error {
	if stmt == nil || stmt.Core == nil {
		return nil
	}

	if stmt.Core.Assignment != nil {
		return t.visitExpression(stmt.Core.Assignment.Value)
	}

	if stmt.Core.TypedAssignment != nil {
		return t.visitExpression(stmt.Core.TypedAssignment.Value)
	}

	if stmt.Core.Reassignment != nil {
		return t.visitExpression(stmt.Core.Reassignment.Value)
	}

	if stmt.Core.If != nil {
		if err := t.visitOrExpr(stmt.Core.If.Condition); err != nil {
			return err
		}
		for _, bodyStmt := range stmt.Core.If.Body {
			if err := t.visitStatement(bodyStmt); err != nil {
				return err
			}
		}
	}

	if stmt.Core.Expression != nil {
		return t.visitExpression(stmt.Core.Expression.Expr)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitExpression(expr *parser.Expression) error {
	if expr == nil {
		return nil
	}

	if expr.Ternary != nil {
		if expr.Ternary.TrueVal == nil && expr.Ternary.FalseVal == nil {
			return t.transformConditionIfNeeded(expr)
		}
		return t.visitTernaryExpr(expr.Ternary)
	}

	if expr.Call != nil {
		return t.visitCallExprArgs(expr.Call)
	}

	if expr.Array != nil {
		for _, elem := range expr.Array.Elements {
			if err := t.visitTernaryExpr(elem); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *IffToTernaryTransformer) transformConditionIfNeeded(expr *parser.Expression) error {
	if expr.Ternary == nil || expr.Ternary.Condition == nil {
		return nil
	}

	call := t.findIffCallInOrExpr(expr.Ternary.Condition)
	if call == nil {
		return t.visitOrExpr(expr.Ternary.Condition)
	}

	ternary, err := t.convertIffCallToTernary(call)
	if err != nil {
		return err
	}

	expr.Ternary = ternary

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInOrExpr(orExpr *parser.OrExpr) *parser.CallExpr {
	if orExpr == nil {
		return nil
	}

	if call := t.findIffCallInAndExpr(orExpr.Left); call != nil {
		return call
	}

	if orExpr.Right != nil {
		return t.findIffCallInOrExpr(orExpr.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInAndExpr(andExpr *parser.AndExpr) *parser.CallExpr {
	if andExpr == nil {
		return nil
	}

	if call := t.findIffCallInCompExpr(andExpr.Left); call != nil {
		return call
	}

	if andExpr.Right != nil {
		return t.findIffCallInAndExpr(andExpr.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInCompExpr(compExpr *parser.CompExpr) *parser.CallExpr {
	if compExpr == nil {
		return nil
	}

	if call := t.findIffCallInArithExpr(compExpr.Left); call != nil {
		return call
	}

	if compExpr.Right != nil {
		return t.findIffCallInCompExpr(compExpr.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInArithExpr(arithExpr *parser.ArithExpr) *parser.CallExpr {
	if arithExpr == nil {
		return nil
	}

	if call := t.findIffCallInTerm(arithExpr.Left); call != nil {
		return call
	}

	if arithExpr.Right != nil {
		return t.findIffCallInArithExpr(arithExpr.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInTerm(term *parser.Term) *parser.CallExpr {
	if term == nil {
		return nil
	}

	if call := t.findIffCallInFactor(term.Left); call != nil {
		return call
	}

	if term.Right != nil {
		return t.findIffCallInTerm(term.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInFactor(factor *parser.Factor) *parser.CallExpr {
	if factor == nil {
		return nil
	}

	if factor.Postfix != nil {
		return t.findIffCallInPostfix(factor.Postfix)
	}

	return nil
}

func (t *IffToTernaryTransformer) findIffCallInPostfix(postfix *parser.PostfixExpr) *parser.CallExpr {
	if postfix == nil || postfix.Primary == nil {
		return nil
	}

	primary := postfix.Primary

	if primary.Call != nil && t.isIffCall(primary.Call) {
		return primary.Call
	}

	if primary.Paren != nil && primary.Paren.Call != nil && t.isIffCall(primary.Paren.Call) {
		return primary.Paren.Call
	}

	return nil
}

func (t *IffToTernaryTransformer) transformCallToTernary(expr *parser.Expression) error {
	call := expr.Call
	if call == nil || call.Callee == nil {
		return nil
	}

	if !t.isIffCall(call) {
		return t.visitCallExprArgs(call)
	}

	ternary, err := t.convertIffCallToTernary(call)
	if err != nil {
		return err
	}

	expr.Call = nil
	expr.Ternary = ternary

	return nil
}

func (t *IffToTernaryTransformer) isIffCall(call *parser.CallExpr) bool {
	if call.Callee.Ident == nil {
		return false
	}
	return *call.Callee.Ident == "iff"
}

func (t *IffToTernaryTransformer) convertIffCallToTernary(call *parser.CallExpr) (*parser.TernaryExpr, error) {
	if len(call.Args) != 3 {
		return nil, fmt.Errorf("iff() requires exactly 3 arguments, got %d", len(call.Args))
	}

	conditionExpr := call.Args[0].Value
	if conditionExpr == nil {
		return nil, fmt.Errorf("iff() condition argument is nil")
	}

	condition, err := t.extractOrExprFromExpression(conditionExpr)
	if err != nil {
		return nil, fmt.Errorf("iff() condition: %w", err)
	}

	consequent := call.Args[1].Value
	if consequent == nil {
		return nil, fmt.Errorf("iff() consequent argument is nil")
	}

	alternate := call.Args[2].Value
	if alternate == nil {
		return nil, fmt.Errorf("iff() alternate argument is nil")
	}

	if err := t.visitExpression(consequent); err != nil {
		return nil, err
	}

	if err := t.visitExpression(alternate); err != nil {
		return nil, err
	}

	return &parser.TernaryExpr{
		Condition: condition,
		TrueVal:   consequent,
		FalseVal:  alternate,
	}, nil
}

func (t *IffToTernaryTransformer) extractOrExprFromExpression(expr *parser.Expression) (*parser.OrExpr, error) {
	if expr.Ternary != nil {
		return t.convertTernaryToOrExpr(expr.Ternary)
	}

	compExpr := t.expressionToCompExpr(expr)
	return &parser.OrExpr{
		Left:  &parser.AndExpr{Left: compExpr},
		Right: nil,
	}, nil
}

func (t *IffToTernaryTransformer) convertTernaryToOrExpr(ternary *parser.TernaryExpr) (*parser.OrExpr, error) {
	if ternary.Condition != nil {
		return ternary.Condition, nil
	}
	return nil, fmt.Errorf("ternary condition is nil")
}

func (t *IffToTernaryTransformer) expressionToCompExpr(expr *parser.Expression) *parser.CompExpr {
	arithExpr := t.expressionToArithExpr(expr)
	return &parser.CompExpr{
		Left:  arithExpr,
		Op:    nil,
		Right: nil,
	}
}

func (t *IffToTernaryTransformer) expressionToArithExpr(expr *parser.Expression) *parser.ArithExpr {
	term := t.expressionToTerm(expr)
	return &parser.ArithExpr{
		Left:  term,
		Op:    nil,
		Right: nil,
	}
}

func (t *IffToTernaryTransformer) expressionToTerm(expr *parser.Expression) *parser.Term {
	factor := t.expressionToFactor(expr)
	return &parser.Term{
		Left:  factor,
		Op:    nil,
		Right: nil,
	}
}

func (t *IffToTernaryTransformer) expressionToFactor(expr *parser.Expression) *parser.Factor {
	factor := &parser.Factor{}

	if expr.Array != nil {
		factor.Array = expr.Array
	} else if expr.Ident != nil {
		factor.Ident = expr.Ident
	} else if expr.Number != nil {
		factor.Number = expr.Number
	} else if expr.String != nil {
		factor.String = expr.String
	} else if expr.Call != nil {
		factor.Postfix = &parser.PostfixExpr{
			Primary: &parser.PrimaryExpr{Call: expr.Call},
		}
	} else if expr.MemberAccess != nil {
		factor.MemberAccess = expr.MemberAccess
	}

	return factor
}

func (t *IffToTernaryTransformer) visitCallExprArgs(call *parser.CallExpr) error {
	for _, arg := range call.Args {
		if arg.Value != nil {
			if err := t.visitExpression(arg.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *IffToTernaryTransformer) visitTernaryExpr(ternary *parser.TernaryExpr) error {
	if ternary == nil {
		return nil
	}

	if err := t.visitOrExpr(ternary.Condition); err != nil {
		return err
	}

	if err := t.visitExpression(ternary.TrueVal); err != nil {
		return err
	}

	if err := t.visitExpression(ternary.FalseVal); err != nil {
		return err
	}

	return nil
}

func (t *IffToTernaryTransformer) visitOrExpr(or *parser.OrExpr) error {
	if or == nil {
		return nil
	}

	if err := t.visitAndExpr(or.Left); err != nil {
		return err
	}

	if or.Right != nil {
		return t.visitOrExpr(or.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitAndExpr(and *parser.AndExpr) error {
	if and == nil {
		return nil
	}

	if err := t.visitCompExpr(and.Left); err != nil {
		return err
	}

	if and.Right != nil {
		return t.visitAndExpr(and.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitCompExpr(comp *parser.CompExpr) error {
	if comp == nil {
		return nil
	}

	if err := t.visitArithExpr(comp.Left); err != nil {
		return err
	}

	if comp.Right != nil {
		return t.visitCompExpr(comp.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitArithExpr(arith *parser.ArithExpr) error {
	if arith == nil {
		return nil
	}

	if err := t.visitTerm(arith.Left); err != nil {
		return err
	}

	if arith.Right != nil {
		return t.visitArithExpr(arith.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitTerm(term *parser.Term) error {
	if term == nil {
		return nil
	}

	if err := t.visitFactor(term.Left); err != nil {
		return err
	}

	if term.Right != nil {
		return t.visitTerm(term.Right)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitFactor(factor *parser.Factor) error {
	if factor == nil {
		return nil
	}

	if factor.Array != nil {
		for _, elem := range factor.Array.Elements {
			if err := t.visitTernaryExpr(elem); err != nil {
				return err
			}
		}
	}

	if factor.Postfix != nil {
		return t.visitPostfixExpr(factor.Postfix)
	}

	return nil
}

func (t *IffToTernaryTransformer) visitPostfixExpr(postfix *parser.PostfixExpr) error {
	if postfix == nil {
		return nil
	}

	if postfix.Primary != nil {
		if postfix.Primary.Paren != nil {
			if err := t.visitExpression(postfix.Primary.Paren); err != nil {
				return err
			}
		}
		if postfix.Primary.Call != nil {
			if err := t.visitCallExprArgs(postfix.Primary.Call); err != nil {
				return err
			}
		}
	}

	if postfix.Subscript != nil {
		return t.visitArithExpr(postfix.Subscript)
	}

	return nil
}

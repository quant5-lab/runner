package codegen

type OperatorPrecedence int

const (
	PrecLowest OperatorPrecedence = iota
	PrecLogicalOr
	PrecLogicalAnd
	PrecEquality
	PrecComparison
	PrecAdditive
	PrecMultiplicative
	PrecUnary
	PrecHighest
)

var operatorPrecedenceMap = map[string]OperatorPrecedence{
	"||": PrecLogicalOr,
	"or": PrecLogicalOr,

	"&&":  PrecLogicalAnd,
	"and": PrecLogicalAnd,

	"==": PrecEquality,
	"!=": PrecEquality,

	"<":  PrecComparison,
	"<=": PrecComparison,
	">":  PrecComparison,
	">=": PrecComparison,

	"+": PrecAdditive,
	"-": PrecAdditive,

	"*": PrecMultiplicative,
	"/": PrecMultiplicative,
	"%": PrecMultiplicative,
}

func GetOperatorPrecedence(operator string) OperatorPrecedence {
	if prec, exists := operatorPrecedenceMap[operator]; exists {
		return prec
	}
	return PrecLowest
}

func NeedsParentheses(childOp string, parentOp string, isRightChild bool) bool {
	childPrec := GetOperatorPrecedence(childOp)
	parentPrec := GetOperatorPrecedence(parentOp)

	if childPrec < parentPrec {
		return true
	}

	if childPrec == parentPrec && isRightChild {
		return true
	}

	return false
}

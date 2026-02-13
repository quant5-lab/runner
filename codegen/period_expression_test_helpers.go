package codegen

func P(period int) PeriodExpression {
	return NewConstantPeriod(period)
}

func R(varName string) PeriodExpression {
	return NewRuntimePeriod(varName)
}

func C(goExpr string) PeriodExpression {
	return NewComputedPeriod(goExpr)
}

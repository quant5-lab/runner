package codegen

/* Test helper functions for PeriodExpression */

/* P wraps integer as ConstantPeriod - convenience for tests */
func P(period int) PeriodExpression {
	return NewConstantPeriod(period)
}

/* R creates RuntimePeriod - convenience for tests */
func R(varName string) PeriodExpression {
	return NewRuntimePeriod(varName)
}

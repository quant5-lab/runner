package codegen

/* securityExpressionArgument is the position of the evaluated expression in
 * security(symbol, timeframe, expression, ...) */
const securityExpressionArgument = 2

func IsSecurityFunction(funcName string) bool {
	return funcName == "security" || funcName == "request.security"
}

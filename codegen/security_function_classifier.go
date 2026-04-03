package codegen

func IsSecurityFunction(funcName string) bool {
	return funcName == "security" || funcName == "request.security"
}

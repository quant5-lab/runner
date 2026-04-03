package codegen

type StringFunctionRegistry struct {
	signatures map[string]StringFunctionSignature
}

func NewStringFunctionRegistry() *StringFunctionRegistry {
	registry := &StringFunctionRegistry{
		signatures: make(map[string]StringFunctionSignature),
	}
	registry.registerAll()
	return registry
}

func (r *StringFunctionRegistry) registerAll() {
	signatures := []StringFunctionSignature{
		NewStringFunctionSignature("str.tostring", 1, 2, "string"),
		NewStringFunctionSignature("str.tonumber", 1, 1, "float"),
		NewStringFunctionSignature("str.length", 1, 1, "int"),
		NewStringFunctionSignature("str.format", 1, -1, "string"),
		NewStringFunctionSignature("str.contains", 2, 2, "bool"),
		NewStringFunctionSignature("str.pos", 2, 2, "int"),
		NewStringFunctionSignature("str.substring", 2, 3, "string"),
		NewStringFunctionSignature("str.lower", 1, 1, "string"),
		NewStringFunctionSignature("str.upper", 1, 1, "string"),
		NewStringFunctionSignature("str.trim", 1, 1, "string"),
		NewStringFunctionSignature("str.replace", 3, 4, "string"),
		NewStringFunctionSignature("str.replace_all", 3, 3, "string"),
		NewStringFunctionSignature("str.split", 2, 2, "[]string"),
		NewStringFunctionSignature("str.startswith", 2, 2, "bool"),
		NewStringFunctionSignature("str.endswith", 2, 2, "bool"),
		NewStringFunctionSignature("str.repeat", 2, 3, "string"),
		NewStringFunctionSignature("str.match", 2, 2, "string"),
		NewStringFunctionSignature("str.format_time", 2, 3, "string"),
	}

	for _, sig := range signatures {
		r.signatures[sig.Name] = sig
	}
}

func (r *StringFunctionRegistry) GetSignature(funcName string) (StringFunctionSignature, bool) {
	sig, exists := r.signatures[funcName]
	return sig, exists
}

func (r *StringFunctionRegistry) IsRegistered(funcName string) bool {
	_, exists := r.signatures[funcName]
	return exists
}

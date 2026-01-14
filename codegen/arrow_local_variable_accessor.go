package codegen

/* ArrowLocalVariableAccessor resolves scalar vs series access for arrow function local variables */
type ArrowLocalVariableAccessor struct {
	localVars map[string]bool
}

func NewArrowLocalVariableAccessor() *ArrowLocalVariableAccessor {
	return &ArrowLocalVariableAccessor{
		localVars: make(map[string]bool),
	}
}

/* RegisterLocalVariable registers a variable as a local arrow function variable */
func (a *ArrowLocalVariableAccessor) RegisterLocalVariable(varName string) {
	a.localVars[varName] = true
}

/* IsLocalVariable checks if a variable is registered as a local variable */
func (a *ArrowLocalVariableAccessor) IsLocalVariable(varName string) bool {
	return a.localVars[varName]
}

/* GenerateAccess generates scalar (offset=0) or Series.Get(offset) access code */
func (a *ArrowLocalVariableAccessor) GenerateAccess(varName string, offset int) string {
	if offset == 0 {
		return varName
	}
	return varName + "Series.Get(" + itoa(offset) + ")"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	if negative {
		digits = append(digits, '-')
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}

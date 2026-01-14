package codegen

import "fmt"

type CodeGenerationLimits struct {
	MaxStatementsPerPass int
	MaxSecurityCalls     int
}

func NewCodeGenerationLimits() CodeGenerationLimits {
	return CodeGenerationLimits{
		MaxStatementsPerPass: 10000,
		MaxSecurityCalls:     100,
	}
}

type StatementCounter struct {
	count  int
	limits CodeGenerationLimits
}

func NewStatementCounter(limits CodeGenerationLimits) *StatementCounter {
	return &StatementCounter{
		count:  0,
		limits: limits,
	}
}

func (sc *StatementCounter) Increment() error {
	sc.count++
	if sc.count > sc.limits.MaxStatementsPerPass {
		return fmt.Errorf("exceeded maximum statement limit (%d) - possible infinite loop", sc.limits.MaxStatementsPerPass)
	}
	return nil
}

func (sc *StatementCounter) Reset() {
	sc.count = 0
}

func (sc *StatementCounter) Count() int {
	return sc.count
}

type SecurityCallValidator struct {
	limits CodeGenerationLimits
}

func NewSecurityCallValidator(limits CodeGenerationLimits) *SecurityCallValidator {
	return &SecurityCallValidator{limits: limits}
}

func (scv *SecurityCallValidator) ValidateCallCount(actualCalls int) error {
	if actualCalls > scv.limits.MaxSecurityCalls {
		return fmt.Errorf("exceeded maximum security() calls (%d) - possible infinite loop or resource exhaustion", scv.limits.MaxSecurityCalls)
	}
	return nil
}

type RuntimeSafetyGuard struct {
	MaxBarsPerExecution int
}

func NewRuntimeSafetyGuard() RuntimeSafetyGuard {
	return RuntimeSafetyGuard{
		MaxBarsPerExecution: 1000000,
	}
}

func (rsg RuntimeSafetyGuard) GenerateBarCountValidation() string {
	return fmt.Sprintf(`const maxBars = %d
barCount := len(ctx.Data)
if barCount > maxBars {
	fmt.Fprintf(os.Stderr, "Error: bar count (%%d) exceeds safety limit (%%d)\n", barCount, maxBars)
	os.Exit(1)
}`, rsg.MaxBarsPerExecution)
}

func (rsg RuntimeSafetyGuard) GenerateIterationVariableReference() string {
	return "i"
}

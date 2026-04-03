package codegen

import "fmt"

type StringFunctionSignature struct {
	Name       string
	MinArgs    int
	MaxArgs    int
	ReturnType string
}

type StringCodeGenerator interface {
	CanGenerate(funcName string) bool
	Generate(g *generator, funcName string, args []string) (string, error)
}

func NewStringFunctionSignature(name string, minArgs, maxArgs int, returnType string) StringFunctionSignature {
	return StringFunctionSignature{
		Name:       name,
		MinArgs:    minArgs,
		MaxArgs:    maxArgs,
		ReturnType: returnType,
	}
}

func (s StringFunctionSignature) ValidateArgCount(actual int) error {
	if s.MaxArgs == -1 {
		if actual < s.MinArgs {
			return fmt.Errorf("%s: expected at least %d arguments, got %d", s.Name, s.MinArgs, actual)
		}
		return nil
	}
	if actual < s.MinArgs || actual > s.MaxArgs {
		if s.MinArgs == s.MaxArgs {
			return fmt.Errorf("%s: expected %d arguments, got %d", s.Name, s.MinArgs, actual)
		}
		return fmt.Errorf("%s: expected %d-%d arguments, got %d", s.Name, s.MinArgs, s.MaxArgs, actual)
	}
	return nil
}

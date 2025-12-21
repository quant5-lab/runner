package codegen

import "fmt"

/* ArrowContextScope represents a call site requiring ArrowContext wrapping */
type ArrowContextScope struct {
	FunctionName   string
	ContextVarName string
	ResultVarNames []string
	ArgumentList   string
}

/* ArrowContextWrapperGenerator produces ArrowContext creation and cleanup code for call sites */
type ArrowContextWrapperGenerator struct {
	indentation string
}

func NewArrowContextWrapperGenerator(indent string) *ArrowContextWrapperGenerator {
	return &ArrowContextWrapperGenerator{
		indentation: indent,
	}
}

func (g *ArrowContextWrapperGenerator) GenerateWrapper(scope ArrowContextScope) string {
	code := ""

	code += g.indentation + fmt.Sprintf("%s := context.NewArrowContext(ctx)\n", scope.ContextVarName)

	resultAssignment := g.buildResultAssignment(scope.ResultVarNames)
	code += g.indentation + fmt.Sprintf("%s := %s(%s, %s)\n",
		resultAssignment,
		scope.FunctionName,
		scope.ContextVarName,
		scope.ArgumentList,
	)

	code += g.indentation + fmt.Sprintf("%s.AdvanceAll()\n", scope.ContextVarName)

	return code
}

func (g *ArrowContextWrapperGenerator) buildResultAssignment(varNames []string) string {
	if len(varNames) == 0 {
		return "_"
	}
	if len(varNames) == 1 {
		return varNames[0]
	}

	assignment := ""
	for i, name := range varNames {
		if i > 0 {
			assignment += ", "
		}
		assignment += name
	}
	return assignment
}

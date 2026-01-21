package codegen

type VariableOperationType int

const (
	VariableDeclaration VariableOperationType = iota
	VariableReassignment
)

func (t VariableOperationType) GoAssignmentOperator() string {
	if t == VariableReassignment {
		return "="
	}
	return ":="
}

func OperationTypeFromASTKind(kind string) VariableOperationType {
	if kind == "var" {
		return VariableReassignment
	}
	return VariableDeclaration
}

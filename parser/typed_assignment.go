package parser

type TypedAssignment struct {
	TypeHint string      `parser:"@('float' | 'int' | 'bool' | 'string' | 'color')"`
	Name     string      `parser:"@Ident '='"`
	Value    *Expression `parser:"@@"`
}

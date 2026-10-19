package parser

// TypedAssignment is a bare (non-var) typed declaration, e.g. `float x = expr`
// or its array form `float[] x = array.new_float()`.
// Drawing types (line, label, box, table, linefill, polyline) degrade to NaN.
type TypedAssignment struct {
	TypeHint string      `parser:"@('float' | 'int' | 'bool' | 'string' | 'color' | 'line' | 'label' | 'box' | 'table' | 'linefill' | 'polyline') ('[' ']')?"`
	Name     string      `parser:"@Ident '='"`
	Value    *Expression `parser:"@@"`
}

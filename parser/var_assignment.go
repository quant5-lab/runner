package parser

/*
VarAssignment represents Pine Script var/varip prefixed variable declarations.

Pine Script semantics:
  - var: Initialize once on bar 0, persist across bars. On bars > 0, the declaration
    is skipped and the variable retains its value from the previous bar.
  - varip: "var intrabar persist" — identical to var in historical-only mode.
    In realtime mode (not yet supported), varip values survive intrabar recalculations
    while var values roll back to the bar's opening value on each tick.

Syntax variants:
  - var x = expr                        (untyped scalar)
  - var float x = expr                  (typed scalar)
  - var float[] x = array.new_float()   (typed scalar array)
  - varip x = expr                      (untyped scalar, intrabar persist)
  - varip float x = expr                (typed scalar, intrabar persist)
  - var [a, b] = expr                   (tuple destructuring)
  - varip [a, b] = expr                 (tuple destructuring, intrabar persist)
  - var line[] x = array.new_line()     (typed drawing-object array, display-only)
  - var label[] x = array.new_label()   (typed drawing-object array, display-only)
  - var box[] x = array.new_box()       (typed drawing-object array, display-only)

Note: drawing types (line, label, box, table, linefill, polyline) are ONLY treated as
TypeHints when followed by [] — otherwise they parse as regular variable names.
This avoids ambiguity with common variable names like `var label = "initial"`.
Scalar typed arrays (float[], int[], bool[], string[], color[]) are accepted with
or without [] suffix; the RHS expression determines the actual runtime behavior.
*/
type VarAssignment struct {
	Modifier   string      `parser:"@('var' | 'varip')"`
	TypeHint   *string     `parser:"( @('line' | 'label' | 'box' | 'table' | 'linefill' | 'polyline' | 'float' | 'int' | 'bool' | 'string' | 'color') '[' ']' | @('float' | 'int' | 'bool' | 'string' | 'color') )?"`
	TupleNames []string    `parser:"( '[' @Ident ( ',' @Ident )* ']'"`
	Name       *string     `parser:"| @Ident ) '='"`
	Value      *Expression `parser:"@@"`
}

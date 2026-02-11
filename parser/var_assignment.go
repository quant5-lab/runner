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
  - var x = expr           (untyped scalar)
  - var float x = expr     (typed scalar)
  - varip x = expr         (untyped scalar, intrabar persist)
  - varip float x = expr   (typed scalar, intrabar persist)
  - var [a, b] = expr      (tuple destructuring)
  - varip [a, b] = expr    (tuple destructuring, intrabar persist)
*/
type VarAssignment struct {
	Modifier   string      `parser:"@('var' | 'varip')"`
	TypeHint   *string     `parser:"@('float' | 'int' | 'bool' | 'string' | 'color')?"`
	TupleNames []string    `parser:"( '[' @Ident ( ',' @Ident )* ']'"`
	Name       *string     `parser:"| @Ident ) '='"`
	Value      *Expression `parser:"@@"`
}

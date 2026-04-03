package parser

type InlineStatementList struct {
	Statements      []*InlineStatement `parser:"@@+"`
	FinalExpression *TernaryExpr       `parser:"@@"`
}

type InlineStatement struct {
	Name  string       `parser:"(?= @Ident ( '=' | ':=' ) ) @Ident"`
	Op    string       `parser:"@( '=' | ':=' )"`
	Value *TernaryExpr `parser:"@@ ','"`
}

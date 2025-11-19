package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Script struct {
	Version    *VersionDirective `parser:"@@?"`
	Statements []*Statement      `parser:"@@*"`
}

type VersionDirective struct {
	Value int `parser:"Comment"`
}

type Statement struct {
	Assignment *Assignment     `parser:"@@"`
	If         *IfStatement    `parser:"| @@"`
	Expression *ExpressionStmt `parser:"| @@"`
}

type IfStatement struct {
	Condition *Comparison `parser:"'if' ( '(' @@ ')' | @@ )"`
	Body      *Statement  `parser:"@@"`
}

type Assignment struct {
	Name  string      `parser:"@Ident '='"`
	Value *Expression `parser:"@@"`
}

type ExpressionStmt struct {
	Expr *Expression `parser:"@@"`
}

type Expression struct {
	Ternary      *TernaryExpr  `parser:"@@"`
	Call         *CallExpr     `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

type TernaryExpr struct {
	Condition *OrExpr     `parser:"@@"`
	TrueVal   *Expression `parser:"( '?' @@"`
	FalseVal  *Expression `parser:"':' @@ )?"`
}

type OrExpr struct {
	Left  *AndExpr `parser:"@@"`
	Right *OrExpr  `parser:"( ( 'or' | '||' ) @@ )?"`
}

type AndExpr struct {
	Left  *CompExpr `parser:"@@"`
	Right *AndExpr  `parser:"( ( 'and' | '&&' ) @@ )?"`
}

type CompExpr struct {
	Left  *ArithExpr `parser:"@@"`
	Op    *string    `parser:"( @( '>' | '<' | '>=' | '<=' | '==' | '!=' )"`
	Right *CompExpr  `parser:"@@ )?"`
}

type ArithExpr struct {
	Left  *Term      `parser:"@@"`
	Op    *string    `parser:"( @( '+' | '-' )"`
	Right *ArithExpr `parser:"@@ )?"`
}

type Term struct {
	Left  *Factor `parser:"@@"`
	Op    *string `parser:"( @( '*' | '/' | '%' )"`
	Right *Term   `parser:"@@ )?"`
}

type Factor struct {
	Paren        *ArithExpr    `parser:"( '(' @@ ')' )"`
	Call         *CallExpr     `parser:"| @@"`
	Subscript    *Subscript    `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Boolean      *bool         `parser:"| ( @'true' | @'false' )"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

type Subscript struct {
	Object string     `parser:"@Ident"`
	Index  *ArithExpr `parser:"'[' @@ ']'"`
}

type Comparison struct {
	Left  *ComparisonTerm `parser:"@@"`
	Op    *string         `parser:"( @( '>' | '<' | '>=' | '<=' | '==' | '!=' | 'and' | 'or' )"`
	Right *ComparisonTerm `parser:"@@ )?"`
}

type ComparisonTerm struct {
	Call         *CallExpr     `parser:"@@"`
	Subscript    *Subscript    `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Boolean      *bool         `parser:"| ( @'true' | @'false' )"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

type MemberAccess struct {
	Object   string `parser:"@Ident"`
	Property string `parser:"'.' @Ident"`
}

type CallExpr struct {
	Callee *CallCallee `parser:"@@"`
	Args   []*Argument `parser:"'(' ( @@ ( ',' @@ )* )? ')'"`
}

type CallCallee struct {
	MemberAccess *MemberAccess `parser:"@@"`
	Ident        *string       `parser:"| @Ident"`
}

type Argument struct {
	Name  *string      `parser:"( @Ident '=' )?"`
	Value *TernaryExpr `parser:"@@"`
}

type Value struct {
	CallExpr     *CallExpr     `parser:"@@"`
	Subscript    *Subscript    `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Boolean      *bool         `parser:"| ( @'true' | @'false' )"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

var pineLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Comment", Pattern: `//[^\n]*`},
	{Name: "Whitespace", Pattern: `[ \t\r\n]+`},
	{Name: "String", Pattern: `"[^"]*"|'[^']*'`},
	{Name: "Float", Pattern: `\d+\.\d+`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "Punct", Pattern: `==|!=|>=|<=|&&|\|\||[(),=@/.><!?:+\-*%\[\]]`},
})

func NewParser() (*participle.Parser[Script], error) {
	return participle.Build[Script](
		participle.Lexer(pineLexer),
		participle.Elide("Comment", "Whitespace"),
		participle.UseLookahead(4),
	)
}

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
	Expression *ExpressionStmt `parser:"| @@"`
}

type Assignment struct {
	Name  string      `parser:"@Ident '='"`
	Value *Expression `parser:"@@"`
}

type ExpressionStmt struct {
	Expr *Expression `parser:"@@"`
}

type Expression struct {
	Call         *CallExpr     `parser:"@@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

type MemberAccess struct {
	Object   string `parser:"@Ident '.'"`
	Property string `parser:"@Ident"`
}

type CallExpr struct {
	Namespace *string     `parser:"( @Ident '.' )?"`
	Function  string      `parser:"@Ident"`
	Args      []*Argument `parser:"'(' ( @@ ( ',' @@ )* )? ')'"`
}

type Argument struct {
	Name  *string `parser:"( @Ident '=' )?"`
	Value *Value  `parser:"@@"`
}

type Value struct {
	Member  *MemberAccess `parser:"@@"`
	Boolean *string       `parser:"| ( @'true' | @'false' )"`
	Ident   *string       `parser:"| @Ident"`
	Number  *float64      `parser:"| ( @Float | @Int )"`
	String  *string       `parser:"| @String"`
}

var pineLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Comment", Pattern: `//[^\n]*`},
	{Name: "Whitespace", Pattern: `[ \t\r\n]+`},
	{Name: "String", Pattern: `"[^"]*"`},
	{Name: "Float", Pattern: `\d+\.\d+`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "Punct", Pattern: `[(),=@/.]`},
})

func NewParser() (*participle.Parser[Script], error) {
	return participle.Build[Script](
		participle.Lexer(pineLexer),
		participle.Elide("Comment", "Whitespace"),
	)
}

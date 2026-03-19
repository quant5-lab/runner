package parser

import (
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	indentlexer "github.com/quant5-lab/runner/lexer"
)

type Script struct {
	Version    *VersionDirective `parser:"@@?"`
	Statements []*Statement      `parser:"@@*"`
}

type VersionDirective struct {
	Value int `parser:"Comment"`
}

type Statement struct {
	Core          *StatementCore `parser:"@@"`
	TrailingComma *string        `parser:"@','?"`
}

type StatementCore struct {
	TupleAssignment *TupleAssignment `parser:"@@"`
	If              *IfStatement     `parser:"| @@"`
	ForIn           *ForInStatement  `parser:"| @@"`
	For             *ForStatement    `parser:"| @@"`
	While           *WhileStatement  `parser:"| @@"`
	Switch          *SwitchExpr      `parser:"| @@"`
	FunctionDecl    *FunctionDecl    `parser:"| @@"`
	VarAssignment   *VarAssignment   `parser:"| @@"`
	TypedAssignment *TypedAssignment `parser:"| @@"`
	Assignment      *Assignment      `parser:"| @@"`
	Reassignment    *Reassignment    `parser:"| @@"`
	Break           *BreakStmt       `parser:"| @@"`
	Continue        *ContinueStmt    `parser:"| @@"`
	Expression      *ExpressionStmt  `parser:"| @@"`
}

type IfStatement struct {
	Condition  *OrExpr      `parser:"'if' @@"`
	Indent     *string      `parser:"@Indent"`
	Body       []*Statement `parser:"@@+"`
	Dedent     *string      `parser:"@Dedent"`
	ElseClause *ElseClause  `parser:"( @@ )?"`
}

type ElseClause struct {
	ElseIf   *IfStatement `parser:"'else' ( @@"`
	ElseBody []*Statement `parser:"| Indent @@+ Dedent )"`
}

type ForStatement struct {
	Counter string       `parser:"'for' @Ident '='"`
	From    *ArithExpr   `parser:"@@"`
	To      *ArithExpr   `parser:"'to' @@"`
	Step    *ArithExpr   `parser:"( 'by' @@ )?"`
	Indent  *string      `parser:"@Indent"`
	Body    []*Statement `parser:"@@+"`
	Dedent  *string      `parser:"@Dedent"`
}

type ForInVars struct {
	TupleIndex    *string `parser:"'[' @Ident ','"`
	TupleElement  *string `parser:"@Ident ']'"`
	SingleElement *string `parser:"| @Ident"`
}

type ForInStatement struct {
	Vars       *ForInVars   `parser:"'for' @@"`
	Collection *ArithExpr   `parser:"'in' @@"`
	Indent     *string      `parser:"@Indent"`
	Body       []*Statement `parser:"@@+"`
	Dedent     *string      `parser:"@Dedent"`
}

type WhileStatement struct {
	Condition *OrExpr      `parser:"'while' @@"`
	Indent    *string      `parser:"@Indent"`
	Body      []*Statement `parser:"@@+"`
	Dedent    *string      `parser:"@Dedent"`
}

type FunctionDecl struct {
	Name                string               `parser:"@Ident"`
	Params              []string             `parser:"'(' ( @Ident ( ',' @Ident )* )? ')'"`
	Arrow               string               `parser:"@'=>'"`
	MultiLineIndent     *string              `parser:"( Newline? @Indent"`
	MultiLineBody       []*Statement         `parser:"@@+"`
	MultiLineDedent     *string              `parser:"@Dedent"`
	InlineStatementList *InlineStatementList `parser:"| Newline? @@"`
	InlineBody          *Expression          `parser:"| Newline? @@ )"`
}

type TupleAssignment struct {
	Names []string    `parser:"'[' @Ident ( ',' @Ident )* ']'"`
	Eq    *string     `parser:"( @'=' )?"`
	Value *Expression `parser:"@@?"`
}

type Assignment struct {
	Name  string      `parser:"@Ident '='"`
	Value *Expression `parser:"@@"`
}

type Reassignment struct {
	Name  string      `parser:"@Ident ':='"`
	Value *Expression `parser:"@@"`
}

type ExpressionStmt struct {
	Expr *Expression `parser:"@@"`
}

type BreakStmt struct {
	Keyword string `parser:"@'break'"`
}

type ContinueStmt struct {
	Keyword string `parser:"@'continue'"`
}

type ArrayLiteral struct {
	Elements []*TernaryExpr `parser:"'[' ( @@ ( ',' @@ )* )? ']'"`
}

type ForExpr struct {
	Counter string       `parser:"'for' @Ident '='"`
	From    *ArithExpr   `parser:"@@"`
	To      *ArithExpr   `parser:"'to' @@"`
	Step    *ArithExpr   `parser:"( 'by' @@ )?"`
	Indent  *string      `parser:"@Indent"`
	Body    []*Statement `parser:"@@+"`
	Dedent  *string      `parser:"@Dedent"`
}

type ForInExpr struct {
	Vars       *ForInVars   `parser:"'for' @@"`
	Collection *ArithExpr   `parser:"'in' @@"`
	Indent     *string      `parser:"@Indent"`
	Body       []*Statement `parser:"@@+"`
	Dedent     *string      `parser:"@Dedent"`
}

type IfExpr struct {
	Condition  *OrExpr      `parser:"'if' @@"`
	Indent     *string      `parser:"@Indent"`
	Body       []*Statement `parser:"@@+"`
	Dedent     *string      `parser:"@Dedent"`
	ElseClause *ElseClause  `parser:"( @@ )?"`
}

type SwitchExpr struct {
	Subject *OrExpr       `parser:"'switch' @@?"`
	Indent  *string       `parser:"@Indent"`
	Cases   []*SwitchCase `parser:"@@*"`
	Dedent  *string       `parser:"@Dedent"`
}

type SwitchCase struct {
	Condition  *OrExpr      `parser:"@@? '=>'"`
	Indent     *string      `parser:"( @Indent"`
	Body       []*Statement `parser:"@@+"`
	Dedent     *string      `parser:"@Dedent"`
	InlineBody *Expression  `parser:"| @@ )"`
}

type WhileExpr struct {
	Condition *OrExpr      `parser:"'while' @@"`
	Indent    *string      `parser:"@Indent"`
	Body      []*Statement `parser:"@@+"`
	Dedent    *string      `parser:"@Dedent"`
}

type Expression struct {
	ForInExpr    *ForInExpr    `parser:"@@"`
	ForExpr      *ForExpr      `parser:"| @@"`
	WhileExpr    *WhileExpr    `parser:"| @@"`
	IfExpr       *IfExpr       `parser:"| @@"`
	SwitchExpr   *SwitchExpr   `parser:"| @@"`
	Ternary      *TernaryExpr  `parser:"| @@"`
	Array        *ArrayLiteral `parser:"| @@"`
	Call         *CallExpr     `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
	HexColor     *string       `parser:"| @HexColor"`
}

type TernaryExpr struct {
	Condition *OrExpr     `parser:"@@"`
	TrueVal   *Expression `parser:"( '?' ( Newline | Indent | Dedent )* @@"`
	FalseVal  *Expression `parser:"( Newline | Indent | Dedent )* ':' ( Newline | Indent | Dedent )* @@ )?"`
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
	Array        *ArrayLiteral `parser:"@@"`
	Unary        *UnaryExpr    `parser:"| @@"`
	True         *string       `parser:"| @'true'"`
	False        *string       `parser:"| @'false'"`
	Postfix      *PostfixExpr  `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
	HexColor     *string       `parser:"| @HexColor"`
}

type PostfixExpr struct {
	Primary   *PrimaryExpr `parser:"@@"`
	Subscript *ArithExpr   `parser:"( '[' @@ ']' )?"`
}

type PrimaryExpr struct {
	Paren        *Expression   `parser:"'(' @@ ')'"`
	Call         *CallExpr     `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
}

type UnaryExpr struct {
	Op      string  `parser:"@( '-' | '+' | 'not' | '!' )"`
	Operand *Factor `parser:"@@"`
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
	True         *string       `parser:"@'true'"`
	False        *string       `parser:"| @'false'"`
	Postfix      *PostfixExpr  `parser:"| @@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
}

type MemberAccess struct {
	Object     string   `parser:"@Ident"`
	Properties []string `parser:"( '.' @Ident )+"`
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
	Name  *string     `parser:"( @Ident '=' )?"`
	Value *Expression `parser:"@@"`
}

type Value struct {
	Postfix      *PostfixExpr  `parser:"@@"`
	MemberAccess *MemberAccess `parser:"| @@"`
	True         *string       `parser:"| @'true'"`
	False        *string       `parser:"| @'false'"`
	Ident        *string       `parser:"| @Ident"`
	Number       *float64      `parser:"| ( @Float | @Int )"`
	String       *string       `parser:"| @String"`
	HexColor     *string       `parser:"| @HexColor"`
}

var pineLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Comment", Pattern: `//[^\n]*`},
	{Name: "Newline", Pattern: `\r?\n`},
	{Name: "Whitespace", Pattern: `[ \t]+`},
	{Name: "Keyword", Pattern: `\b(if|for|in|to|by|while|switch|and|or|not|true|false|break|continue|var|varip)\b`},
	{Name: "String", Pattern: `"[^"]*"|'[^']*'`},
	{Name: "HexColor", Pattern: `#[0-9A-Fa-f]{6}([0-9A-Fa-f]{2})?`},
	{Name: "Float", Pattern: `\d+[eE][+-]?\d+|\d*\.\d+([eE][+-]?\d+)?|\d+\.([eE][+-]?\d+)?`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "Punct", Pattern: `:=|=>|==|!=|>=|<=|&&|\|\||[(),=@/.><!?:+\-*%\[\]]`},
})

var indentAwareLexer = indentlexer.NewIndentationDefinition(pineLexer)

/* Singleton — grammar is static, concurrent Build() on shared lexer definition races */
var (
	cachedParser    *participle.Parser[Script]
	cachedParserErr error
	parserOnce      sync.Once
)

func NewParser() (*participle.Parser[Script], error) {
	parserOnce.Do(func() {
		cachedParser, cachedParserErr = participle.Build[Script](
			participle.Lexer(indentAwareLexer),
			participle.Elide("Comment", "Whitespace", "Newline"),
			participle.UseLookahead(16),
		)
	})
	return cachedParser, cachedParserErr
}

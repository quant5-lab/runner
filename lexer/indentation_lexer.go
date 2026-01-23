package lexer

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

type IndentationDefinition struct {
	base lexer.Definition
}

func NewIndentationDefinition(base lexer.Definition) *IndentationDefinition {
	return &IndentationDefinition{base: base}
}

func (d *IndentationDefinition) Symbols() map[string]lexer.TokenType {
	symbols := d.base.Symbols()
	nextType := lexer.TokenType(len(symbols) + 1)
	symbols["Indent"] = nextType
	symbols["Dedent"] = nextType + 1
	symbols["Newline"] = nextType + 2
	return symbols
}

func (d *IndentationDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	baseLexer, err := d.base.Lex(filename, r)
	if err != nil {
		return nil, err
	}
	return NewIndentationLexer(baseLexer, d.Symbols()), nil
}

type IndentationLexer struct {
	base            lexer.Lexer
	symbols         map[string]lexer.TokenType
	indentStack     []int
	pending         []lexer.Token
	previousLine    int
	atLineStart     bool
	whitespaceType  lexer.TokenType
	lastTokenValue  string
	expectingIndent bool // Set to true after => or if
	inTernary       bool
	parenDepth      int
}

func NewIndentationLexer(base lexer.Lexer, symbols map[string]lexer.TokenType) *IndentationLexer {
	return &IndentationLexer{
		base:            base,
		symbols:         symbols,
		indentStack:     []int{0},
		pending:         []lexer.Token{},
		previousLine:    0,
		atLineStart:     true,
		whitespaceType:  symbols["Whitespace"],
		lastTokenValue:  "",
		expectingIndent: false,
		inTernary:       false,
		parenDepth:      0,
	}
}

func (l *IndentationLexer) Next() (lexer.Token, error) {
	for {
		if len(l.pending) > 0 {
			token := l.pending[0]
			l.pending = l.pending[1:]
			if token.Type != l.symbols["Indent"] && token.Type != l.symbols["Dedent"] {
				l.lastTokenValue = token.Value
				if l.isControlFlowKeyword(token.Value) {
					l.expectingIndent = true
				}
			}
			return token, nil
		}

		token, err := l.base.Next()
		if err != nil {
			return token, err
		}

		if token.Type == lexer.EOF {
			for len(l.indentStack) > 1 {
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
				l.pending = append(l.pending, lexer.Token{
					Type: l.symbols["Dedent"],
					Pos:  token.Pos,
				})
			}
			l.pending = append(l.pending, token)
			continue
		}

		// Skip whitespace but track lines
		if token.Type == l.whitespaceType {
			continue
		}

		if newlineType, exists := l.symbols["Newline"]; exists && token.Type == newlineType {
			continue
		}

		// Skip comments - don't process their indentation
		if commentType, exists := l.symbols["Comment"]; exists && token.Type == commentType {
			return token, nil
		}

		tokenValue := token.Value

		if l.isControlFlowKeyword(tokenValue) {
			l.expectingIndent = true
		}

		l.lastTokenValue = tokenValue

		if token.Pos.Line > l.previousLine {
			l.previousLine = token.Pos.Line
			indent := token.Pos.Column - 1
			currentIndent := l.indentStack[len(l.indentStack)-1]

			// Only emit INDENT if we're expecting it
			if l.expectingIndent && indent > currentIndent {
				l.indentStack = append(l.indentStack, indent)
				l.pending = append(l.pending, lexer.Token{
					Type: l.symbols["Indent"],
					Pos:  token.Pos,
				})
				l.pending = append(l.pending, token)
				l.expectingIndent = false
				// Re-check if the pending token itself requires indentation
				if l.isControlFlowKeyword(token.Value) {
					l.expectingIndent = true
				}
				continue
			}

			if indent < currentIndent {
				for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
					l.indentStack = l.indentStack[:len(l.indentStack)-1]
					l.pending = append(l.pending, lexer.Token{
						Type: l.symbols["Dedent"],
						Pos:  token.Pos,
					})
				}
				l.pending = append(l.pending, token)
				continue
			}
		}

		return token, nil
	}
}

func (l *IndentationLexer) isControlFlowKeyword(value string) bool {
	return value == "=>" || value == "if" || value == "for" || value == "while"
}

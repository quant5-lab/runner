package lexer

import (
	"io"
	"sync"

	"github.com/alecthomas/participle/v2/lexer"
)

type IndentationDefinition struct {
	base        lexer.Definition
	symbolsOnce sync.Once
	symbolsMap  map[string]lexer.TokenType
}

func NewIndentationDefinition(base lexer.Definition) *IndentationDefinition {
	return &IndentationDefinition{base: base}
}

func (d *IndentationDefinition) Symbols() map[string]lexer.TokenType {
	d.symbolsOnce.Do(func() {
		/* Mutate base map in-place — StatefulLexer.Next() reads token types from it */
		d.symbolsMap = d.base.Symbols()
		nextType := lexer.TokenType(len(d.symbolsMap) + 1)
		d.symbolsMap["Indent"] = nextType
		d.symbolsMap["Dedent"] = nextType + 1
		d.symbolsMap["Newline"] = nextType + 2
	})
	return d.symbolsMap
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
	expectingIndent bool
}

func NewIndentationLexer(base lexer.Lexer, symbols map[string]lexer.TokenType) *IndentationLexer {
	return &IndentationLexer{
		base:           base,
		symbols:        symbols,
		indentStack:    []int{0},
		pending:        []lexer.Token{},
		whitespaceType: symbols["Whitespace"],
	}
}

func (l *IndentationLexer) Next() (lexer.Token, error) {
	for {
		if len(l.pending) > 0 {
			token := l.pending[0]
			l.pending = l.pending[1:]
			if token.Type != l.symbols["Indent"] && token.Type != l.symbols["Dedent"] {
				l.lastTokenValue = token.Value
				if isControlFlowKeyword(token.Value) {
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

		if token.Type == l.whitespaceType {
			continue
		}

		if newlineType, exists := l.symbols["Newline"]; exists && token.Type == newlineType {
			continue
		}

		if commentType, exists := l.symbols["Comment"]; exists && token.Type == commentType {
			return token, nil
		}

		tokenValue := token.Value

		previousTokenValue := l.lastTokenValue

		if isControlFlowKeyword(tokenValue) {
			l.expectingIndent = true
		}

		l.lastTokenValue = tokenValue

		if token.Pos.Line > l.previousLine {
			l.previousLine = token.Pos.Line
			indent := token.Pos.Column - 1
			currentIndent := l.indentStack[len(l.indentStack)-1]

			// Resolve block expectation at the line boundary: a keyword whose body
			// sits on the same line never produces indent > current, so the flag
			// must not carry forward past that line.
			wasExpectingIndent := l.expectingIndent
			l.expectingIndent = isControlFlowKeyword(tokenValue)

			if wasExpectingIndent && indent > currentIndent {
				if isContinuationOperator(previousTokenValue) {
					// Block body comes after the continuation; keep expectingIndent
					// so the body line claims the INDENT.
					l.expectingIndent = true
					return token, nil
				}

				l.indentStack = append(l.indentStack, indent)
				l.pending = append(l.pending, lexer.Token{
					Type: l.symbols["Indent"],
					Pos:  token.Pos,
				})
				l.pending = append(l.pending, token)
				l.expectingIndent = isControlFlowKeyword(token.Value)
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

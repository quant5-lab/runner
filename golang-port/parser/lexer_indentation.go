package parser

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

type IndentationLexer struct {
	underlying   lexer.Lexer
	buffer       []lexer.Token
	indentStack  []int
	atLineStart  bool
	pendingToken *lexer.Token
}

type IndentationDefinition struct {
	underlying lexer.Definition
}

func NewIndentationDefinition(underlying lexer.Definition) *IndentationDefinition {
	return &IndentationDefinition{underlying: underlying}
}

func (d *IndentationDefinition) Symbols() map[string]lexer.TokenType {
	symbols := d.underlying.Symbols()
	nextType := lexer.TokenType(len(symbols) + 1)
	symbols["INDENT"] = nextType
	symbols["DEDENT"] = nextType + 1
	symbols["NEWLINE"] = nextType + 2
	return symbols
}

func (d *IndentationDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	underlyingLexer, err := d.underlying.Lex(filename, r)
	if err != nil {
		return nil, err
	}
	return NewIndentationLexer(underlyingLexer, d.Symbols()), nil
}

func NewIndentationLexer(underlying lexer.Lexer, symbols map[string]lexer.TokenType) *IndentationLexer {
	return &IndentationLexer{
		underlying:  underlying,
		buffer:      []lexer.Token{},
		indentStack: []int{0},
		atLineStart: true,
	}
}

func (l *IndentationLexer) Next() (lexer.Token, error) {
	if len(l.buffer) > 0 {
		token := l.buffer[0]
		l.buffer = l.buffer[1:]
		return token, nil
	}

	if l.pendingToken != nil {
		token := *l.pendingToken
		l.pendingToken = nil
		return token, nil
	}

	token, err := l.underlying.Next()
	if err != nil {
		if err == io.EOF {
			return l.handleEOF()
		}
		return token, err
	}

	if l.isWhitespaceToken(token) {
		if l.atLineStart {
			return l.handleIndentation(token)
		}
		return l.Next()
	}

	if l.isNewlineToken(token) {
		l.atLineStart = true
		return l.emitNewline(token.Pos), nil
	}

	if l.atLineStart {
		l.atLineStart = false
		currentIndent := l.indentStack[len(l.indentStack)-1]
		if currentIndent > 0 {
			dedents := l.generateDedents(0, token.Pos)
			if len(dedents) > 0 {
				l.buffer = append(dedents, token)
				return l.Next()
			}
		}
	}

	return token, nil
}

func (l *IndentationLexer) handleIndentation(wsToken lexer.Token) (lexer.Token, error) {
	indentLevel := l.calculateIndentLevel(wsToken.Value)
	currentIndent := l.indentStack[len(l.indentStack)-1]

	nextToken, err := l.underlying.Next()
	if err != nil {
		if err == io.EOF {
			return l.handleEOF()
		}
		return lexer.Token{}, err
	}

	if l.isNewlineToken(nextToken) {
		l.atLineStart = true
		return l.emitNewline(nextToken.Pos), nil
	}

	l.atLineStart = false

	if indentLevel > currentIndent {
		l.indentStack = append(l.indentStack, indentLevel)
		l.pendingToken = &nextToken
		return l.emitIndent(wsToken.Pos), nil
	}

	if indentLevel < currentIndent {
		dedents := l.generateDedents(indentLevel, wsToken.Pos)
		l.pendingToken = &nextToken
		if len(dedents) > 0 {
			l.buffer = dedents[1:]
			return dedents[0], nil
		}
	}

	return nextToken, nil
}

func (l *IndentationLexer) calculateIndentLevel(whitespace string) int {
	level := 0
	for _, ch := range whitespace {
		if ch == ' ' {
			level++
		} else if ch == '\t' {
			level += 4
		}
	}
	return level
}

func (l *IndentationLexer) generateDedents(targetIndent int, pos lexer.Position) []lexer.Token {
	var dedents []lexer.Token
	for len(l.indentStack) > 0 && l.indentStack[len(l.indentStack)-1] > targetIndent {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		dedents = append(dedents, l.emitDedent(pos))
	}
	return dedents
}

func (l *IndentationLexer) handleEOF() (lexer.Token, error) {
	if len(l.indentStack) > 1 {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		return l.emitDedent(lexer.Position{}), nil
	}
	return lexer.Token{Type: lexer.EOF}, io.EOF
}

func (l *IndentationLexer) isWhitespaceToken(token lexer.Token) bool {
	if token.Value == "" {
		return false
	}
	for _, ch := range token.Value {
		if ch != ' ' && ch != '\t' {
			return false
		}
	}
	return len(token.Value) > 0
}

func (l *IndentationLexer) isNewlineToken(token lexer.Token) bool {
	return token.Value == "\n" || token.Value == "\r\n"
}

func (l *IndentationLexer) emitIndent(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  lexer.TokenType(-1),
		Value: "INDENT",
		Pos:   pos,
	}
}

func (l *IndentationLexer) emitDedent(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  lexer.TokenType(-2),
		Value: "DEDENT",
		Pos:   pos,
	}
}

func (l *IndentationLexer) emitNewline(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  lexer.TokenType(-3),
		Value: "NEWLINE",
		Pos:   pos,
	}
}

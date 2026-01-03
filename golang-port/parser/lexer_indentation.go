package parser

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/participle/v2/lexer"
)

type IndentationLexer struct {
	underlying      lexer.Lexer
	buffer          []lexer.Token
	indentStack     []int
	atLineStart     bool
	pendingToken    *lexer.Token
	lastToken       lexer.Token
	expectingIndent bool
	indentType      lexer.TokenType
	dedentType      lexer.TokenType
	newlineType     lexer.TokenType
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
	symbols["Indent"] = nextType
	symbols["Dedent"] = nextType + 1
	symbols["Newline"] = nextType + 2
	return symbols
}

func (d *IndentationDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	underlyingLexer, err := d.underlying.Lex(filename, r)
	if err != nil {
		return nil, err
	}
	symbols := d.Symbols()
	return NewIndentationLexer(underlyingLexer, symbols), nil
}

func NewIndentationLexer(underlying lexer.Lexer, symbols map[string]lexer.TokenType) *IndentationLexer {
	return &IndentationLexer{
		underlying:      underlying,
		buffer:          []lexer.Token{},
		indentStack:     []int{0},
		atLineStart:     true,
		expectingIndent: false,
		indentType:      symbols["Indent"],
		dedentType:      symbols["Dedent"],
		newlineType:     symbols["Newline"],
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

	if token.Pos.Line >= 3 {
		fmt.Printf("LEXER_DEBUG L%d:C%d type=%d val=%q\n",
			token.Pos.Line, token.Pos.Column, token.Type, token.Value)
	}

	if l.isWhitespaceToken(token) {
		if token.Pos.Line == 4 {
			fmt.Fprintf(os.Stderr, "INDENT_DEBUG line 4 whitespace: atLineStart=%v value=%q\n", l.atLineStart, token.Value)
		}
		if l.atLineStart {
			if token.Pos.Line == 4 {
				fmt.Fprintf(os.Stderr, "INDENT_DEBUG calling handleIndentation for line 4\n")
			}
			return l.handleIndentation(token)
		}
		return l.Next()
	}

	if l.isNewlineToken(token) {
		l.atLineStart = true
		// expectingIndent is set when we see keywords like 'if', not here
		l.lastToken = token
		return l.emitNewline(token.Pos), nil
	}

	if l.atLineStart {
		l.atLineStart = false
		currentIndent := l.indentStack[len(l.indentStack)-1]
		if currentIndent > 0 {
			dedents := l.generateDedents(0, token.Pos)
			if len(dedents) > 0 {
				l.buffer = append(dedents, token)
				l.lastToken = token
				return l.Next()
			}
		}
	}

	// Track keywords that require indented blocks
	if l.shouldExpectIndent(token) {
		l.expectingIndent = true
	}

	l.lastToken = token
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

	// Debug for problematic lines
	if wsToken.Pos.Line == 4 {
		fmt.Fprintf(os.Stderr, "INDENT_DEBUG L%d: indentLevel=%d currentIndent=%d expectingIndent=%v nextToken=%q\n",
			wsToken.Pos.Line, indentLevel, currentIndent, l.expectingIndent, nextToken.Value)
	}

	if l.isNewlineToken(nextToken) {
		l.atLineStart = true
		return l.emitNewline(nextToken.Pos), nil
	}

	l.atLineStart = false

	// TradingView allows ±1 space tolerance within a block
	isWithinTolerance := currentIndent > 0 &&
		indentLevel >= currentIndent-1 &&
		indentLevel <= currentIndent+1

	if indentLevel > currentIndent && !isWithinTolerance {
		l.indentStack = append(l.indentStack, indentLevel)
		l.pendingToken = &nextToken
		l.expectingIndent = false
		return l.emitIndent(wsToken.Pos), nil
	}

	if (indentLevel == currentIndent || isWithinTolerance) && l.expectingIndent {
		l.indentStack = append(l.indentStack, indentLevel)
		l.pendingToken = &nextToken
		l.expectingIndent = false
		return l.emitIndent(wsToken.Pos), nil
	}

	// If within tolerance, treat as same level (no INDENT/DEDENT)
	if isWithinTolerance {
		l.pendingToken = &nextToken
		l.expectingIndent = false
		return nextToken, nil
	}

	if indentLevel < currentIndent {
		dedents := l.generateDedents(indentLevel, wsToken.Pos)
		l.pendingToken = &nextToken
		l.expectingIndent = false
		if len(dedents) > 0 {
			l.buffer = dedents[1:]
			return dedents[0], nil
		}
	}

	l.expectingIndent = false
	if wsToken.Pos.Line >= 340 && wsToken.Pos.Line <= 345 {
		fmt.Fprintf(os.Stderr, "DEBUG returning nextToken at line %d: %q\n", wsToken.Pos.Line, nextToken.Value)
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

func (l *IndentationLexer) shouldExpectIndent(token lexer.Token) bool {
	return token.Value == "if" || token.Value == "for" || token.Value == "while" ||
		token.Value == "=>" || token.Value == ":"
}

func (l *IndentationLexer) emitIndent(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  l.indentType,
		Value: "INDENT",
		Pos:   pos,
	}
}

func (l *IndentationLexer) emitDedent(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  l.dedentType,
		Value: "DEDENT",
		Pos:   pos,
	}
}

func (l *IndentationLexer) emitNewline(pos lexer.Position) lexer.Token {
	return lexer.Token{
		Type:  l.newlineType,
		Value: "NEWLINE",
		Pos:   pos,
	}
}

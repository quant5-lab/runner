package lexer

// isControlFlowKeyword reports whether the token value introduces a new indented block.
func isControlFlowKeyword(value string) bool {
	switch value {
	case "=>", "if", "else", "for", "while", "switch":
		return true
	}
	return false
}

// isContinuationOperator reports whether a token at the end of a line means the next
// line continues the same expression. The lexer suppresses INDENT emission in that case
// so the token that actually opens a block can claim the INDENT instead.
func isContinuationOperator(value string) bool {
	switch value {
	case "or", "and", "||", "&&",
		"+", "-", "*", "/", "%",
		">", "<", ">=", "<=", "==", "!=",
		",", "?", ":",
		"(", "[":
		return true
	}
	return false
}

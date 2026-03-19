package lexer

import "testing"

func TestIsControlFlowKeyword(t *testing.T) {
	keywords := []string{"=>", "if", "else", "for", "while", "switch"}
	for _, kw := range keywords {
		if !isControlFlowKeyword(kw) {
			t.Errorf("isControlFlowKeyword(%q) = false, want true", kw)
		}
	}

	nonKeywords := []string{
		"or", "and", "+", ":", "?", ":=",
		"x", "1", "true", "false", "",
		"IF", "Else", "FOR",
	}
	for _, v := range nonKeywords {
		if isControlFlowKeyword(v) {
			t.Errorf("isControlFlowKeyword(%q) = true, want false", v)
		}
	}
}

func TestIsContinuationOperator(t *testing.T) {
	operators := []string{
		"or", "and", "||", "&&",
		"+", "-", "*", "/", "%",
		">", "<", ">=", "<=", "==", "!=",
		",", "?", ":",
		"(", "[",
	}
	for _, op := range operators {
		if !isContinuationOperator(op) {
			t.Errorf("isContinuationOperator(%q) = false, want true", op)
		}
	}

	nonOperators := []string{
		")", "]",
		"if", "else", "for", "while", "switch", "=>",
		":=", "=",
		"x", "close", "1", "true", "false", "",
		"OR", "AND",
	}
	for _, v := range nonOperators {
		if isContinuationOperator(v) {
			t.Errorf("isContinuationOperator(%q) = true, want false", v)
		}
	}
}

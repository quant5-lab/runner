package preprocessor

import "strings"

type LineInfo struct {
	Raw       string
	Trimmed   string
	Indent    int
	IsEmpty   bool
	IsComment bool
}

func scanLine(line string) LineInfo {
	trimmed := strings.TrimSpace(line)
	return LineInfo{
		Raw:       line,
		Trimmed:   trimmed,
		Indent:    getIndentation(line),
		IsEmpty:   trimmed == "",
		IsComment: strings.HasPrefix(trimmed, "//"),
	}
}

func shouldSkipLine(info LineInfo) bool {
	return info.IsEmpty || info.IsComment
}

func isBodyEnd(currentIndent, baseIndent int) bool {
	return currentIndent <= baseIndent
}

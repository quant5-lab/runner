package preprocessor

import (
	"regexp"
	"strings"
)

var arrowFunctionPattern = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)\s*=>`)

type FunctionDeclInfo struct {
	Name       string
	ParamsText string
	FullHeader string
}

func matchFunctionDecl(trimmed string) (FunctionDeclInfo, bool) {
	matches := arrowFunctionPattern.FindStringSubmatch(trimmed)
	if matches == nil {
		return FunctionDeclInfo{}, false
	}

	return FunctionDeclInfo{
		Name:       matches[1],
		ParamsText: matches[2],
		FullHeader: trimmed,
	}, true
}

func NormalizeFunctionBlocks(script string) string {
	lines := strings.Split(script, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		lineInfo := scanLine(line)

		funcInfo, isFunc := matchFunctionDecl(lineInfo.Trimmed)
		if !isFunc {
			result = append(result, line)
			i++
			continue
		}

		baseIndent := lineInfo.Indent
		indentStr := strings.Repeat(" ", baseIndent)
		i++

		var bodyStatements []string
		for i < len(lines) {
			nextInfo := scanLine(lines[i])

			if shouldSkipLine(nextInfo) {
				i++
				continue
			}

			if isBodyEnd(nextInfo.Indent, baseIndent) {
				break
			}

			bodyStatements = append(bodyStatements, nextInfo.Trimmed)
			i++
		}

		if len(bodyStatements) == 0 {
			result = append(result, line)
			continue
		}

		result = append(result, indentStr+funcInfo.FullHeader+" @BEGIN")
		for _, stmt := range bodyStatements {
			result = append(result, indentStr+"    "+stmt)
		}
		result = append(result, indentStr+"@END")
	}

	return strings.Join(result, "\n")
}

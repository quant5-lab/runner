package preprocessor

import (
	"bufio"
	"regexp"
	"strings"
)

var arrayConstructorPattern = regexp.MustCompile(`\barray\.new<(float|int|bool|color|string)>`)
var matrixConstructorPattern = regexp.MustCompile(`\bmatrix\.new<(float|int|bool|color|string)>`)
var mapConstructorPattern = regexp.MustCompile(`\bmap\.new<([a-z]+)\s*,\s*([a-z]+)>`)

func RewriteGenericTypeSyntax(source string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(source))

	for scanner.Scan() {
		line := scanner.Text()
		rewritten := rewriteLine(line)
		result.WriteString(rewritten)
		result.WriteString("\n")
	}

	output := result.String()
	if len(output) > 0 && !strings.HasSuffix(source, "\n") {
		output = strings.TrimSuffix(output, "\n")
	}

	return output
}

func rewriteLine(line string) string {
	if isCommentLine(line) {
		return line
	}

	codeBeforeComment, comment := splitLineAtComment(line)

	rewritten := rewriteArrayConstructors(codeBeforeComment)
	rewritten = rewriteMatrixConstructors(rewritten)
	rewritten = rewriteMapConstructors(rewritten)

	if comment != "" {
		return rewritten + comment
	}
	return rewritten
}

func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "//")
}

func splitLineAtComment(line string) (code, comment string) {
	inString := false
	escapeNext := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if ch == '\\' && inString {
			escapeNext = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if !inString && i < len(line)-1 {
			if ch == '/' && line[i+1] == '*' {
				return line[:i], line[i:]
			}
			if ch == '/' && line[i+1] == '/' {
				return line[:i], line[i:]
			}
		}
	}

	return line, ""
}

func rewriteArrayConstructors(source string) string {
	return arrayConstructorPattern.ReplaceAllStringFunc(source, func(match string) string {
		submatch := arrayConstructorPattern.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return match
		}
		elemType := submatch[1]
		return "array.new_" + elemType
	})
}

func rewriteMatrixConstructors(source string) string {
	return matrixConstructorPattern.ReplaceAllStringFunc(source, func(match string) string {
		submatch := matrixConstructorPattern.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return match
		}
		elemType := submatch[1]
		return "matrix.new_" + elemType
	})
}

func rewriteMapConstructors(source string) string {
	return mapConstructorPattern.ReplaceAllStringFunc(source, func(match string) string {
		submatch := mapConstructorPattern.FindStringSubmatch(match)
		if len(submatch) != 3 {
			return match
		}
		keyType := submatch[1]
		valueType := submatch[2]
		return "map.new_" + keyType + "_" + valueType
	})
}

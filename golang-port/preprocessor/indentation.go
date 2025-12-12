package preprocessor

import (
	"regexp"
	"strings"
)

/* Normalize indented if statement blocks to single-line format for parser */
func NormalizeIfBlocks(script string) string {
	lines := strings.Split(script, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check if line is if statement
		if strings.HasPrefix(trimmed, "if ") {
			condition := strings.TrimPrefix(trimmed, "if ")
			indent := getIndentation(line)
			indentStr := strings.Repeat(" ", indent)
			i++

			// Collect multi-line condition (indented continuations before body)
			for i < len(lines) {
				nextLine := lines[i]
				nextIndent := getIndentation(nextLine)
				nextTrimmed := strings.TrimSpace(nextLine)

				// Empty line - skip
				if nextTrimmed == "" {
					i++
					continue
				}

				// Comment - skip
				if strings.HasPrefix(nextTrimmed, "//") {
					i++
					continue
				}

				// Indented line that looks like condition continuation (not body statement)
				if nextIndent > indent && !looksLikeBodyStatement(nextTrimmed) {
					condition += " " + nextTrimmed
					i++
					continue
				}

				// Body statement or same/less indent - end of condition
				break
			}

			// Collect body statements (next indented lines)
			var bodyStatements []string
			for i < len(lines) {
				nextLine := lines[i]
				nextIndent := getIndentation(nextLine)
				nextTrimmed := strings.TrimSpace(nextLine)

				// Empty line - preserve but don't end body collection
				if nextTrimmed == "" {
					i++
					continue
				}

				// Comment - preserve
				if strings.HasPrefix(nextTrimmed, "//") {
					i++
					continue
				}

				// Body statement (more indented than if)
				if nextIndent > indent {
					bodyStatements = append(bodyStatements, nextTrimmed)
					i++
					continue
				}

				// Same or less indent - end of body
				break
			}

			// Generate single-line if statement for each body statement
			// Use newline to separate condition from body for parser
			for _, stmt := range bodyStatements {
				result = append(result, indentStr+"if "+condition)
				result = append(result, indentStr+"    "+stmt)
			}
			continue
		}

		// Non-if line - keep as is
		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

func getIndentation(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Treat tab as 4 spaces
		} else {
			break
		}
	}
	return count
}

func looksLikeBodyStatement(trimmed string) bool {
	// Body statements typically start with: strategy., plot(, identifiers with assignment/calls
	return strings.HasPrefix(trimmed, "strategy.") ||
		strings.HasPrefix(trimmed, "plot(") ||
		regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\s*[:=]`).MatchString(trimmed) ||
		regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\(`).MatchString(trimmed)
}

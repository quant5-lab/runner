package preprocessor

import "strings"

// ExpandTabs normalises source text so the indentation-aware lexer sees
// consistent column numbers regardless of the file's origin platform.
func ExpandTabs(source string) string {
	source = strings.ReplaceAll(source, "\r", "")
	return strings.ReplaceAll(source, "\t", "    ")
}

package preprocessor

import "strings"

/* Tab expansion for lexer column consistency */
func ExpandTabs(source string) string {
	return strings.ReplaceAll(source, "\t", "    ")
}

package codegen

import (
	"fmt"
	"strings"
)

/*
ConditionalWrapperGenerator wraps code in conditional block when condition present.
Pure function following SRP - single responsibility for if-block generation.
*/
type ConditionalWrapperGenerator struct{}

/*
WrapIfNeeded wraps body code in if-block when condition present.
Returns original body if condition empty (backward compatibility).
*/
func (c *ConditionalWrapperGenerator) WrapIfNeeded(condition string, bodyCode string, indent string) string {
	if condition == "" {
		return bodyCode
	}

	lines := strings.Split(strings.TrimRight(bodyCode, "\n"), "\n")
	wrapped := indent + fmt.Sprintf("if value.IsTrue(%s) {\n", condition)

	for _, line := range lines {
		if line != "" {
			wrapped += "\t" + line + "\n"
		}
	}

	wrapped += indent + "}\n"
	return wrapped
}

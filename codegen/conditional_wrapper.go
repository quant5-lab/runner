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
WrapIfNeeded wraps body code in an if-block when condition present.

The condition must already be a Go boolean expression, which is what
ArgumentExtractor.ExtractWhenCondition yields; this generator owns block structure only.
Returns original body if condition empty (backward compatibility).
*/
func (c *ConditionalWrapperGenerator) WrapIfNeeded(condition string, bodyCode string, indent string) string {
	if condition == "" {
		return bodyCode
	}

	lines := strings.Split(strings.TrimRight(bodyCode, "\n"), "\n")
	wrapped := indent + fmt.Sprintf("if %s {\n", condition)

	for _, line := range lines {
		if line != "" {
			wrapped += "\t" + line + "\n"
		}
	}

	wrapped += indent + "}\n"
	return wrapped
}

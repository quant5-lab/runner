package integration

import (
	"strings"
	"testing"
)

/* ParseGeneratedFilePath extracts the generated Go file path from pine-gen output.
 * Pine-gen now creates unique temp files to support parallel test execution.
 */
func ParseGeneratedFilePath(t *testing.T, pineGenOutput []byte) string {
	t.Helper()

	outputStr := string(pineGenOutput)
	genPrefix := "Generated: "
	startIdx := strings.Index(outputStr, genPrefix)
	if startIdx == -1 {
		t.Fatalf("Could not find 'Generated: ' in pine-gen output: %s", outputStr)
	}
	startIdx += len(genPrefix)
	endIdx := strings.Index(outputStr[startIdx:], "\n")
	if endIdx == -1 {
		endIdx = len(outputStr)
	} else {
		endIdx += startIdx
	}
	return outputStr[startIdx:endIdx]
}

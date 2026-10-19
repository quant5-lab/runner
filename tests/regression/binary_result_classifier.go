package regression

import (
	"fmt"
	"strings"
	"testing"
)

// securityFetchErrorPrefix must stay in sync with the "Failed to fetch" literal
// in codegen/security_inject.go — both sides detect the same stderr line.
const securityFetchErrorPrefix = "Failed to fetch "

// binaryOutcome is the disposition of a single binary execution in the smoke test.
type binaryOutcome int

const (
	outcomePass  binaryOutcome = iota // clean run or recovered runtime error (exit 0/2)
	outcomeSkip                       // security() fixture absent — result is uninformative
	outcomeFatal                      // unrecovered panic or unexpected binary failure
)

// binaryResult carries the classified outcome and a human-readable reason.
type binaryResult struct {
	outcome binaryOutcome
	reason  string
}

// isSecurityFixtureMissing reports whether the binary exited because a
// security() fixture file was not found in the data directory.
func isSecurityFixtureMissing(stderr string) bool {
	return strings.Contains(stderr, securityFetchErrorPrefix)
}

// containsRawPanic reports whether stderr carries an unrecovered goroutine
// dump.  Errors caught by defer/recover emit "Runtime error: …" without a
// stack trace, so the two cases are distinguishable.
func containsRawPanic(stderr string) bool {
	return strings.Contains(stderr, "goroutine ") || strings.Contains(stderr, "panic:")
}

// classifyBinaryResult is the single authoritative decision point for how the
// smoke test reacts to a binary execution outcome.  Raw panics take precedence
// over exit codes because a broken recover() must never be silently accepted.
func classifyBinaryResult(exitCode int, stderr string) binaryResult {
	if containsRawPanic(stderr) {
		return binaryResult{outcomeFatal, "raw panic in stderr (defer/recover not working):\n" + stderr}
	}
	switch exitCode {
	case 0, 2:
		return binaryResult{outcome: outcomePass}
	case 1:
		if isSecurityFixtureMissing(stderr) {
			return binaryResult{outcomeSkip,
				"security() fixture unavailable in smoke test environment: " + strings.TrimRight(stderr, "\n")}
		}
		return binaryResult{outcomeFatal, "binary exited 1 (unexpected failure):\n" + stderr}
	default:
		return binaryResult{outcomeFatal, fmt.Sprintf("binary exited %d (unexpected exit code):\n%s", exitCode, stderr)}
	}
}

// handleBinaryResult applies classifyBinaryResult as t.Skip / t.Fatal / pass.
func handleBinaryResult(t *testing.T, exitCode int, stderr string) {
	t.Helper()
	result := classifyBinaryResult(exitCode, stderr)
	switch result.outcome {
	case outcomePass:
	case outcomeSkip:
		t.Skip(result.reason)
	case outcomeFatal:
		t.Fatal(result.reason)
	}
}

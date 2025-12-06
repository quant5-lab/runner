package clock

import (
	"os"
	"strings"
	"time"
)

// Now is the function used across the codebase to get current time.
// In test mode (when running `go test`), this automatically returns a fixed
// deterministic time (2020-09-13 12:26:40 UTC) to ensure reproducible tests.
// Tests can override this by calling Set() for specific time scenarios.
var Now = defaultNow

// defaultNow returns the current time, or a fixed deterministic time during tests.
// This makes all tests deterministic by default without requiring explicit clock.Set() calls.
func defaultNow() time.Time {
	// Detect test mode: go test sets a unique temp directory in GOCACHE or we can check for test binary
	// Most reliable: check if we're running under 'go test' via the presence of test flags
	if isTestMode() {
		// Return fixed epoch: 2020-09-13 12:26:40 UTC (Unix: 1600000000)
		return time.Unix(1600000000, 0)
	}
	return time.Now()
}

// isTestMode detects if code is running under 'go test'.
// We check for test binary name patterns that go test creates.
func isTestMode() bool {
	if len(os.Args) == 0 {
		return false
	}

	// go test creates binaries with .test extension or contains .test. in the path
	binaryName := os.Args[0]

	// Check for .test suffix (Linux/Mac test binaries)
	if strings.HasSuffix(binaryName, ".test") {
		return true
	}

	// Check for .test. in path (temporary test binaries)
	if strings.Contains(binaryName, ".test.") {
		return true
	}

	// Check for test flags in any position
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}

	return false
}

// Set replaces the Now function and returns a restore function which
// restores the previous Now when called. Use this in tests that need
// specific timestamps different from the default deterministic time.
func Set(f func() time.Time) func() {
	prev := Now
	Now = f
	return func() { Now = prev }
}

package regression

import (
	"strings"
	"testing"
)

func TestContainsRawPanic(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		want   bool
	}{
		{"goroutine dump first line", "goroutine 1 [running]:\nmain.main()\n", true},
		{"panic colon prefix", "panic: runtime error: invalid memory address or nil pointer dereference\n", true},
		{"goroutine and panic both present", "panic: nil pointer\n\ngoroutine 1 [running]:\n", true},
		{"goroutine pattern mid-output", "startup ok\ngoroutine 2 [chan receive]:\n", true},
		{"recovered error prefix", "Runtime error: index out of range\n", false},
		{"normal strategy output", "Symbol: TEST\nNet profit: 1234.56\nTotal trades: 12\n", false},
		{"security fixture missing", "Failed to fetch TEST:1D: open /tmp/x.json: no such file or directory\n", false},
		{"empty stderr", "", false},
		{"goroutine without trailing space", "goroutinedump text\n", false},
		{"panic without colon", "panicking due to out-of-memory\n", false},
		{"panic word only", "panic\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := containsRawPanic(tc.stderr); got != tc.want {
				t.Errorf("containsRawPanic(%q) = %v, want %v", tc.stderr, got, tc.want)
			}
		})
	}
}

func TestIsSecurityFixtureMissing(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		want   bool
	}{
		{"symbol and timeframe present", "Failed to fetch TEST:1D: open /tmp/x/TEST_1D.json: no such file\n", true},
		{"different symbol and timeframe", "Failed to fetch EURUSD:60: open /data/EURUSD_60.json: no such file\n", true},
		{"prefix at start with content after", "Failed to fetch SYM:D: some OS error\n", true},
		{"prefix in middle of multi-line output", "init ok\nFailed to fetch SYM:1D: err\nshutdown\n", true},
		{"empty stderr", "", false},
		{"recovered runtime error", "Runtime error: index out of range\n", false},
		{"goroutine dump", "goroutine 1 [running]:\nmain.main()\n", false},
		{"different data-file error", "Failed to read data file: open /tmp/data.json: no such file\n", false},
		{"normal strategy output", "Symbol: TEST\nNet profit: 1000.00\n", false},
		{"fetch without trailing space or continuation", "Failed to fetch\n", false},
		{"fetch followed immediately by non-space character", "Failed to fetchX:1D: err\n", false},
		{"lowercase prefix", "failed to fetch TEST:1D: err\n", false},
		{"uppercase prefix", "FAILED TO FETCH TEST:1D: err\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSecurityFixtureMissing(tc.stderr); got != tc.want {
				t.Errorf("isSecurityFixtureMissing(%q) = %v, want %v", tc.stderr, got, tc.want)
			}
		})
	}
}

// TestClassifyBinaryResult covers the complete decision table for binary
// execution outcomes.  The panic-override row exists because a broken
// recover() must never be silently accepted regardless of exit code.
func TestClassifyBinaryResult(t *testing.T) {
	cases := []struct {
		name        string
		exitCode    int
		stderr      string
		wantOutcome binaryOutcome
		wantReason  bool
	}{
		{"exit 0 empty stderr", 0, "", outcomePass, false},
		{"exit 0 normal output", 0, "Symbol: TEST\nNet profit: 0.00\n", outcomePass, false},
		{"exit 2 empty stderr", 2, "", outcomePass, false},
		{"exit 2 recovered runtime error", 2, "Runtime error: index out of range\n", outcomePass, false},
		{"exit 1 fixture missing", 1, "Failed to fetch TEST:1D: open /tmp/TEST_1D.json: no such file\n", outcomeSkip, true},
		{"exit 1 fixture missing alternate symbol", 1, "Failed to fetch EURUSD:60: open /data/EURUSD_60.json: no such file\n", outcomeSkip, true},
		{"exit 1 unrelated error", 1, "Failed to read data file: open /tmp/data.json: no such file\n", outcomeFatal, true},
		{"exit 1 empty stderr", 1, "", outcomeFatal, true},
		{"exit 1 generic message", 1, "something went wrong\n", outcomeFatal, true},
		{"exit 3", 3, "", outcomeFatal, true},
		{"exit 127 command not found", 127, "command not found\n", outcomeFatal, true},
		{"exit -1 process error", -1, "", outcomeFatal, true},
		{"exit 0 with goroutine dump", 0, "goroutine 1 [running]:\nmain.main()\n", outcomeFatal, true},
		{"exit 2 with panic colon", 2, "panic: nil pointer dereference\n", outcomeFatal, true},
		{"exit 1 fixture miss and goroutine dump", 1, "goroutine 1 [running]:\nFailed to fetch TEST:1D: err\n", outcomeFatal, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyBinaryResult(tc.exitCode, tc.stderr)
			if got.outcome != tc.wantOutcome {
				t.Errorf("classifyBinaryResult(%d, %q).outcome = %v, want %v",
					tc.exitCode, tc.stderr, got.outcome, tc.wantOutcome)
			}
			if tc.wantReason && got.reason == "" {
				t.Errorf("classifyBinaryResult(%d, %q).reason is empty, want non-empty", tc.exitCode, tc.stderr)
			}
			if !tc.wantReason && got.reason != "" {
				t.Errorf("classifyBinaryResult(%d, %q).reason = %q, want empty", tc.exitCode, tc.stderr, got.reason)
			}
		})
	}
}

// TestClassifyBinaryResult_SkipReasonContainsStderr verifies that the skip
// reason propagates stderr so a missing fixture is identifiable in the test log.
func TestClassifyBinaryResult_SkipReasonContainsStderr(t *testing.T) {
	const line = "Failed to fetch SBERP:1D: open /data/SBERP_1D.json: no such file or directory"
	result := classifyBinaryResult(1, line+"\n")
	if result.outcome != outcomeSkip {
		t.Fatalf("expected outcomeSkip, got %v", result.outcome)
	}
	if !strings.Contains(result.reason, line) {
		t.Errorf("skip reason %q does not contain stderr line %q", result.reason, line)
	}
}

// TestClassifyBinaryResult_FatalReasonContainsStderr verifies that the fatal
// reason propagates stderr so unexpected exit-1 failures are diagnosable.
func TestClassifyBinaryResult_FatalReasonContainsStderr(t *testing.T) {
	const line = "some unexpected binary error"
	result := classifyBinaryResult(1, line+"\n")
	if result.outcome != outcomeFatal {
		t.Fatalf("expected outcomeFatal, got %v", result.outcome)
	}
	if !strings.Contains(result.reason, line) {
		t.Errorf("fatal reason %q does not contain stderr line %q", result.reason, line)
	}
}

// TestContainsRawPanic_MutualExclusionWithFixtureMiss guards against a fixture-
// miss stderr line triggering the panic detector, since both feed classifyBinaryResult.
func TestContainsRawPanic_MutualExclusionWithFixtureMiss(t *testing.T) {
	stderr := "Failed to fetch TEST:1D: open /tmp/TEST_1D.json: no such file or directory\n"
	if containsRawPanic(stderr) {
		t.Error("fixture-missing stderr must not trigger containsRawPanic")
	}
	if !isSecurityFixtureMissing(stderr) {
		t.Error("fixture-missing stderr must be detected by isSecurityFixtureMissing")
	}
}

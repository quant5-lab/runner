package regression

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

/*
TestCodegenPineVersionThreading_Ratchet ratchets the
generator-side half of the BB7 phantom-shorts fix: the emitted Go source
for a v4 strategy must bind the Pine version literal '4' as the trailing
argument to session.TimeFuncWithVersion, and the corresponding v5 fixture
must bind '5'. If a future refactor accidentally strips the trailing
version argument or hardcodes a version, this test fails before the
weekend-entry ratchet would ever fire.

Why not the version-aware session call directly? Because the codegen
emission is the link between Program.PineVersion and the runtime session
filter. A bug at that seam (e.g. always emitting 5) cannot be caught at
the runtime/session layer.
*/
func TestCodegenPineVersionThreading_Ratchet(t *testing.T) {
	root := projectRootFromCwd()

	cases := []struct {
		name        string
		pineSource  string
		expectArg   int
		unexpectArg int
	}{
		{
			name: "v4_strategy_emits_TimeFuncWithVersion_4",
			pineSource: `//@version=4
strategy("v4 session ratchet", overlay=true)
trading_session = input("0950-1645", title="Session", type=input.session)
in_sess = na(time(timeframe.period, trading_session)) ? false : true
if (in_sess)
    strategy.entry("L", strategy.long)
`,
			expectArg:   4,
			unexpectArg: 5,
		},
		{
			name: "v5_strategy_emits_TimeFuncWithVersion_5",
			pineSource: `//@version=5
strategy("v5 session ratchet", overlay=true)
trading_session = input.session("0950-1645", title="Session")
in_sess = not na(time(timeframe.period, trading_session))
if in_sess
    strategy.entry("L", strategy.long)
`,
			expectArg:   5,
			unexpectArg: 4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			built, ok := codegenAndBuild(t, tmpDir, "version_threading", tc.pineSource, root)
			if !ok {
				t.Fatal("codegen/build failed for fixture")
			}

			generated, err := os.ReadFile(built.GeneratedPath)
			if err != nil {
				t.Fatalf("read generated Go source: %v", err)
			}
			text := string(generated)

			if !strings.Contains(text, "session.TimeFuncWithVersion") {
				t.Fatalf("generated source missing session.TimeFuncWithVersion call:\n%s",
					previewLine(text, "session.Time"))
			}

			// Match: session.TimeFuncWithVersion(<...>, <pineVersion>)
			// The version is the last positional argument before the closing paren.
			rx := regexp.MustCompile(`session\.TimeFuncWithVersion\([^)]*,\s*(\d+)\s*\)`)
			matches := rx.FindAllStringSubmatch(text, -1)
			if len(matches) == 0 {
				t.Fatalf("could not extract version arg from emitted session.TimeFuncWithVersion call:\n%s",
					previewLine(text, "TimeFuncWithVersion"))
			}
			for _, m := range matches {
				if m[1] != itoa(tc.expectArg) {
					t.Fatalf("expected pineVersion=%d in emitted call, got %q (full match: %s)",
						tc.expectArg, m[1], m[0])
				}
				if m[1] == itoa(tc.unexpectArg) {
					t.Fatalf("unexpected pineVersion=%d leaked into emitted call: %s", tc.unexpectArg, m[0])
				}
			}
		})
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

func previewLine(text, needle string) string {
	idx := strings.Index(text, needle)
	if idx < 0 {
		return "<needle not found>"
	}
	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + 200
	if end > len(text) {
		end = len(text)
	}
	return text[start:end]
}

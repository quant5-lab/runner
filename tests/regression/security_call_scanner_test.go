package regression

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type securityRef struct {
	RelStrategyPath string
	Symbol          string // raw arg: string literal stripped of quotes, or "syminfo.tickerid"
	Timeframe       string
}

// Group 1: symbol argument (double-quoted, single-quoted, or syminfo.tickerid).
// Group 2: timeframe argument (double-quoted or single-quoted string).
var reSecurityCall = regexp.MustCompile(
	`(?:request\.)?security\(\s*` +
		`((?:"[^"]*")|(?:'[^']*')|syminfo\.tickerid)` +
		`\s*,\s*` +
		`((?:"[^"]*")|(?:'[^']*'))`,
)

func stripQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

func extractSecurityRefsFromContent(content, relPath string) []securityRef {
	seen := map[string]bool{}
	var refs []securityRef

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		for _, match := range reSecurityCall.FindAllSubmatch([]byte(line), -1) {
			symbol := stripQuotes(string(match[1]))
			tf := stripQuotes(string(match[2]))
			key := symbol + "|" + tf
			if seen[key] {
				continue
			}
			seen[key] = true
			refs = append(refs, securityRef{RelStrategyPath: relPath, Symbol: symbol, Timeframe: tf})
		}
	}
	return refs
}

func extractSecurityRefsFromFile(t *testing.T, path, relPath string) []securityRef {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Logf("read %s: %v (skipped)", relPath, err)
		return nil
	}
	return extractSecurityRefsFromContent(string(data), relPath)
}

func scanStrategiesForSecurityCalls(t *testing.T, strategiesDir string) []securityRef {
	t.Helper()
	var refs []securityRef
	err := filepath.WalkDir(strategiesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isPineStrategy(path) {
			return nil
		}
		rel, _ := filepath.Rel(strategiesDir, path)
		refs = append(refs, extractSecurityRefsFromFile(t, path, rel)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk strategies dir: %v", err)
	}
	return refs
}

func isPineStrategy(path string) bool {
	return filepath.Ext(path) == ".pine" && !strings.HasSuffix(path, ".pine.skip")
}

// Runtime-symbol refs use syminfo.tickerid; their required fixture depends on the runtime symbol.
func partitionBySymbolType(refs []securityRef) (literal, runtime []securityRef) {
	for _, r := range refs {
		if r.Symbol == "syminfo.tickerid" {
			runtime = append(runtime, r)
		} else {
			literal = append(literal, r)
		}
	}
	return
}

func TestStripQuotes(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`"BTCUSDT"`, "BTCUSDT"},
		{`'SBERP'`, "SBERP"},
		{`"1D"`, "1D"},
		{`'D'`, "D"},
		{"syminfo.tickerid", "syminfo.tickerid"},
		{"", ""},
		{`"a"`, "a"},
		{`""`, ""},
		{`''`, ""},
		{`"unclosed`, `"unclosed`},
		{`'mixed"`, `'mixed"`},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := stripQuotes(tc.input); got != tc.want {
				t.Errorf("stripQuotes(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestExtractSecurityRefsFromContent(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    []securityRef
	}{
		{
			name:    "double-quoted symbol and timeframe",
			content: `sma_daily = security("BTCUSDT", "1D", ta.sma(close, 20))`,
			want:    []securityRef{{Symbol: "BTCUSDT", Timeframe: "1D"}},
		},
		{
			name:    "single-quoted symbol and timeframe",
			content: `sma_daily = security('SBERP', 'D', ta.sma(close, 50))`,
			want:    []securityRef{{Symbol: "SBERP", Timeframe: "D"}},
		},
		{
			name:    "syminfo.tickerid",
			content: `d = security(syminfo.tickerid, "1D", close)`,
			want:    []securityRef{{Symbol: "syminfo.tickerid", Timeframe: "1D"}},
		},
		{
			name:    "request.security prefix",
			content: `d = request.security(syminfo.tickerid, "D", close)`,
			want:    []securityRef{{Symbol: "syminfo.tickerid", Timeframe: "D"}},
		},
		{
			name:    "full-line comment skipped",
			content: "// sma = security(\"BTC\", \"1D\", close)\nreal = security(\"ETH\", \"1D\", close)",
			want:    []securityRef{{Symbol: "ETH", Timeframe: "1D"}},
		},
		{
			name:    "inline comment not skipped (security before //)",
			content: `x = security("BTC", "1D", close) // daily close`,
			want:    []securityRef{{Symbol: "BTC", Timeframe: "1D"}},
		},
		{
			name:    "duplicate calls deduplicated within file",
			content: "a = security(\"X\", \"1D\", close)\nb = security(\"X\", \"1D\", open)",
			want:    []securityRef{{Symbol: "X", Timeframe: "1D"}},
		},
		{
			name:    "different timeframes not deduplicated",
			content: "a = security(\"X\", \"1D\", close)\nb = security(\"X\", \"1h\", close)",
			want: []securityRef{
				{Symbol: "X", Timeframe: "1D"},
				{Symbol: "X", Timeframe: "1h"},
			},
		},
		{
			name:    "different symbols not deduplicated",
			content: "a = security(\"X\", \"1D\", close)\nb = security(\"Y\", \"1D\", close)",
			want: []securityRef{
				{Symbol: "X", Timeframe: "1D"},
				{Symbol: "Y", Timeframe: "1D"},
			},
		},
		{
			name:    "exchange-prefixed literal symbol preserved",
			content: `d = request.security("BINANCE:BTCUSDT", "D", close)`,
			want:    []securityRef{{Symbol: "BINANCE:BTCUSDT", Timeframe: "D"}},
		},
		{
			name:    "empty content produces no refs",
			content: "",
			want:    nil,
		},
		{
			name:    "variable-symbol call not matched",
			content: `d = security(ticker_var, "1D", close)`,
			want:    nil,
		},
		{
			name:    "whitespace between args",
			content: `d = security( "SYM" , "1D" , close)`,
			want:    []securityRef{{Symbol: "SYM", Timeframe: "1D"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractSecurityRefsFromContent(tc.content, "test.pine")

			if len(got) != len(tc.want) {
				t.Fatalf("got %d refs %v, want %d refs %v", len(got), got, len(tc.want), tc.want)
			}
			for i, w := range tc.want {
				if got[i].Symbol != w.Symbol {
					t.Errorf("[%d] Symbol: got %q, want %q", i, got[i].Symbol, w.Symbol)
				}
				if got[i].Timeframe != w.Timeframe {
					t.Errorf("[%d] Timeframe: got %q, want %q", i, got[i].Timeframe, w.Timeframe)
				}
			}
		})
	}
}

func TestPartitionBySymbolType(t *testing.T) {
	refs := []securityRef{
		{Symbol: "syminfo.tickerid", Timeframe: "1D"},
		{Symbol: "BTCUSDT", Timeframe: "1D"},
		{Symbol: "syminfo.tickerid", Timeframe: "D"},
		{Symbol: "SBERP", Timeframe: "1h"},
	}
	literal, runtime := partitionBySymbolType(refs)

	if len(literal) != 2 {
		t.Errorf("literal count: got %d, want 2", len(literal))
	}
	if len(runtime) != 2 {
		t.Errorf("runtime count: got %d, want 2", len(runtime))
	}
	for _, r := range literal {
		if r.Symbol == "syminfo.tickerid" {
			t.Errorf("literal slice contains syminfo.tickerid: %+v", r)
		}
	}
	for _, r := range runtime {
		if r.Symbol != "syminfo.tickerid" {
			t.Errorf("runtime slice contains literal symbol: %+v", r)
		}
	}
}

func TestPartitionBySymbolType_AllLiteral(t *testing.T) {
	refs := []securityRef{{Symbol: "BTC", Timeframe: "1D"}}
	literal, runtime := partitionBySymbolType(refs)
	if len(literal) != 1 || len(runtime) != 0 {
		t.Errorf("got literal=%d runtime=%d, want 1/0", len(literal), len(runtime))
	}
}

func TestPartitionBySymbolType_AllRuntime(t *testing.T) {
	refs := []securityRef{
		{Symbol: "syminfo.tickerid", Timeframe: "1D"},
		{Symbol: "syminfo.tickerid", Timeframe: "D"},
	}
	literal, runtime := partitionBySymbolType(refs)
	if len(literal) != 0 || len(runtime) != 2 {
		t.Errorf("got literal=%d runtime=%d, want 0/2", len(literal), len(runtime))
	}
}

func TestPartitionBySymbolType_Empty(t *testing.T) {
	literal, runtime := partitionBySymbolType(nil)
	if literal != nil || runtime != nil {
		t.Errorf("expected nil slices for empty input")
	}
}

func TestExtractSecurityRefsFromFile_ReadsAndParsesCorrectly(t *testing.T) {
	dir := t.TempDir()
	content := "x = security(\"ETHUSDT\", \"1D\", close)\n// y = security(\"SKIP\", \"1D\", close)"
	path := filepath.Join(dir, "strat.pine")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got := extractSecurityRefsFromFile(t, path, "strat.pine")
	if len(got) != 1 || got[0].Symbol != "ETHUSDT" {
		t.Errorf("got %v, want single ETHUSDT ref", got)
	}
}

func TestExtractSecurityRefsFromContent_MultipleCallsOnOneLine(t *testing.T) {
	line := `a = security("X", "1D", close) b = security("Y", "4h", open)`
	got := extractSecurityRefsFromContent(line, "test.pine")
	if len(got) != 2 {
		t.Fatalf("got %d refs, want 2: %v", len(got), got)
	}
	if got[0].Symbol != "X" || got[0].Timeframe != "1D" {
		t.Errorf("ref[0]: got (%q, %q), want (X, 1D)", got[0].Symbol, got[0].Timeframe)
	}
	if got[1].Symbol != "Y" || got[1].Timeframe != "4h" {
		t.Errorf("ref[1]: got (%q, %q), want (Y, 4h)", got[1].Symbol, got[1].Timeframe)
	}
}

func TestExtractSecurityRefsFromContent_NoSecurityCalls(t *testing.T) {
	content := "//@version=5\nstrategy(\"no security\", overlay=true)\nplot(close)"
	got := extractSecurityRefsFromContent(content, "plain.pine")
	if len(got) != 0 {
		t.Errorf("got %v refs, want none", got)
	}
}

func TestIsPineStrategy(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"strategies/foo.pine", true},
		{"strategies/foo.PINE", false},
		{"strategies/foo.pine.skip", false},
		{"strategies/foo.skip", false},
		{"strategies/foo.go", false},
		{"strategies/foo", false},
		{"strategies/dir/nested.pine", true},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := isPineStrategy(tc.path); got != tc.want {
				t.Errorf("isPineStrategy(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestScanStrategiesForSecurityCalls_WalksSubdirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	rootPine := `security("ROOT", "1D", close)`
	subPine := `security("SUB", "1h", close)`
	skipPine := `security("SKIP", "1D", close)` // must not appear

	for path, content := range map[string]string{
		filepath.Join(dir, "root.pine"):         rootPine,
		filepath.Join(sub, "nested.pine"):       subPine,
		filepath.Join(dir, "ignored.pine.skip"): skipPine,
	} {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	refs := scanStrategiesForSecurityCalls(t, dir)

	symbols := make(map[string]bool)
	for _, r := range refs {
		symbols[r.Symbol] = true
	}

	if !symbols["ROOT"] {
		t.Error("expected ROOT symbol from root.pine")
	}
	if !symbols["SUB"] {
		t.Error("expected SUB symbol from sub/nested.pine")
	}
	if symbols["SKIP"] {
		t.Error("SKIP symbol from .pine.skip file must not be included")
	}
}

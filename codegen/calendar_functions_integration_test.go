package codegen

import (
	"os"
	"strings"
	"testing"
)

/* TestCalendarFunctions_FixtureCompilation validates .pine fixture files compile through full codegen pipeline */
func TestCalendarFunctions_FixtureCompilation(t *testing.T) {
	fixturesDir := "../e2e/fixtures/strategies"

	tests := []struct {
		name        string
		fixture     string
		mustContain []string
	}{
		{
			name:    "calendar extraction functions",
			fixture: "test-calendar-extraction.pine",
			mustContain: []string{
				"calendar.Year(", "calendar.Month(",
				"calendar.DayOfWeek(", "calendar.DayOfMonth(",
				"calendar.Hour(", "calendar.Minute(",
				"calendar.Second(", "calendar.WeekOfYear(",
			},
		},
		{
			name:    "timestamp overloads",
			fixture: "test-timestamp.pine",
			mustContain: []string{
				"calendar.TimestampFromString(",
				"calendar.Timestamp(",
			},
		},
		{
			name:    "timeframe functions",
			fixture: "test-timeframe-functions.pine",
			mustContain: []string{
				"context.TimeframeToSeconds(",
				"context.TimeframeFromSeconds(",
				"context.AlignTimestampToPeriod(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(fixturesDir + "/" + tt.fixture)
			if err != nil {
				t.Fatalf("Failed to read fixture %s: %v", tt.fixture, err)
			}

			code, err := compilePineScript(string(content))
			if err != nil {
				t.Fatalf("Compilation failed for %s: %v", tt.fixture, err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}

			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO placeholder\nGenerated code:\n%s", code)
			}
		})
	}
}

/* TestCalendarFunctions_ExplicitTimezone validates timezone arg propagates through codegen */
func TestCalendarFunctions_ExplicitTimezone(t *testing.T) {
	content, err := os.ReadFile("../e2e/fixtures/strategies/test-calendar-extraction.pine")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	code, err := compilePineScript(string(content))
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	/* year(time, "UTC") and hour(time, "UTC") must produce explicit "UTC" in codegen */
	utcCount := strings.Count(code, `"UTC"`)
	if utcCount < 2 {
		t.Errorf("Expected at least 2 explicit UTC timezone args, found %d\nGenerated:\n%s", utcCount, code)
	}
}

/* TestTimestampOverloads_AllVariants validates each timestamp arity produces distinct codegen */
func TestTimestampOverloads_AllVariants(t *testing.T) {
	content, err := os.ReadFile("../e2e/fixtures/strategies/test-timestamp.pine")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	code, err := compilePineScript(string(content))
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	/* 1-arg string overload → TimestampFromString */
	if !strings.Contains(code, "calendar.TimestampFromString(") {
		t.Errorf("Missing TimestampFromString for 1-arg string overload\nGenerated:\n%s", code)
	}

	/* 5/6/7-arg component overloads → Timestamp with 7 args */
	timestampCalls := strings.Count(code, "calendar.Timestamp(")
	if timestampCalls < 4 {
		t.Errorf("Expected at least 4 calendar.Timestamp calls (5/6/tz5/7 arg), found %d\nGenerated:\n%s", timestampCalls, code)
	}
}

/* TestTimeframeFunctions_ChangeIIFE validates timeframe.change produces IIFE pattern */
func TestTimeframeFunctions_ChangeIIFE(t *testing.T) {
	content, err := os.ReadFile("../e2e/fixtures/strategies/test-timeframe-functions.pine")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	code, err := compilePineScript(string(content))
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	/* timeframe.change produces an IIFE with AlignTimestampToPeriod */
	if !strings.Contains(code, "(func()") {
		t.Errorf("Missing IIFE pattern for timeframe.change\nGenerated:\n%s", code)
	}

	if !strings.Contains(code, "currAligned") {
		t.Errorf("Missing currAligned variable in timeframe.change IIFE\nGenerated:\n%s", code)
	}

	if !strings.Contains(code, "prevAligned") {
		t.Errorf("Missing prevAligned variable in timeframe.change IIFE\nGenerated:\n%s", code)
	}
}

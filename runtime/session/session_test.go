package session

import (
	"math"
	"testing"
	"time"
)

func TestParse_ValidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkFields func(*testing.T, *Session)
	}{
		{
			name:    "regular_trading_hours",
			input:   "0950-1645",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.tr.startMinutes != 9*60+50 {
					t.Errorf("startMinutes = %d, want %d", s.tr.startMinutes, 9*60+50)
				}
				if s.tr.endMinutes != 16*60+45 {
					t.Errorf("endMinutes = %d, want %d", s.tr.endMinutes, 16*60+45)
				}
				if s.tr.is24Hour {
					t.Error("is24Hour = true, want false")
				}
			},
		},
		{
			name:    "24_hour_session",
			input:   "0000-2359",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if !s.tr.is24Hour {
					t.Error("is24Hour = false, want true")
				}
			},
		},
		{
			name:    "overnight_session",
			input:   "1800-0600",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.tr.startMinutes != 18*60 {
					t.Errorf("startMinutes = %d, want %d", s.tr.startMinutes, 18*60)
				}
				if s.tr.endMinutes != 6*60 {
					t.Errorf("endMinutes = %d, want %d", s.tr.endMinutes, 6*60)
				}
			},
		},
		{
			name:    "full_cycle_overnight",
			input:   "1700-1700",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.tr.startMinutes != 17*60 || s.tr.endMinutes != 17*60 {
					t.Errorf("start=%d end=%d, want both 1020", s.tr.startMinutes, s.tr.endMinutes)
				}
				if s.tr.is24Hour {
					t.Error("is24Hour = true, want false for full-cycle overnight")
				}
			},
		},
		{
			name:    "midnight_start",
			input:   "0000-1200",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.tr.startMinutes != 0 {
					t.Errorf("startMinutes = %d, want 0", s.tr.startMinutes)
				}
			},
		},
		{
			name:    "late_night_end",
			input:   "1200-2359",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.tr.endMinutes != 23*60+59 {
					t.Errorf("endMinutes = %d, want %d", s.tr.endMinutes, 23*60+59)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFields != nil {
				tt.checkFields(t, s)
			}
		})
	}
}

func TestParse_InvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"missing_hyphen", "09501645"},
		{"wrong_separator", "0950/1645"},
		{"too_short_start", "950-1645"},
		{"too_short_end", "0950-645"},
		{"too_long_start", "00950-1645"},
		{"too_long_end", "0950-16450"},
		{"invalid_hour_25", "2500-1645"},
		{"invalid_minute_60", "0960-1645"},
		{"negative_hour", "-100-1645"},
		{"non_numeric", "abcd-1645"},
		{"single_number", "0950"},
		{"too_many_parts", "0950-1645-1800"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.input); err == nil {
				t.Errorf("Parse(%q) = nil error, want error", tt.input)
			}
		})
	}
}

// TestIsInSession_RegularHours covers intraday time-range membership including exact
// second boundaries. Parse() defaults to v5 (all 7 days) so the weekday mask is
// irrelevant; all assertions are about HHMM filtering only.
func TestIsInSession_RegularHours(t *testing.T) {
	s, err := Parse("0950-1645")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tests := []struct {
		name      string
		timestamp string
		wantIn    bool
	}{
		{"before_start", "2025-11-15T09:49:59Z", false},
		{"exact_start", "2025-11-15T09:50:00Z", true},
		{"mid_session", "2025-11-15T12:00:00Z", true},
		{"exact_end_minute", "2025-11-15T16:45:00Z", true},
		{"one_second_past_end", "2025-11-15T16:45:01Z", false},
		{"after_session", "2025-11-15T18:00:00Z", false},
		{"midnight", "2025-11-15T00:00:00Z", false},
		{"early_morning", "2025-11-15T06:00:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			got := s.IsInSession(tm.UnixMilli(), "UTC")
			if got != tt.wantIn {
				t.Errorf("IsInSession(%s) = %v, want %v", tt.timestamp, got, tt.wantIn)
			}
		})
	}
}

func TestIsInSession_24HourSession(t *testing.T) {
	s, err := Parse("0000-2359")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	timestamps := []string{
		"2025-11-15T00:00:00Z",
		"2025-11-15T06:30:00Z",
		"2025-11-15T12:00:00Z",
		"2025-11-15T18:45:00Z",
		"2025-11-15T23:59:00Z",
	}
	for _, ts := range timestamps {
		t.Run(ts, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, ts)
			if !s.IsInSession(tm.UnixMilli(), "UTC") {
				t.Errorf("24-hour session must be IN for every timestamp; got OUT at %s", ts)
			}
		})
	}
}

// TestIsInSession_OvernightSession covers time-range membership for sessions that cross
// midnight, including exact second boundaries at start and end. Parse() defaults to v5
// (all 7 days) so weekday mask is irrelevant here; day-mask behavior for overnight
// sessions is covered in TestSession_DayFilter.
func TestIsInSession_OvernightSession(t *testing.T) {
	s, err := Parse("1800-0600")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tests := []struct {
		name      string
		timestamp string
		wantIn    bool
	}{
		// Gap between end and start — not in session.
		{"gap_afternoon", "2025-11-15T15:00:00Z", false},
		{"one_second_before_start", "2025-11-15T17:59:59Z", false},
		// Session start boundary.
		{"overnight_exact_start", "2025-11-15T18:00:00Z", true},
		// Pre-midnight half.
		{"overnight_pre_midnight_evening", "2025-11-15T22:00:00Z", true},
		// Post-midnight half.
		{"overnight_post_midnight", "2025-11-16T00:00:00Z", true},
		{"overnight_early_morning", "2025-11-16T03:00:00Z", true},
		// Session end boundary.
		{"overnight_exact_end_minute", "2025-11-16T06:00:00Z", true},
		{"overnight_one_second_past_end", "2025-11-16T06:00:01Z", false},
		// Back in gap after end.
		{"gap_morning", "2025-11-16T09:00:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			got := s.IsInSession(tm.UnixMilli(), "UTC")
			if got != tt.wantIn {
				t.Errorf("IsInSession(%s) = %v, want %v", tt.timestamp, got, tt.wantIn)
			}
		})
	}
}

// TestIsInSession_FullCycleOvernightSession covers the "HHMM-HHMM" form where start
// equals end (e.g. "1700-1700"). Pine interprets this as a session covering all 24
// hours on the specified days: every time of day falls within the time range. Parse()
// defaults to v5 (all 7 days), so every bar is in session regardless of time.
func TestIsInSession_FullCycleOvernightSession(t *testing.T) {
	s, err := Parse("1700-1700")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	timestamps := []string{
		"2025-09-01T17:00:00Z", // exact start/end minute
		"2025-09-01T16:59:59Z", // one second before the "start" minute (post-midnight half)
		"2025-09-01T17:00:01Z", // one second after the "start" minute (pre-midnight half)
		"2025-09-01T00:00:00Z", // midnight
		"2025-09-01T12:00:00Z", // midday
		"2025-09-01T23:59:59Z", // end of calendar day
	}
	for _, ts := range timestamps {
		t.Run(ts, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, ts)
			if !s.IsInSession(tm.UnixMilli(), "UTC") {
				t.Errorf("full-cycle overnight must be IN for every time; got OUT at %s", ts)
			}
		})
	}
}

func TestTimeFunc_WithinSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()
	result := TimeFunc(timestamp, "1h", "0950-1645", "UTC")
	if math.IsNaN(result) {
		t.Error("TimeFunc() returned NaN for in-session bar")
	}
	if result != float64(timestamp) {
		t.Errorf("TimeFunc() = %v, want %v", result, float64(timestamp))
	}
}

func TestTimeFunc_OutsideSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T18:00:00Z")
	result := TimeFunc(tm.UnixMilli(), "1h", "0950-1645", "UTC")
	if !math.IsNaN(result) {
		t.Errorf("TimeFunc() = %v for out-of-session bar, want NaN", result)
	}
}

func TestTimeFunc_EmptySession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()
	result := TimeFunc(timestamp, "1h", "", "UTC")
	if result != float64(timestamp) {
		t.Errorf("TimeFunc() with empty session = %v, want %v (pass-through)", result, float64(timestamp))
	}
}

func TestTimeFunc_InvalidSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	result := TimeFunc(tm.UnixMilli(), "1h", "invalid-format", "UTC")
	if !math.IsNaN(result) {
		t.Errorf("TimeFunc() with invalid session = %v, want NaN", result)
	}
}

func TestTimeFunc_24HourSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T03:30:00Z")
	result := TimeFunc(tm.UnixMilli(), "1h", "0000-2359", "UTC")
	if math.IsNaN(result) {
		t.Error("TimeFunc() with 24-hour session returned NaN, want timestamp")
	}
}

// TestTimeFunc_PineScriptPattern_NA verifies the na(time(period, session)) Pine idiom:
// time() returns NaN outside session, non-NaN inside. session_open := not na(result).
func TestTimeFunc_PineScriptPattern_NA(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		session   string
		wantNA    bool
	}{
		{"in_session", "2025-11-15T12:00:00Z", "0950-1645", false},
		{"out_of_session", "2025-11-15T18:00:00Z", "0950-1645", true},
		{"24h_never_na", "2025-11-15T03:00:00Z", "0000-2359", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			result := TimeFunc(tm.UnixMilli(), "1h", tt.session, "UTC")
			if math.IsNaN(result) != tt.wantNA {
				t.Errorf("na(TimeFunc(%s)) = %v, want %v", tt.timestamp, math.IsNaN(result), tt.wantNA)
			}
		})
	}
}

func BenchmarkIsInSession_RegularHours(b *testing.B) {
	s, _ := Parse("0950-1645")
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.IsInSession(timestamp, "UTC")
	}
}

func BenchmarkIsInSession_24Hour(b *testing.B) {
	s, _ := Parse("0000-2359")
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.IsInSession(timestamp, "UTC")
	}
}

func BenchmarkTimeFunc(b *testing.B) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeFunc(timestamp, "1h", "0950-1645", "UTC")
	}
}

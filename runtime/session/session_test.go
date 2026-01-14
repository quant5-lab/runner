package session

import (
	"math"
	"testing"
	"time"
)

/* Test Suite: Session Parsing (Format Validation) */

func TestParse_ValidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkFields func(*testing.T, *Session)
	}{
		{
			name:    "Regular trading hours",
			input:   "0950-1645",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.startHour != 9 || s.startMinute != 50 {
					t.Errorf("start time = %02d:%02d, want 09:50", s.startHour, s.startMinute)
				}
				if s.endHour != 16 || s.endMinute != 45 {
					t.Errorf("end time = %02d:%02d, want 16:45", s.endHour, s.endMinute)
				}
				if s.is24Hour {
					t.Error("is24Hour = true, want false")
				}
			},
		},
		{
			name:    "24-hour session",
			input:   "0000-2359",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if !s.is24Hour {
					t.Error("is24Hour = false, want true")
				}
			},
		},
		{
			name:    "Overnight session",
			input:   "1800-0600",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.startHour != 18 || s.startMinute != 0 {
					t.Errorf("start time = %02d:%02d, want 18:00", s.startHour, s.startMinute)
				}
				if s.endHour != 6 || s.endMinute != 0 {
					t.Errorf("end time = %02d:%02d, want 06:00", s.endHour, s.endMinute)
				}
			},
		},
		{
			name:    "Midnight start",
			input:   "0000-1200",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.startHour != 0 {
					t.Errorf("startHour = %d, want 0", s.startHour)
				}
			},
		},
		{
			name:    "Late night end",
			input:   "1200-2359",
			wantErr: false,
			checkFields: func(t *testing.T, s *Session) {
				if s.endHour != 23 || s.endMinute != 59 {
					t.Errorf("end time = %02d:%02d, want 23:59", s.endHour, s.endMinute)
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
		name    string
		input   string
		wantErr bool
	}{
		{"Empty string", "", true},
		{"Missing hyphen", "09501645", true},
		{"Wrong separator", "0950/1645", true},
		{"Too short start", "950-1645", true},
		{"Too short end", "0950-645", true},
		{"Too long start", "00950-1645", true},
		{"Too long end", "0950-16450", true},
		{"Invalid hour (25)", "2500-1645", true},
		{"Invalid minute (60)", "0960-1645", true},
		{"Negative hour", "-100-1645", true},
		{"Non-numeric", "abcd-1645", true},
		{"Single number", "0950", true},
		{"Too many parts", "0950-1645-1800", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

/* Test Suite: Session Filtering (IsInSession) */

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
		// Before session
		{"Before session start", "2025-11-15T09:49:59Z", false},
		// Session boundaries
		{"Exact session start", "2025-11-15T09:50:00Z", true},
		{"During session", "2025-11-15T12:00:00Z", true},
		{"Exact session end", "2025-11-15T16:45:00Z", true},
		// After session
		{"One second after end", "2025-11-15T16:45:01Z", false},
		{"After session", "2025-11-15T18:00:00Z", false},
		// Edge cases
		{"Midnight (out)", "2025-11-15T00:00:00Z", false},
		{"Early morning (out)", "2025-11-15T06:00:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			timestamp := tm.UnixMilli()
			got := s.IsInSession(timestamp, "UTC")
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

	// 24-hour session should ALWAYS return true (fast path optimization)
	tests := []string{
		"2025-11-15T00:00:00Z",
		"2025-11-15T06:30:00Z",
		"2025-11-15T12:00:00Z",
		"2025-11-15T18:45:00Z",
		"2025-11-15T23:59:00Z",
	}

	for _, timestamp := range tests {
		t.Run(timestamp, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, timestamp)
			if !s.IsInSession(tm.UnixMilli(), "UTC") {
				t.Errorf("24-hour session should always be IN, got OUT for %s", timestamp)
			}
		})
	}
}

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
		// Before overnight period
		{"Afternoon (out)", "2025-11-15T15:00:00Z", false},
		// Evening (start of overnight)
		{"Session start", "2025-11-15T18:00:00Z", true},
		{"Late evening", "2025-11-15T22:00:00Z", true},
		{"Midnight", "2025-11-16T00:00:00Z", true},
		// Early morning (end of overnight)
		{"Early morning", "2025-11-16T03:00:00Z", true},
		{"Session end", "2025-11-16T06:00:00Z", true},
		// After overnight period
		{"One minute after", "2025-11-16T06:01:00Z", false},
		{"Morning (out)", "2025-11-16T09:00:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			timestamp := tm.UnixMilli()
			got := s.IsInSession(timestamp, "UTC")
			if got != tt.wantIn {
				t.Errorf("IsInSession(%s) = %v, want %v", tt.timestamp, got, tt.wantIn)
			}
		})
	}
}

/* Test Suite: TimeFunc (Pine Script time() function) */

func TestTimeFunc_WithinSession(t *testing.T) {
	// Regular trading hours: 09:50-16:45
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli() // Seconds, not milliseconds

	result := TimeFunc(timestamp, "1h", "0950-1645", "UTC")

	if math.IsNaN(result) {
		t.Error("TimeFunc() returned NaN, want valid timestamp")
	}
	if result != float64(timestamp) {
		t.Errorf("TimeFunc() = %v, want %v", result, float64(timestamp))
	}
}

func TestTimeFunc_OutsideSession(t *testing.T) {
	// Outside trading hours: 09:50-16:45
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T18:00:00Z")
	timestamp := tm.UnixMilli()

	result := TimeFunc(timestamp, "1h", "0950-1645", "UTC")

	if !math.IsNaN(result) {
		t.Errorf("TimeFunc() = %v, want NaN", result)
	}
}

func TestTimeFunc_EmptySession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()

	result := TimeFunc(timestamp, "1h", "", "UTC")

	// Empty session string = no filtering, return timestamp
	if result != float64(timestamp) {
		t.Errorf("TimeFunc() with empty session = %v, want %v", result, float64(timestamp))
	}
}

func TestTimeFunc_InvalidSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T12:00:00Z")
	timestamp := tm.UnixMilli()

	result := TimeFunc(timestamp, "1h", "invalid-format", "UTC")

	// Invalid session = always NaN
	if !math.IsNaN(result) {
		t.Errorf("TimeFunc() with invalid session = %v, want NaN", result)
	}
}

func TestTimeFunc_24HourSession(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2025-11-15T03:30:00Z")
	timestamp := tm.UnixMilli()

	result := TimeFunc(timestamp, "1h", "0000-2359", "UTC")

	if math.IsNaN(result) {
		t.Error("TimeFunc() with 24-hour session returned NaN, want timestamp")
	}
}

/* Test Suite: Pine Script Usage Patterns */

func TestTimeFunc_PineScriptPattern_NA(t *testing.T) {
	// Pine pattern: session_open = na(time(timeframe.period, "0950-1645")) ? false : true

	tests := []struct {
		name            string
		timestamp       string
		session         string
		expectNA        bool
		expectInSession bool
	}{
		{
			name:            "During session - not NA",
			timestamp:       "2025-11-15T12:00:00Z",
			session:         "0950-1645",
			expectNA:        false,
			expectInSession: true,
		},
		{
			name:            "Outside session - is NA",
			timestamp:       "2025-11-15T18:00:00Z",
			session:         "0950-1645",
			expectNA:        true,
			expectInSession: false,
		},
		{
			name:            "24-hour session - never NA",
			timestamp:       "2025-11-15T03:00:00Z",
			session:         "0000-2359",
			expectNA:        false,
			expectInSession: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tt.timestamp)
			result := TimeFunc(tm.UnixMilli(), "1h", tt.session, "UTC")

			isNA := math.IsNaN(result)
			if isNA != tt.expectNA {
				t.Errorf("na(result) = %v, want %v", isNA, tt.expectNA)
			}

			// Pine script pattern: session_open = not na(result)
			sessionOpen := !isNA
			if sessionOpen != tt.expectInSession {
				t.Errorf("session_open = %v, want %v", sessionOpen, tt.expectInSession)
			}
		})
	}
}

/* Benchmark: Performance Validation */

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

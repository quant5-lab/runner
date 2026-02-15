package context

import (
	"testing"
	"time"
)

func TestParseTimeframeComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMult int
		wantUnit byte
	}{
		{"uppercase daily", "D", 1, 'D'},
		{"lowercase daily", "d", 1, 'd'},
		{"numeric daily", "1D", 1, 'D'},
		{"multi-day", "5D", 5, 'D'},
		{"uppercase weekly", "W", 1, 'W'},
		{"lowercase weekly", "w", 1, 'w'},
		{"numeric weekly", "1W", 1, 'W'},
		{"multi-week 2W", "2W", 2, 'W'},
		{"multi-week 4W", "4W", 4, 'W'},
		{"uppercase monthly", "M", 1, 'M'},
		{"numeric monthly", "1M", 1, 'M'},
		{"quarterly", "3M", 3, 'M'},
		{"semi-annual", "6M", 6, 'M'},
		{"annual", "12M", 12, 'M'},
		{"hourly", "1h", 1, 'h'},
		{"4-hour", "4h", 4, 'h'},
		{"minute", "1m", 1, 'm'},
		{"5-minute", "5m", 5, 'm'},
		{"15-minute", "15m", 15, 'm'},
		{"pure numeric 1", "1", 1, 'm'},
		{"pure numeric 5", "5", 5, 'm'},
		{"pure numeric 15", "15", 15, 'm'},
		{"pure numeric 60", "60", 60, 'm'},
		{"pure numeric 240", "240", 240, 'm'},
		{"second", "1s", 1, 's'},
		{"30-second", "30s", 30, 's'},
		{"empty string", "", 0, byte(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mult, unit := parseTimeframeComponents(tt.input)
			if mult != tt.wantMult || unit != tt.wantUnit {
				t.Errorf("parseTimeframeComponents(%q) = (%d, %c), want (%d, %c)",
					tt.input, mult, unit, tt.wantMult, tt.wantUnit)
			}
		})
	}
}

func TestTimeframeBoundaryAligner_FixedDuration(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	tests := []struct {
		name      string
		timestamp int64
		timeframe string
		expected  int64
	}{
		{"1-second at HH:MM:SS.500", mustUnix(t, "2024-06-15 14:33:27"), "1s", mustUnix(t, "2024-06-15 14:33:27")},
		{"30-second at HH:MM:17", mustUnix(t, "2024-06-15 14:33:17"), "30s", mustUnix(t, "2024-06-15 14:33:00")},
		{"1-minute at HH:MM:30", mustUnix(t, "2024-06-15 14:33:30"), "1m", mustUnix(t, "2024-06-15 14:33:00")},
		{"5-minute at HH:33:00", mustUnix(t, "2024-06-15 14:33:00"), "5m", mustUnix(t, "2024-06-15 14:30:00")},
		{"15-minute at HH:33:00", mustUnix(t, "2024-06-15 14:33:00"), "15m", mustUnix(t, "2024-06-15 14:30:00")},
		{"60-minute (numeric) at HH:33", mustUnix(t, "2024-06-15 14:33:00"), "60", mustUnix(t, "2024-06-15 14:00:00")},
		{"1-hour at HH:33:00", mustUnix(t, "2024-06-15 14:33:00"), "1h", mustUnix(t, "2024-06-15 14:00:00")},
		{"4-hour at 14:33", mustUnix(t, "2024-06-15 14:33:00"), "4h", mustUnix(t, "2024-06-15 12:00:00")},
		{"1-day uppercase at 14:33", mustUnix(t, "2024-06-15 14:33:00"), "1D", mustUnix(t, "2024-06-15 00:00:00")},
		{"1-day lowercase at 14:33", mustUnix(t, "2024-06-15 14:33:00"), "1d", mustUnix(t, "2024-06-15 00:00:00")},
		{"5-day at mid-period", mustUnix(t, "2024-06-18 12:00:00"), "5D", mustUnix(t, "2024-06-16 00:00:00")},
		{"already aligned to hour", mustUnix(t, "2024-06-15 14:00:00"), "1h", mustUnix(t, "2024-06-15 14:00:00")},
		{"already aligned to day", mustUnix(t, "2024-06-15 00:00:00"), "1D", mustUnix(t, "2024-06-15 00:00:00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.AlignToPeriod(tt.timestamp, tt.timeframe)
			if result != tt.expected {
				t.Errorf("AlignToPeriod(%s, %q) = %s, want %s",
					fmtUnix(tt.timestamp), tt.timeframe, fmtUnix(result), fmtUnix(tt.expected))
			}
			if result > tt.timestamp {
				t.Errorf("aligned timestamp %s is after input %s (should align backwards)",
					fmtUnix(result), fmtUnix(tt.timestamp))
			}
		})
	}
}

func TestTimeframeBoundaryAligner_WeeklyAlignment(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	tests := []struct {
		name      string
		timestamp int64
		timeframe string
		expected  int64
	}{
		{"Monday 10:00 → Monday 00:00", mustUnix(t, "2024-06-17 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Tuesday → Monday", mustUnix(t, "2024-06-18 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Wednesday → Monday", mustUnix(t, "2024-06-19 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Thursday → Monday", mustUnix(t, "2024-06-20 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Friday → Monday", mustUnix(t, "2024-06-21 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Saturday → Monday", mustUnix(t, "2024-06-22 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"Sunday → Monday", mustUnix(t, "2024-06-23 10:00:00"), "W", mustUnix(t, "2024-06-17 00:00:00")},
		{"next Monday is new week", mustUnix(t, "2024-06-24 10:00:00"), "W", mustUnix(t, "2024-06-24 00:00:00")},
		{"lowercase w", mustUnix(t, "2024-06-19 14:30:00"), "w", mustUnix(t, "2024-06-17 00:00:00")},
		{"2W first week", mustUnix(t, "2024-06-18 12:00:00"), "2W", mustUnix(t, "2024-06-10 00:00:00")},
		{"2W second week", mustUnix(t, "2024-06-25 12:00:00"), "2W", mustUnix(t, "2024-06-24 00:00:00")},
		{"4W alignment", mustUnix(t, "2024-07-08 12:00:00"), "4W", mustUnix(t, "2024-07-08 00:00:00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.AlignToPeriod(tt.timestamp, tt.timeframe)
			if result != tt.expected {
				t.Errorf("AlignToPeriod(%s, %q) = %s, want %s",
					fmtUnix(tt.timestamp), tt.timeframe, fmtUnix(result), fmtUnix(tt.expected))
			}
			weekday := time.Unix(result, 0).UTC().Weekday()
			if weekday != time.Monday {
				t.Errorf("aligned timestamp %s is %s, expected Monday", fmtUnix(result), weekday)
			}
		})
	}
}

func TestTimeframeBoundaryAligner_MonthlyAlignment(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	tests := []struct {
		name      string
		timestamp int64
		timeframe string
		expected  int64
	}{
		{"mid-January → Jan 1", mustUnix(t, "2024-01-15 14:30:00"), "M", mustUnix(t, "2024-01-01 00:00:00")},
		{"Jan 31 23:59 → Jan 1", mustUnix(t, "2024-01-31 23:59:59"), "M", mustUnix(t, "2024-01-01 00:00:00")},
		{"Feb 1 00:01 → Feb 1", mustUnix(t, "2024-02-01 00:00:01"), "M", mustUnix(t, "2024-02-01 00:00:00")},
		{"Feb 29 leap year → Feb 1", mustUnix(t, "2024-02-29 12:00:00"), "M", mustUnix(t, "2024-02-01 00:00:00")},
		{"Feb 28 non-leap → Feb 1", mustUnix(t, "2023-02-28 12:00:00"), "M", mustUnix(t, "2023-02-01 00:00:00")},
		{"Mar 1 non-leap → Mar 1", mustUnix(t, "2023-03-01 00:00:00"), "M", mustUnix(t, "2023-03-01 00:00:00")},
		{"Dec 31 → Dec 1", mustUnix(t, "2024-12-31 23:59:59"), "M", mustUnix(t, "2024-12-01 00:00:00")},
		{"year boundary Dec → Dec 1", mustUnix(t, "2023-12-15 12:00:00"), "M", mustUnix(t, "2023-12-01 00:00:00")},
		{"year boundary Jan → Jan 1", mustUnix(t, "2024-01-05 12:00:00"), "M", mustUnix(t, "2024-01-01 00:00:00")},
		{"30-day month (June)", mustUnix(t, "2024-06-30 12:00:00"), "M", mustUnix(t, "2024-06-01 00:00:00")},
		{"31-day month (July)", mustUnix(t, "2024-07-31 12:00:00"), "M", mustUnix(t, "2024-07-01 00:00:00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.AlignToPeriod(tt.timestamp, tt.timeframe)
			if result != tt.expected {
				t.Errorf("AlignToPeriod(%s, %q) = %s, want %s",
					fmtUnix(tt.timestamp), tt.timeframe, fmtUnix(result), fmtUnix(tt.expected))
			}
			day := time.Unix(result, 0).UTC().Day()
			if day != 1 {
				t.Errorf("aligned timestamp %s is day %d, expected day 1", fmtUnix(result), day)
			}
		})
	}
}

func TestTimeframeBoundaryAligner_MultiMonthAlignment(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	tests := []struct {
		name      string
		timestamp int64
		timeframe string
		expected  int64
	}{
		{"3M: Jan → Q1 (Jan 1)", mustUnix(t, "2024-01-15 00:00:00"), "3M", mustUnix(t, "2024-01-01 00:00:00")},
		{"3M: Feb → Q1 (Jan 1)", mustUnix(t, "2024-02-15 00:00:00"), "3M", mustUnix(t, "2024-01-01 00:00:00")},
		{"3M: Mar → Q1 (Jan 1)", mustUnix(t, "2024-03-15 00:00:00"), "3M", mustUnix(t, "2024-01-01 00:00:00")},
		{"3M: Apr → Q2 (Apr 1)", mustUnix(t, "2024-04-15 00:00:00"), "3M", mustUnix(t, "2024-04-01 00:00:00")},
		{"3M: May → Q2 (Apr 1)", mustUnix(t, "2024-05-15 00:00:00"), "3M", mustUnix(t, "2024-04-01 00:00:00")},
		{"3M: Jun → Q2 (Apr 1)", mustUnix(t, "2024-06-15 00:00:00"), "3M", mustUnix(t, "2024-04-01 00:00:00")},
		{"3M: Jul → Q3 (Jul 1)", mustUnix(t, "2024-07-15 00:00:00"), "3M", mustUnix(t, "2024-07-01 00:00:00")},
		{"3M: Oct → Q4 (Oct 1)", mustUnix(t, "2024-10-15 00:00:00"), "3M", mustUnix(t, "2024-10-01 00:00:00")},
		{"3M: Dec → Q4 (Oct 1)", mustUnix(t, "2024-12-15 00:00:00"), "3M", mustUnix(t, "2024-10-01 00:00:00")},
		{"6M: Jan → H1 (Jan 1)", mustUnix(t, "2024-01-15 00:00:00"), "6M", mustUnix(t, "2024-01-01 00:00:00")},
		{"6M: Jun → H1 (Jan 1)", mustUnix(t, "2024-06-30 00:00:00"), "6M", mustUnix(t, "2024-01-01 00:00:00")},
		{"6M: Jul → H2 (Jul 1)", mustUnix(t, "2024-07-01 00:00:00"), "6M", mustUnix(t, "2024-07-01 00:00:00")},
		{"6M: Dec → H2 (Jul 1)", mustUnix(t, "2024-12-31 00:00:00"), "6M", mustUnix(t, "2024-07-01 00:00:00")},
		{"12M: mid-year → Jan 1", mustUnix(t, "2024-06-15 00:00:00"), "12M", mustUnix(t, "2024-01-01 00:00:00")},
		{"12M: Dec 31 → Jan 1", mustUnix(t, "2024-12-31 23:59:59"), "12M", mustUnix(t, "2024-01-01 00:00:00")},
		{"12M: Jan 1 → Jan 1", mustUnix(t, "2024-01-01 00:00:00"), "12M", mustUnix(t, "2024-01-01 00:00:00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.AlignToPeriod(tt.timestamp, tt.timeframe)
			if result != tt.expected {
				t.Errorf("AlignToPeriod(%s, %q) = %s, want %s",
					fmtUnix(tt.timestamp), tt.timeframe, fmtUnix(result), fmtUnix(tt.expected))
			}
		})
	}
}

func TestTimeframeBoundaryAligner_EdgeCases(t *testing.T) {
	a := NewTimeframeBoundaryAligner()

	t.Run("empty timeframe returns input unchanged", func(t *testing.T) {
		ts := mustUnix(t, "2024-06-15 14:30:00")
		result := a.AlignToPeriod(ts, "")
		if result != ts {
			t.Errorf("empty timeframe should return input %s, got %s", fmtUnix(ts), fmtUnix(result))
		}
	})

	t.Run("unknown unit returns input unchanged", func(t *testing.T) {
		ts := mustUnix(t, "2024-06-15 14:30:00")
		result := a.AlignToPeriod(ts, "5x")
		if result != ts {
			t.Errorf("unknown unit should return input %s, got %s", fmtUnix(ts), fmtUnix(result))
		}
	})

	t.Run("epoch timestamp (zero)", func(t *testing.T) {
		result := a.AlignToPeriod(0, "1D")
		expected := int64(0)
		if result != expected {
			t.Errorf("AlignToPeriod(epoch, 1D) = %d, want %d", result, expected)
		}
	})

	t.Run("negative timestamp", func(t *testing.T) {
		ts := int64(-86400)
		result := a.AlignToPeriod(ts, "1D")
		if result > ts {
			t.Errorf("aligned negative timestamp %d should be <= input %d", result, ts)
		}
	})

	t.Run("far future timestamp 2050", func(t *testing.T) {
		ts := mustUnix(t, "2050-12-15 14:30:00")
		expected := mustUnix(t, "2050-12-01 00:00:00")
		result := a.AlignToPeriod(ts, "M")
		if result != expected {
			t.Errorf("AlignToPeriod(2050-12-15, M) = %s, want %s", fmtUnix(result), fmtUnix(expected))
		}
	})

	t.Run("leap year Feb 29 2024", func(t *testing.T) {
		ts := mustUnix(t, "2024-02-29 12:00:00")
		expected := mustUnix(t, "2024-02-01 00:00:00")
		result := a.AlignToPeriod(ts, "M")
		if result != expected {
			t.Errorf("leap year Feb 29 alignment failed: got %s, want %s", fmtUnix(result), fmtUnix(expected))
		}
	})

	t.Run("non-leap year Feb 28 2023", func(t *testing.T) {
		ts := mustUnix(t, "2023-02-28 12:00:00")
		expected := mustUnix(t, "2023-02-01 00:00:00")
		result := a.AlignToPeriod(ts, "M")
		if result != expected {
			t.Errorf("non-leap year Feb 28 alignment failed: got %s, want %s", fmtUnix(result), fmtUnix(expected))
		}
	})
}

func TestTimeframeBoundaryAligner_AlignmentProperties(t *testing.T) {
	a := NewTimeframeBoundaryAligner()
	timeframes := []string{"1s", "30s", "1m", "5m", "15m", "1h", "4h", "1D", "W", "M", "3M"}

	t.Run("aligned timestamp always <= input", func(t *testing.T) {
		timestamps := []int64{
			mustUnix(t, "2024-06-15 14:33:27"),
			mustUnix(t, "2024-02-29 12:00:00"),
			mustUnix(t, "2023-12-31 23:59:59"),
			0,
		}
		for _, ts := range timestamps {
			for _, tf := range timeframes {
				result := a.AlignToPeriod(ts, tf)
				if result > ts {
					t.Errorf("AlignToPeriod(%s, %q) = %s is after input (violates alignment property)",
						fmtUnix(ts), tf, fmtUnix(result))
				}
			}
		}
	})

	t.Run("idempotent alignment", func(t *testing.T) {
		ts := mustUnix(t, "2024-06-15 14:33:27")
		for _, tf := range timeframes {
			first := a.AlignToPeriod(ts, tf)
			second := a.AlignToPeriod(first, tf)
			if first != second {
				t.Errorf("AlignToPeriod is not idempotent for %q: first=%s, second=%s",
					tf, fmtUnix(first), fmtUnix(second))
			}
		}
	})
}

func TestAlignTimestampToPeriod_PackageFunction(t *testing.T) {
	ts := mustUnix(t, "2024-06-15 14:33:00")

	t.Run("delegates to boundary aligner", func(t *testing.T) {
		expected := mustUnix(t, "2024-06-15 14:30:00")
		result := AlignTimestampToPeriod(ts, "5m")
		if result != expected {
			t.Errorf("AlignTimestampToPeriod(%s, 5m) = %s, want %s",
				fmtUnix(ts), fmtUnix(result), fmtUnix(expected))
		}
	})

	t.Run("weekly alignment via package function", func(t *testing.T) {
		monday := mustUnix(t, "2024-06-17 00:00:00")
		wednesday := mustUnix(t, "2024-06-19 14:30:00")
		result := AlignTimestampToPeriod(wednesday, "W")
		if result != monday {
			t.Errorf("AlignTimestampToPeriod(Wed, W) = %s, want %s", fmtUnix(result), fmtUnix(monday))
		}
	})

	t.Run("monthly alignment via package function", func(t *testing.T) {
		expected := mustUnix(t, "2024-06-01 00:00:00")
		result := AlignTimestampToPeriod(ts, "M")
		if result != expected {
			t.Errorf("AlignTimestampToPeriod(%s, M) = %s, want %s",
				fmtUnix(ts), fmtUnix(result), fmtUnix(expected))
		}
	})
}

func mustUnix(t *testing.T, dateStr string) int64 {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		t.Fatalf("failed to parse %q: %v", dateStr, err)
	}
	return parsed.Unix()
}

func fmtUnix(sec int64) string {
	return time.Unix(sec, 0).UTC().Format("2006-01-02 15:04:05")
}

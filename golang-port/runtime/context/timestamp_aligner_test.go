package context

import "testing"

func TestTimestampAligner_AlignToTimeframe(t *testing.T) {
	aligner := NewTimestampAligner()

	tests := []struct {
		name             string
		timestamp        int64
		timeframeSeconds int64
		expected         int64
	}{
		{
			name:             "Align to daily boundary",
			timestamp:        1704117000, // 2024-01-01 14:30:00
			timeframeSeconds: 86400,      // 1 day
			expected:         1704067200, // 2024-01-01 00:00:00
		},
		{
			name:             "Align to hourly boundary",
			timestamp:        1704117000, // 2024-01-01 14:30:00
			timeframeSeconds: 3600,       // 1 hour
			expected:         1704114000, // 2024-01-01 14:00:00
		},
		{
			name:             "Already aligned",
			timestamp:        1704067200, // 2024-01-01 00:00:00
			timeframeSeconds: 86400,      // 1 day
			expected:         1704067200, // 2024-01-01 00:00:00
		},
		{
			name:             "Zero timeframe",
			timestamp:        1704117000,
			timeframeSeconds: 0,
			expected:         1704117000,
		},
		{
			name:             "Negative timeframe",
			timestamp:        1704117000,
			timeframeSeconds: -86400,
			expected:         1704117000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aligner.AlignToTimeframe(tt.timestamp, tt.timeframeSeconds)
			if result != tt.expected {
				t.Errorf("AlignToTimeframe(%d, %d) = %d, expected %d",
					tt.timestamp, tt.timeframeSeconds, result, tt.expected)
			}
		})
	}
}

func TestTimestampAligner_GetAlignedTimestamp(t *testing.T) {
	aligner := NewTimestampAligner()
	converter := NewTimeframeConverter()

	tests := []struct {
		name         string
		barIndex     int
		dataLen      int
		barTimestamp int64
		secTimeframe string
		expected     int64
	}{
		{
			name:         "Valid bar, daily timeframe",
			barIndex:     5,
			dataLen:      10,
			barTimestamp: 1704117000, // 2024-01-01 14:30:00
			secTimeframe: "1D",
			expected:     1704067200, // 2024-01-01 00:00:00
		},
		{
			name:         "Valid bar, hourly timeframe",
			barIndex:     5,
			dataLen:      10,
			barTimestamp: 1704117000, // 2024-01-01 14:30:00
			secTimeframe: "1h",
			expected:     1704114000, // 2024-01-01 14:00:00
		},
		{
			name:         "Invalid bar index (negative)",
			barIndex:     -1,
			dataLen:      10,
			barTimestamp: 1704117000,
			secTimeframe: "1D",
			expected:     0,
		},
		{
			name:         "Invalid bar index (beyond length)",
			barIndex:     10,
			dataLen:      10,
			barTimestamp: 1704117000,
			secTimeframe: "1D",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				BarIndex: tt.barIndex,
				Data:     make([]OHLCV, tt.dataLen),
			}

			if tt.barIndex >= 0 && tt.barIndex < tt.dataLen {
				ctx.Data[tt.barIndex].Time = tt.barTimestamp
			}

			result := aligner.GetAlignedTimestamp(ctx, tt.secTimeframe, converter)
			if result != tt.expected {
				t.Errorf("GetAlignedTimestamp() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

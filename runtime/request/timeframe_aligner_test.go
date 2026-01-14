package request

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestTimeframeAligner_CurrentBarMode(t *testing.T) {
	aligner := NewTimeframeAligner()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{Time: 0},
			{Time: 86400},
			{Time: 172800},
			{Time: 259200},
		},
	}

	tests := []struct {
		name          string
		currentTime   int64
		expectedIndex int
		description   string
	}{
		{
			name:          "early in first period",
			currentTime:   1000,
			expectedIndex: 0,
			description:   "timestamp before first bar boundary returns first bar",
		},
		{
			name:          "within second period",
			currentTime:   100000,
			expectedIndex: 1,
			description:   "timestamp within second period returns second bar",
		},
		{
			name:          "at third period start",
			currentTime:   172800,
			expectedIndex: 2,
			description:   "timestamp at period start returns that period",
		},
		{
			name:          "within third period",
			currentTime:   200000,
			expectedIndex: 2,
			description:   "timestamp within third period returns third bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aligner.FindSecurityBarIndex(secCtx, tt.currentTime, true)

			if result != tt.expectedIndex {
				t.Errorf("%s: expected index %d, got %d",
					tt.description, tt.expectedIndex, result)
			}
		})
	}
}

func TestTimeframeAligner_PreviousCompletedBarMode(t *testing.T) {
	aligner := NewTimeframeAligner()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{Time: 0},
			{Time: 86400},
			{Time: 172800},
			{Time: 259200},
		},
	}

	tests := []struct {
		name          string
		currentTime   int64
		expectedIndex int
		description   string
	}{
		{
			name:          "early in first period",
			currentTime:   1000,
			expectedIndex: -1,
			description:   "no completed bar before first period",
		},
		{
			name:          "within second period",
			currentTime:   100000,
			expectedIndex: 0,
			description:   "within second period returns first completed bar",
		},
		{
			name:          "within third period",
			currentTime:   200000,
			expectedIndex: 1,
			description:   "within third period returns second completed bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aligner.FindSecurityBarIndex(secCtx, tt.currentTime, false)

			if result != tt.expectedIndex {
				t.Errorf("%s: expected index %d, got %d",
					tt.description, tt.expectedIndex, result)
			}
		})
	}
}

func TestTimeframeAligner_BeyondLastBar(t *testing.T) {
	aligner := NewTimeframeAligner()

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{Time: 0},
			{Time: 86400},
			{Time: 172800},
		},
	}

	t.Run("current bar mode beyond last", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, 999999, true)
		expected := 2

		if result != expected {
			t.Errorf("expected last bar index %d, got %d", expected, result)
		}
	})

	t.Run("previous bar mode beyond last", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, 999999, false)
		expected := 1

		if result != expected {
			t.Errorf("expected previous-to-last bar index %d, got %d", expected, result)
		}
	})
}

func TestTimeframeAligner_EmptyContext(t *testing.T) {
	aligner := NewTimeframeAligner()

	secCtx := &context.Context{
		Data: []context.OHLCV{},
	}

	t.Run("empty context current bar mode", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, 100000, true)

		if result != -1 {
			t.Errorf("expected -1 for empty context, got %d", result)
		}
	})

	t.Run("empty context previous bar mode", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, 100000, false)

		if result != -1 {
			t.Errorf("expected -1 for empty context, got %d", result)
		}
	})
}

func TestTimeframeAligner_RealWorldScenario_HourlyToDaily(t *testing.T) {
	aligner := NewTimeframeAligner()

	dec16Start := int64(1734307200)
	dec17Start := int64(1734393600)
	dec18Start := int64(1734480000)

	secCtx := &context.Context{
		Data: []context.OHLCV{
			{Time: dec16Start},
			{Time: dec17Start},
			{Time: dec18Start},
		},
	}

	dec17At10AM := dec17Start + 10*3600

	t.Run("lookahead on shows current day", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, dec17At10AM, true)
		expected := 1

		if result != expected {
			t.Errorf("lookahead=on at Dec 17 10AM should return Dec 17 (index %d), got %d",
				expected, result)
		}
	})

	t.Run("lookahead off shows previous day", func(t *testing.T) {
		result := aligner.FindSecurityBarIndex(secCtx, dec17At10AM, false)
		expected := 0

		if result != expected {
			t.Errorf("lookahead=off at Dec 17 10AM should return Dec 16 (index %d), got %d",
				expected, result)
		}
	})
}

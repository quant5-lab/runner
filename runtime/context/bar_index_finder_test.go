package context

import "testing"

func TestBarIndexFinder_FindContainingBar(t *testing.T) {
	finder := NewBarIndexFinder()

	data := []OHLCV{
		{Time: 0},
		{Time: 86400},
		{Time: 172800},
		{Time: 259200},
	}

	tests := []struct {
		name              string
		targetTimestamp   int64
		expectedIndex     int
		behaviorAssertion string
	}{
		{
			name:              "timestamp within first period",
			targetTimestamp:   1000,
			expectedIndex:     0,
			behaviorAssertion: "returns first bar when timestamp falls within it",
		},
		{
			name:              "timestamp within second period",
			targetTimestamp:   100000,
			expectedIndex:     1,
			behaviorAssertion: "returns second bar when timestamp falls within it",
		},
		{
			name:              "timestamp at period start",
			targetTimestamp:   172800,
			expectedIndex:     2,
			behaviorAssertion: "returns bar when timestamp matches period start exactly",
		},
		{
			name:              "timestamp beyond all bars",
			targetTimestamp:   999999,
			expectedIndex:     3,
			behaviorAssertion: "returns last bar when timestamp is beyond data",
		},
		{
			name:              "timestamp before first bar",
			targetTimestamp:   -1000,
			expectedIndex:     -1,
			behaviorAssertion: "returns -1 when timestamp is before first bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := finder.FindContainingBar(data, tt.targetTimestamp)

			if result != tt.expectedIndex {
				t.Errorf("%s: expected %d, got %d",
					tt.behaviorAssertion, tt.expectedIndex, result)
			}
		})
	}
}

func TestBarIndexFinder_EmptyData(t *testing.T) {
	finder := NewBarIndexFinder()

	result := finder.FindContainingBar([]OHLCV{}, 100000)

	if result != -1 {
		t.Errorf("empty data should return -1, got %d", result)
	}
}

func TestBarIndexFinder_SingleBar(t *testing.T) {
	finder := NewBarIndexFinder()

	data := []OHLCV{{Time: 100}}

	tests := []struct {
		timestamp int64
		expected  int
	}{
		{timestamp: 50, expected: -1},
		{timestamp: 100, expected: 0},
		{timestamp: 150, expected: 0},
	}

	for _, tt := range tests {
		result := finder.FindContainingBar(data, tt.timestamp)
		if result != tt.expected {
			t.Errorf("timestamp %d: expected %d, got %d",
				tt.timestamp, tt.expected, result)
		}
	}
}

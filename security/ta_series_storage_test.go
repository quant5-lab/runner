package security

import (
	"math"
	"testing"
)

func TestTASeriesStorage_InterfaceContract(t *testing.T) {
	tests := []struct {
		name    string
		storage TASeriesStorage
	}{
		{"ScalarStorage", NewScalarStorage()},
		{"SeriesStorage", NewSeriesStorage(10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.storage.Set(0, 100.0)
			if got := tt.storage.Get(0); got != 100.0 {
				t.Errorf("Set/Get contract broken: Set(0, 100.0) then Get(0) = %f", got)
			}
		})
	}
}

func TestScalarStorage_BehaviorContract(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "stores_last_written_value",
			test: func(t *testing.T) {
				storage := NewScalarStorage()
				storage.Set(5, 100.0)

				if got := storage.Get(5); got != 100.0 {
					t.Errorf("Get(5) = %f, want 100.0", got)
				}
			},
		},
		{
			name: "returns_nan_for_non_current_bar",
			test: func(t *testing.T) {
				storage := NewScalarStorage()
				storage.Set(5, 100.0)

				if got := storage.Get(4); !math.IsNaN(got) {
					t.Errorf("Get(4) after Set(5) = %f, want NaN", got)
				}
				if got := storage.Get(6); !math.IsNaN(got) {
					t.Errorf("Get(6) after Set(5) = %f, want NaN", got)
				}
			},
		},
		{
			name: "overwrites_on_repeated_set",
			test: func(t *testing.T) {
				storage := NewScalarStorage()
				storage.Set(3, 100.0)
				storage.Set(3, 200.0)

				if got := storage.Get(3); got != 200.0 {
					t.Errorf("After Set(3, 100) then Set(3, 200), Get(3) = %f, want 200.0", got)
				}
			},
		},
		{
			name: "only_remembers_last_bar",
			test: func(t *testing.T) {
				storage := NewScalarStorage()
				storage.Set(0, 100.0)
				storage.Set(1, 110.0)
				storage.Set(2, 120.0)

				if got := storage.Get(2); got != 120.0 {
					t.Errorf("Get(2) = %f, want 120.0", got)
				}
				if got := storage.Get(1); !math.IsNaN(got) {
					t.Errorf("Get(1) after Set(2) = %f, want NaN (scalar only remembers last)", got)
				}
				if got := storage.Get(0); !math.IsNaN(got) {
					t.Errorf("Get(0) after Set(2) = %f, want NaN (scalar only remembers last)", got)
				}
			},
		},
		{
			name: "negative_bar_index",
			test: func(t *testing.T) {
				storage := NewScalarStorage()
				storage.Set(-1, 100.0)

				if got := storage.Get(-1); got != 100.0 {
					t.Errorf("Get(-1) after Set(-1, 100) = %f, want 100.0", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestSeriesStorage_ForwardSeriesBufferSemantics(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "sequential_forward_advancement",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(5)

				storage.Set(0, 100.0)
				storage.Set(1, 110.0)
				storage.Set(2, 120.0)

				if got := storage.Get(0); got != 100.0 {
					t.Errorf("Get(0) = %f, want 100.0", got)
				}
				if got := storage.Get(1); got != 110.0 {
					t.Errorf("Get(1) = %f, want 110.0", got)
				}
				if got := storage.Get(2); got != 120.0 {
					t.Errorf("Get(2) = %f, want 120.0", got)
				}
			},
		},
		{
			name: "historical_lookback_after_forward",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(10)

				for bar := 0; bar < 6; bar++ {
					storage.Set(bar, float64(100+bar*10))
				}

				if got := storage.Get(5); got != 150.0 {
					t.Errorf("Get(5) = %f, want 150.0", got)
				}
				if got := storage.Get(3); got != 130.0 {
					t.Errorf("Get(3) after advancing to 5 = %f, want 130.0", got)
				}
				if got := storage.Get(0); got != 100.0 {
					t.Errorf("Get(0) after advancing to 5 = %f, want 100.0", got)
				}
			},
		},
		{
			name: "future_access_returns_nan",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(10)

				storage.Set(0, 100.0)
				storage.Set(1, 110.0)

				if got := storage.Get(2); !math.IsNaN(got) {
					t.Errorf("Get(2) when only 0,1 stored = %f, want NaN", got)
				}
				if got := storage.Get(5); !math.IsNaN(got) {
					t.Errorf("Get(5) when only 0,1 stored = %f, want NaN", got)
				}
			},
		},
		{
			name: "repeated_access_idempotent",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(5)

				storage.Set(0, 100.0)
				storage.Set(1, 110.0)
				storage.Set(2, 120.0)

				val1 := storage.Get(1)
				val2 := storage.Get(1)
				val3 := storage.Get(1)

				if val1 != val2 || val2 != val3 {
					t.Errorf("Repeated Get(1) not idempotent: %f, %f, %f", val1, val2, val3)
				}
			},
		},
		{
			name: "arbitrary_access_order",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(10)

				for bar := 0; bar < 8; bar++ {
					storage.Set(bar, float64(bar*100))
				}

				accessOrder := []int{7, 3, 5, 0, 4, 2, 6, 1}
				for _, bar := range accessOrder {
					expected := float64(bar * 100)
					if got := storage.Get(bar); got != expected {
						t.Errorf("Get(%d) in arbitrary order = %f, want %f", bar, got, expected)
					}
				}
			},
		},
		{
			name: "overwrite_at_current_position",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(5)

				storage.Set(0, 100.0)
				storage.Set(1, 110.0)
				storage.Set(1, 150.0)

				if got := storage.Get(1); got != 150.0 {
					t.Errorf("Get(1) after Set(1,110) then Set(1,150) = %f, want 150.0", got)
				}
			},
		},
		{
			name: "zero_capacity_clamped_to_one",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(0)
				storage.Set(0, 100.0)

				if got := storage.Get(0); got != 100.0 {
					t.Errorf("Zero capacity storage should clamp to 1: Get(0) = %f, want 100.0", got)
				}
			},
		},
		{
			name: "negative_capacity_clamped_to_one",
			test: func(t *testing.T) {
				storage := NewSeriesStorage(-5)
				storage.Set(0, 100.0)

				if got := storage.Get(0); got != 100.0 {
					t.Errorf("Negative capacity storage should clamp to 1: Get(0) = %f, want 100.0", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestSeriesStorage_LargeCapacityHandling(t *testing.T) {
	capacity := 1000
	storage := NewSeriesStorage(capacity)

	for bar := 0; bar < capacity; bar++ {
		storage.Set(bar, float64(bar*100))
	}

	testBars := []int{0, 100, 500, 999}
	for _, bar := range testBars {
		expected := float64(bar * 100)
		if got := storage.Get(bar); got != expected {
			t.Errorf("Large capacity: Get(%d) = %f, want %f", bar, got, expected)
		}
	}
}

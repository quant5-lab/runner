package request

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestSecurityBarMapper_Mode_ReflectsActiveBuildMethod(t *testing.T) {
	oneDailyBar := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}
	oneHourlyBar := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}

	tests := []struct {
		name     string
		build    func(*SecurityBarMapper)
		wantMode MappingMode
	}{
		{
			name:     "fresh mapper defaults to Downscaling",
			build:    func(_ *SecurityBarMapper) {},
			wantMode: ModeDownscaling,
		},
		{
			name: "BuildMappingWithDateFilter",
			build: func(m *SecurityBarMapper) {
				m.BuildMappingWithDateFilter(oneDailyBar, oneHourlyBar, DateRange{}, "UTC")
			},
			wantMode: ModeDownscaling,
		},
		{
			name: "BuildMappingForUpscaling",
			build: func(m *SecurityBarMapper) {
				m.BuildMappingForUpscaling(oneHourlyBar, oneDailyBar, "UTC")
			},
			wantMode: ModeUpscaling,
		},
		{
			name:     "BuildIdentityMapping",
			build:    func(m *SecurityBarMapper) { m.BuildIdentityMapping(100) },
			wantMode: ModeIdentity,
		},
		{
			name:     "BuildMappingFromTransform",
			build:    func(m *SecurityBarMapper) { m.BuildMappingFromTransform([]int{0, 0, 1, 2}) },
			wantMode: ModeTransformed,
		},
		{
			name: "subsequent build overwrites mode",
			build: func(m *SecurityBarMapper) {
				m.BuildMappingFromTransform([]int{0, 1})
				m.BuildIdentityMapping(10)
			},
			wantMode: ModeIdentity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSecurityBarMapper()
			tt.build(m)
			if got := m.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

func TestSecurityBarMapper_MainToSynthetic_ReflectsTransformMapping(t *testing.T) {
	tests := []struct {
		name        string
		build       func(*SecurityBarMapper)
		wantNil     bool
		wantMapping []int
	}{
		{
			name:    "fresh mapper",
			build:   func(_ *SecurityBarMapper) {},
			wantNil: true,
		},
		{
			name:    "BuildIdentityMapping produces nil",
			build:   func(m *SecurityBarMapper) { m.BuildIdentityMapping(100) },
			wantNil: true,
		},
		{
			name: "BuildIdentityMapping after BuildMappingFromTransform clears mapping",
			build: func(m *SecurityBarMapper) {
				m.BuildMappingFromTransform([]int{0, 1})
				m.BuildIdentityMapping(10)
			},
			wantNil: true,
		},
		{
			name: "BuildMappingWithDateFilter produces nil",
			build: func(m *SecurityBarMapper) {
				daily := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}
				hourly := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}
				m.BuildMappingWithDateFilter(daily, hourly, DateRange{}, "UTC")
			},
			wantNil: true,
		},
		{
			name: "BuildMappingForUpscaling produces nil",
			build: func(m *SecurityBarMapper) {
				daily := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}
				weekly := []context.OHLCV{{Time: parseTime("2025-01-01 14:30:00"), Close: 100}}
				m.BuildMappingForUpscaling(daily, weekly, "UTC")
			},
			wantNil: true,
		},
		{
			name:        "all pre-formation sentinels",
			build:       func(m *SecurityBarMapper) { m.BuildMappingFromTransform([]int{-1, -1, -1}) },
			wantMapping: []int{-1, -1, -1},
		},
		{
			name:        "partial pre-formation with multiple main bars per synthetic bar",
			build:       func(m *SecurityBarMapper) { m.BuildMappingFromTransform([]int{-1, -1, 0, 0, 1, 1, 1, 2}) },
			wantMapping: []int{-1, -1, 0, 0, 1, 1, 1, 2},
		},
		{
			name:        "single bar single synthetic bar",
			build:       func(m *SecurityBarMapper) { m.BuildMappingFromTransform([]int{0}) },
			wantMapping: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSecurityBarMapper()
			tt.build(m)
			got := m.MainToSynthetic()

			if tt.wantNil {
				if got != nil {
					t.Errorf("MainToSynthetic() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.wantMapping) {
				t.Fatalf("MainToSynthetic() len = %d, want %d", len(got), len(tt.wantMapping))
			}
			for i, want := range tt.wantMapping {
				if got[i] != want {
					t.Errorf("MainToSynthetic()[%d] = %d, want %d", i, got[i], want)
				}
			}
		})
	}
}

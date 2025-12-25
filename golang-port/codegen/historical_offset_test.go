package codegen

import (
	"testing"
)

// TestHistoricalOffset_Construction validates offset creation and basic properties
func TestHistoricalOffset_Construction(t *testing.T) {
	tests := []struct {
		name       string
		value      int
		wantValue  int
		wantIsZero bool
	}{
		{
			name:       "zero offset - current bar",
			value:      0,
			wantValue:  0,
			wantIsZero: true,
		},
		{
			name:       "positive offset 1 - one bar back",
			value:      1,
			wantValue:  1,
			wantIsZero: false,
		},
		{
			name:       "positive offset 4 - four bars back (BB7 case)",
			value:      4,
			wantValue:  4,
			wantIsZero: false,
		},
		{
			name:       "large positive offset - 100 bars back",
			value:      100,
			wantValue:  100,
			wantIsZero: false,
		},
		{
			name:       "negative offset - future bar (edge case)",
			value:      -1,
			wantValue:  -1,
			wantIsZero: false,
		},
		{
			name:       "large negative offset",
			value:      -50,
			wantValue:  -50,
			wantIsZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.value)

			if offset.Value() != tt.wantValue {
				t.Errorf("NewHistoricalOffset(%d).Value() = %d, want %d",
					tt.value, offset.Value(), tt.wantValue)
			}

			if offset.IsZero() != tt.wantIsZero {
				t.Errorf("NewHistoricalOffset(%d).IsZero() = %v, want %v",
					tt.value, offset.IsZero(), tt.wantIsZero)
			}
		})
	}
}

// TestHistoricalOffset_NoOffset validates zero offset factory method
func TestHistoricalOffset_NoOffset(t *testing.T) {
	offset := NoOffset()

	if offset.Value() != 0 {
		t.Errorf("NoOffset().Value() = %d, want 0", offset.Value())
	}

	if !offset.IsZero() {
		t.Error("NoOffset().IsZero() = false, want true")
	}
}

// TestHistoricalOffset_Add validates offset arithmetic
func TestHistoricalOffset_Add(t *testing.T) {
	tests := []struct {
		name       string
		baseOffset int
		addValue   int
		wantSum    int
	}{
		{
			name:       "zero offset + zero",
			baseOffset: 0,
			addValue:   0,
			wantSum:    0,
		},
		{
			name:       "zero offset + positive",
			baseOffset: 0,
			addValue:   5,
			wantSum:    5,
		},
		{
			name:       "positive offset + positive",
			baseOffset: 4,
			addValue:   10,
			wantSum:    14,
		},
		{
			name:       "offset 4 + period 20 - BB7 warmup case",
			baseOffset: 4,
			addValue:   19, // period - 1
			wantSum:    23,
		},
		{
			name:       "offset 10 + period 50",
			baseOffset: 10,
			addValue:   49, // period - 1
			wantSum:    59,
		},
		{
			name:       "negative offset + positive",
			baseOffset: -5,
			addValue:   10,
			wantSum:    5,
		},
		{
			name:       "positive offset + negative",
			baseOffset: 10,
			addValue:   -3,
			wantSum:    7,
		},
		{
			name:       "negative offset + negative",
			baseOffset: -5,
			addValue:   -3,
			wantSum:    -8,
		},
		{
			name:       "large offset arithmetic",
			baseOffset: 100,
			addValue:   50,
			wantSum:    150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.baseOffset)
			sum := offset.Add(tt.addValue)

			if sum != tt.wantSum {
				t.Errorf("NewHistoricalOffset(%d).Add(%d) = %d, want %d",
					tt.baseOffset, tt.addValue, sum, tt.wantSum)
			}
		})
	}
}

// TestHistoricalOffset_FormatLoopAccess validates loop expression generation
func TestHistoricalOffset_FormatLoopAccess(t *testing.T) {
	tests := []struct {
		name       string
		offset     int
		loopVar    string
		wantFormat string
	}{
		{
			name:       "zero offset with j - no modification",
			offset:     0,
			loopVar:    "j",
			wantFormat: "j",
		},
		{
			name:       "zero offset with i - no modification",
			offset:     0,
			loopVar:    "i",
			wantFormat: "i",
		},
		{
			name:       "offset 1 with j",
			offset:     1,
			loopVar:    "j",
			wantFormat: "j+1",
		},
		{
			name:       "offset 4 with j - BB7 case",
			offset:     4,
			loopVar:    "j",
			wantFormat: "j+4",
		},
		{
			name:       "offset 10 with idx",
			offset:     10,
			loopVar:    "idx",
			wantFormat: "idx+10",
		},
		{
			name:       "large offset with j",
			offset:     100,
			loopVar:    "j",
			wantFormat: "j+100",
		},
		{
			name:       "offset 2 with different var name",
			offset:     2,
			loopVar:    "loopIndex",
			wantFormat: "loopIndex+2",
		},
		{
			name:       "negative offset - edge case",
			offset:     -1,
			loopVar:    "j",
			wantFormat: "j+-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.offset)
			formatted := offset.FormatLoopAccess(tt.loopVar)

			if formatted != tt.wantFormat {
				t.Errorf("NewHistoricalOffset(%d).FormatLoopAccess(%q) = %q, want %q",
					tt.offset, tt.loopVar, formatted, tt.wantFormat)
			}
		})
	}
}

// TestHistoricalOffset_IsZero_BoundaryConditions validates zero detection edge cases
func TestHistoricalOffset_IsZero_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		want   bool
	}{
		{"exactly zero", 0, true},
		{"one above zero", 1, false},
		{"one below zero", -1, false},
		{"large positive", 1000, false},
		{"large negative", -1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.offset)
			if got := offset.IsZero(); got != tt.want {
				t.Errorf("NewHistoricalOffset(%d).IsZero() = %v, want %v",
					tt.offset, got, tt.want)
			}
		})
	}
}

// TestHistoricalOffset_Immutability validates offset values don't mutate
func TestHistoricalOffset_Immutability(t *testing.T) {
	original := NewHistoricalOffset(5)
	originalValue := original.Value()

	// Perform operations that should not mutate original
	_ = original.Add(10)
	_ = original.FormatLoopAccess("j")
	_ = original.IsZero()

	if original.Value() != originalValue {
		t.Errorf("HistoricalOffset mutated: was %d, now %d",
			originalValue, original.Value())
	}
}

// TestHistoricalOffset_CompositeOperations validates multiple operations in sequence
func TestHistoricalOffset_CompositeOperations(t *testing.T) {
	tests := []struct {
		name              string
		offset            int
		operations        func(HistoricalOffset) []interface{}
		wantResults       []interface{}
	}{
		{
			name:   "zero offset - all operations",
			offset: 0,
			operations: func(o HistoricalOffset) []interface{} {
				return []interface{}{
					o.Value(),
					o.IsZero(),
					o.Add(10),
					o.FormatLoopAccess("j"),
				}
			},
			wantResults: []interface{}{0, true, 10, "j"},
		},
		{
			name:   "offset 4 - typical BB7 usage",
			offset: 4,
			operations: func(o HistoricalOffset) []interface{} {
				return []interface{}{
					o.Value(),
					o.IsZero(),
					o.Add(19), // period 20 - 1
					o.FormatLoopAccess("j"),
				}
			},
			wantResults: []interface{}{4, false, 23, "j+4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.offset)
			results := tt.operations(offset)

			for i, want := range tt.wantResults {
				if results[i] != want {
					t.Errorf("Operation %d: got %v, want %v", i, results[i], want)
				}
			}
		})
	}
}

// TestHistoricalOffset_EdgeCaseFormulas validates correct formula application
func TestHistoricalOffset_EdgeCaseFormulas(t *testing.T) {
	tests := []struct {
		name        string
		baseOffset  int
		period      int
		wantWarmup  int // period - 1 + baseOffset
		wantInitial string // Format for initial value access
	}{
		{
			name:        "BB7 bug case: period=20, offset=4",
			baseOffset:  4,
			period:      20,
			wantWarmup:  23,
			wantInitial: "j+4",
		},
		{
			name:        "no offset: period=20, offset=0",
			baseOffset:  0,
			period:      20,
			wantWarmup:  19,
			wantInitial: "j",
		},
		{
			name:        "large period: period=200, offset=4",
			baseOffset:  4,
			period:      200,
			wantWarmup:  203,
			wantInitial: "j+4",
		},
		{
			name:        "minimal: period=1, offset=0",
			baseOffset:  0,
			period:      1,
			wantWarmup:  0,
			wantInitial: "j",
		},
		{
			name:        "minimal with offset: period=1, offset=5",
			baseOffset:  5,
			period:      1,
			wantWarmup:  5,
			wantInitial: "j+5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.baseOffset)

			// Test warmup calculation
			warmup := offset.Add(tt.period - 1)
			if warmup != tt.wantWarmup {
				t.Errorf("Warmup calculation: offset.Add(%d-1) = %d, want %d",
					tt.period, warmup, tt.wantWarmup)
			}

			// Test loop access formatting
			formatted := offset.FormatLoopAccess("j")
			if formatted != tt.wantInitial {
				t.Errorf("Loop access: FormatLoopAccess(\"j\") = %q, want %q",
					formatted, tt.wantInitial)
			}
		})
	}
}

// BenchmarkHistoricalOffset_Operations measures performance of offset operations
func BenchmarkHistoricalOffset_Operations(b *testing.B) {
	b.Run("NewHistoricalOffset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewHistoricalOffset(4)
		}
	})

	b.Run("Value", func(b *testing.B) {
		offset := NewHistoricalOffset(4)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = offset.Value()
		}
	})

	b.Run("IsZero", func(b *testing.B) {
		offset := NewHistoricalOffset(4)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = offset.IsZero()
		}
	})

	b.Run("Add", func(b *testing.B) {
		offset := NewHistoricalOffset(4)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = offset.Add(19)
		}
	})

	b.Run("FormatLoopAccess", func(b *testing.B) {
		offset := NewHistoricalOffset(4)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = offset.FormatLoopAccess("j")
		}
	})
}

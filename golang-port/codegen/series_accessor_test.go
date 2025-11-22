package codegen

import (
	"testing"
)

func TestSeriesVariableAccessor(t *testing.T) {
	tests := []struct {
		name             string
		sourceExpr       string
		shouldMatch      bool
		expectedVarName  string
		expectedAccess   string
		requiresNaNCheck bool
	}{
		{
			name:             "Simple series variable",
			sourceExpr:       "cagr5Series.Get(0)",
			shouldMatch:      true,
			expectedVarName:  "cagr5",
			expectedAccess:   "cagr5Series.Get(10)",
			requiresNaNCheck: true,
		},
		{
			name:             "Series with underscore",
			sourceExpr:       "ema_60Series.Get(0)",
			shouldMatch:      true,
			expectedVarName:  "ema_60",
			expectedAccess:   "ema_60Series.Get(5)",
			requiresNaNCheck: true,
		},
		{
			name:             "Series with numbers",
			sourceExpr:       "var123Series.Get(0)",
			shouldMatch:      true,
			expectedVarName:  "var123",
			expectedAccess:   "var123Series.Get(0)",
			requiresNaNCheck: true,
		},
		{
			name:             "Series starting with underscore",
			sourceExpr:       "_privateSeries.Get(0)",
			shouldMatch:      true,
			expectedVarName:  "_private",
			expectedAccess:   "_privateSeries.Get(20)",
			requiresNaNCheck: true,
		},
		{
			name:        "OHLCV field should not match",
			sourceExpr:  "bar.Close",
			shouldMatch: false,
		},
		{
			name:        "Plain identifier should not match",
			sourceExpr:  "close",
			shouldMatch: false,
		},
		{
			name:        "GetCurrent instead of Get",
			sourceExpr:  "cagr5Series.GetCurrent()",
			shouldMatch: false,
		},
		{
			name:        "Missing Series suffix",
			sourceExpr:  "cagr5.Get(0)",
			shouldMatch: false,
		},
		{
			name:        "Invalid identifier (starts with number)",
			sourceExpr:  "123Series.Get(0)",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewSeriesVariableAccessor(tt.sourceExpr)

			if tt.shouldMatch {
				if accessor == nil {
					t.Errorf("Expected accessor to be created, got nil")
					return
				}

				if !accessor.IsApplicable(tt.sourceExpr) {
					t.Errorf("IsApplicable() = false, want true")
				}

				if got := accessor.GetSourceIdentifier(); got != tt.expectedVarName {
					t.Errorf("GetSourceIdentifier() = %q, want %q", got, tt.expectedVarName)
				}

				if got := accessor.RequiresNaNCheck(); got != tt.requiresNaNCheck {
					t.Errorf("RequiresNaNCheck() = %v, want %v", got, tt.requiresNaNCheck)
				}

				offset := "10"
				if len(tt.expectedAccess) > 0 {
					// Extract offset from expected access for consistent testing
					if tt.sourceExpr == "cagr5Series.Get(0)" {
						offset = "10"
					} else if tt.sourceExpr == "ema_60Series.Get(0)" {
						offset = "5"
					} else if tt.sourceExpr == "var123Series.Get(0)" {
						offset = "0"
					} else if tt.sourceExpr == "_privateSeries.Get(0)" {
						offset = "20"
					}

					if got := accessor.GetAccessExpression(offset); got != tt.expectedAccess {
						t.Errorf("GetAccessExpression(%q) = %q, want %q", offset, got, tt.expectedAccess)
					}
				}
			} else {
				if accessor != nil {
					t.Errorf("Expected accessor to be nil, got %+v", accessor)
				}
			}
		})
	}
}

func TestOHLCVFieldAccessor(t *testing.T) {
	tests := []struct {
		name             string
		sourceExpr       string
		expectedField    string
		expectedAccess   string
		requiresNaNCheck bool
	}{
		{
			name:             "Simple close field",
			sourceExpr:       "close",
			expectedField:    "close",
			expectedAccess:   "ctx.Data[ctx.BarIndex-10].close",
			requiresNaNCheck: false,
		},
		{
			name:             "Bar.Close with dot notation",
			sourceExpr:       "bar.Close",
			expectedField:    "Close",
			expectedAccess:   "ctx.Data[ctx.BarIndex-5].Close",
			requiresNaNCheck: false,
		},
		{
			name:             "Nested dot notation",
			sourceExpr:       "ctx.Data.Close",
			expectedField:    "Close",
			expectedAccess:   "ctx.Data[ctx.BarIndex-0].Close",
			requiresNaNCheck: false,
		},
		{
			name:             "High field",
			sourceExpr:       "high",
			expectedField:    "high",
			expectedAccess:   "ctx.Data[ctx.BarIndex-20].high",
			requiresNaNCheck: false,
		},
		{
			name:             "Low field",
			sourceExpr:       "low",
			expectedField:    "low",
			expectedAccess:   "ctx.Data[ctx.BarIndex-15].low",
			requiresNaNCheck: false,
		},
		{
			name:             "Open field",
			sourceExpr:       "open",
			expectedField:    "open",
			expectedAccess:   "ctx.Data[ctx.BarIndex-1].open",
			requiresNaNCheck: false,
		},
		{
			name:             "Volume field",
			sourceExpr:       "volume",
			expectedField:    "volume",
			expectedAccess:   "ctx.Data[ctx.BarIndex-7].volume",
			requiresNaNCheck: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewOHLCVFieldAccessor(tt.sourceExpr)

			if accessor == nil {
				t.Fatal("Expected accessor to be created, got nil")
			}

			if !accessor.IsApplicable(tt.sourceExpr) {
				t.Errorf("IsApplicable() = false, want true")
			}

			if got := accessor.GetSourceIdentifier(); got != tt.expectedField {
				t.Errorf("GetSourceIdentifier() = %q, want %q", got, tt.expectedField)
			}

			if got := accessor.RequiresNaNCheck(); got != tt.requiresNaNCheck {
				t.Errorf("RequiresNaNCheck() = %v, want %v", got, tt.requiresNaNCheck)
			}

			// Extract offset from expected access
			var offset string
			switch tt.sourceExpr {
			case "close":
				offset = "10"
			case "bar.Close":
				offset = "5"
			case "ctx.Data.Close":
				offset = "0"
			case "high":
				offset = "20"
			case "low":
				offset = "15"
			case "open":
				offset = "1"
			case "volume":
				offset = "7"
			}

			if got := accessor.GetAccessExpression(offset); got != tt.expectedAccess {
				t.Errorf("GetAccessExpression(%q) = %q, want %q", offset, got, tt.expectedAccess)
			}
		})
	}
}

func TestCreateSeriesAccessor(t *testing.T) {
	tests := []struct {
		name               string
		sourceExpr         string
		expectedType       string // "series" or "ohlcv"
		expectedIdentifier string
		requiresNaNCheck   bool
	}{
		{
			name:               "Series variable",
			sourceExpr:         "cagr5Series.Get(0)",
			expectedType:       "series",
			expectedIdentifier: "cagr5",
			requiresNaNCheck:   true,
		},
		{
			name:               "OHLCV close",
			sourceExpr:         "close",
			expectedType:       "ohlcv",
			expectedIdentifier: "close",
			requiresNaNCheck:   false,
		},
		{
			name:               "OHLCV with dot notation",
			sourceExpr:         "bar.High",
			expectedType:       "ohlcv",
			expectedIdentifier: "High",
			requiresNaNCheck:   false,
		},
		{
			name:               "Complex series name",
			sourceExpr:         "my_ema_20Series.Get(0)",
			expectedType:       "series",
			expectedIdentifier: "my_ema_20",
			requiresNaNCheck:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := CreateSeriesAccessor(tt.sourceExpr)

			if accessor == nil {
				t.Fatal("Expected accessor to be created, got nil")
			}

			if got := accessor.GetSourceIdentifier(); got != tt.expectedIdentifier {
				t.Errorf("GetSourceIdentifier() = %q, want %q", got, tt.expectedIdentifier)
			}

			if got := accessor.RequiresNaNCheck(); got != tt.requiresNaNCheck {
				t.Errorf("RequiresNaNCheck() = %v, want %v", got, tt.requiresNaNCheck)
			}

			// Verify the type by checking the access expression format
			accessExpr := accessor.GetAccessExpression("10")
			isSeries := false
			if _, ok := accessor.(*SeriesVariableAccessor); ok {
				isSeries = true
			}

			if tt.expectedType == "series" && !isSeries {
				t.Errorf("Expected SeriesVariableAccessor, got different type")
			}
			if tt.expectedType == "ohlcv" && isSeries {
				t.Errorf("Expected OHLCVFieldAccessor, got SeriesVariableAccessor")
			}

			// Verify access expression format
			if tt.expectedType == "series" {
				expectedPattern := tt.expectedIdentifier + "Series.Get(10)"
				if accessExpr != expectedPattern {
					t.Errorf("GetAccessExpression(10) = %q, want %q", accessExpr, expectedPattern)
				}
			} else {
				expectedPattern := "ctx.Data[ctx.BarIndex-10]." + tt.expectedIdentifier
				if accessExpr != expectedPattern {
					t.Errorf("GetAccessExpression(10) = %q, want %q", accessExpr, expectedPattern)
				}
			}
		})
	}
}

func TestAccessorEdgeCases(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		accessor := CreateSeriesAccessor("")
		if accessor == nil {
			t.Fatal("Expected accessor to be created even for empty string")
		}
		// Should fall back to OHLCV with empty field name
		if _, ok := accessor.(*OHLCVFieldAccessor); !ok {
			t.Error("Expected OHLCVFieldAccessor for empty string")
		}
	})

	t.Run("Whitespace", func(t *testing.T) {
		accessor := CreateSeriesAccessor("  ")
		if accessor == nil {
			t.Fatal("Expected accessor to be created")
		}
		// Should fall back to OHLCV
		if _, ok := accessor.(*OHLCVFieldAccessor); !ok {
			t.Error("Expected OHLCVFieldAccessor for whitespace")
		}
	})

	t.Run("Special characters in expression", func(t *testing.T) {
		accessor := CreateSeriesAccessor("some$weird.field")
		if accessor == nil {
			t.Fatal("Expected accessor to be created")
		}
		// Should extract "field" as field name
		if got := accessor.GetSourceIdentifier(); got != "field" {
			t.Errorf("GetSourceIdentifier() = %q, want %q", got, "field")
		}
	})

	t.Run("Series-like but invalid pattern", func(t *testing.T) {
		accessor := CreateSeriesAccessor("Series.Get(0)")
		// Missing variable name before "Series", should fall back to OHLCV
		if _, ok := accessor.(*OHLCVFieldAccessor); !ok {
			t.Error("Expected OHLCVFieldAccessor for invalid Series pattern")
		}
	})
}

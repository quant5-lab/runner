package codegen

import (
	"strings"
	"testing"
)

func TestSeriesWindowExtractor_SeriesAccess(t *testing.T) {
	extractor := NewSeriesWindowExtractor()
	indenter := func() string { return "\t" }

	code := extractor.GenerateExtractionCode("closeSeries", "i+1", indenter)

	expectedPatterns := []string{
		"sourceWindow := make([]float64, i+1)",
		"for j := 0; j < i+1; j++",
		"closeSeries.Get(j)",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing pattern: %s\nGenerated:\n%s", pattern, code)
		}
	}
}

func TestSeriesWindowExtractor_BarFieldAccess(t *testing.T) {
	extractor := NewSeriesWindowExtractor()
	indenter := func() string { return "\t" }

	code := extractor.GenerateExtractionCode("bar.Close", "i+1", indenter)

	if !strings.Contains(code, "ctx.Data[j].Close") {
		t.Errorf("Expected bar.Close → ctx.Data[j].Close conversion\nGenerated:\n%s", code)
	}
}

func TestSeriesWindowExtractor_SeriesGetCurrentAccess(t *testing.T) {
	extractor := NewSeriesWindowExtractor()

	access := extractor.buildHistoricalAccess("priceSeries.GetCurrent()", "j")
	expected := "priceSeries.Get(j)"

	if access != expected {
		t.Errorf("Expected %s, got %s", expected, access)
	}
}

func TestSeriesWindowExtractor_EdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		sourceExpression   string
		periodExpression   string
		mustContainPattern []string
		shouldError        bool
	}{
		{
			name:             "Empty source expression",
			sourceExpression: "",
			periodExpression: "i+1",
			mustContainPattern: []string{
				"sourceWindow := make([]float64, i+1)",
			},
			shouldError: false,
		},
		{
			name:             "Zero period expression",
			sourceExpression: "closeSeries",
			periodExpression: "0",
			mustContainPattern: []string{
				"sourceWindow := make([]float64, 0)",
			},
			shouldError: false,
		},
		{
			name:             "Complex period expression",
			sourceExpression: "priceSeries",
			periodExpression: "max(10, i+1)",
			mustContainPattern: []string{
				"sourceWindow := make([]float64, max(10, i+1))",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSeriesWindowExtractor()
			indenter := func() string { return "\t" }

			code := extractor.GenerateExtractionCode(tt.sourceExpression, tt.periodExpression, indenter)

			for _, pattern := range tt.mustContainPattern {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern %q\nGenerated:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestSeriesWindowExtractor_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name             string
		sourceExpression string
		expectedAccess   string
	}{
		{
			name:             "Single character series",
			sourceExpression: "xSeries",
			expectedAccess:   "xSeries.Get(j)",
		},
		{
			name:             "Long identifier series",
			sourceExpression: "veryLongSeriesNameForTestingSeries",
			expectedAccess:   "veryLongSeriesNameForTestingSeries.Get(j)",
		},
		{
			name:             "Underscore prefix series",
			sourceExpression: "_privateSeries",
			expectedAccess:   "_privateSeries.Get(j)",
		},
		{
			name:             "Numeric suffix series",
			sourceExpression: "series1Series",
			expectedAccess:   "series1Series.Get(j)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSeriesWindowExtractor()
			indenter := func() string { return "\t" }

			code := extractor.GenerateExtractionCode(tt.sourceExpression, "i+1", indenter)

			if !strings.Contains(code, tt.expectedAccess) {
				t.Errorf("Missing expected access %q\nGenerated:\n%s", tt.expectedAccess, code)
			}
		})
	}
}

func TestSeriesWindowExtractor_BarFieldConversions(t *testing.T) {
	tests := []struct {
		name           string
		barField       string
		expectedAccess string
	}{
		{
			name:           "bar.Close",
			barField:       "bar.Close",
			expectedAccess: "ctx.Data[j].Close",
		},
		{
			name:           "bar.Open",
			barField:       "bar.Open",
			expectedAccess: "ctx.Data[j].Open",
		},
		{
			name:           "bar.High",
			barField:       "bar.High",
			expectedAccess: "ctx.Data[j].High",
		},
		{
			name:           "bar.Low",
			barField:       "bar.Low",
			expectedAccess: "ctx.Data[j].Low",
		},
		{
			name:           "bar.Volume",
			barField:       "bar.Volume",
			expectedAccess: "ctx.Data[j].Volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSeriesWindowExtractor()
			indenter := func() string { return "\t" }

			code := extractor.GenerateExtractionCode(tt.barField, "i+1", indenter)

			if !strings.Contains(code, tt.expectedAccess) {
				t.Errorf("Missing expected bar field access %q\nGenerated:\n%s", tt.expectedAccess, code)
			}
		})
	}
}

func TestSeriesWindowExtractor_GetCurrentConversions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		indexVar     string
		expectedCode string
	}{
		{
			name:         "Simple GetCurrent",
			input:        "priceSeries.GetCurrent()",
			indexVar:     "j",
			expectedCode: "priceSeries.Get(j)",
		},
		{
			name:         "Different index variable",
			input:        "closeSeries.GetCurrent()",
			indexVar:     "idx",
			expectedCode: "closeSeries.Get(idx)",
		},
		{
			name:         "Long series name",
			input:        "myCustomIndicatorSeries.GetCurrent()",
			indexVar:     "j",
			expectedCode: "myCustomIndicatorSeries.Get(j)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSeriesWindowExtractor()

			result := extractor.buildHistoricalAccess(tt.input, tt.indexVar)

			if result != tt.expectedCode {
				t.Errorf("Expected %q, got %q", tt.expectedCode, result)
			}
		})
	}
}

func TestSeriesWindowExtractor_IndentationConsistency(t *testing.T) {
	tests := []struct {
		name     string
		indenter func() string
	}{
		{
			name:     "Single tab",
			indenter: func() string { return "\t" },
		},
		{
			name:     "Double tab",
			indenter: func() string { return "\t\t" },
		},
		{
			name:     "Four spaces",
			indenter: func() string { return "    " },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewSeriesWindowExtractor()
			code := extractor.GenerateExtractionCode("closeSeries", "i+1", tt.indenter)

			lines := strings.Split(code, "\n")
			for i, line := range lines {
				if len(line) == 0 {
					continue
				}
				if !strings.HasPrefix(line, tt.indenter()) {
					t.Errorf("Line %d missing proper indentation: %q", i+1, line)
				}
			}
		})
	}
}

func TestSeriesWindowExtractor_CodeStructureValidation(t *testing.T) {
	extractor := NewSeriesWindowExtractor()
	indenter := func() string { return "\t" }

	code := extractor.GenerateExtractionCode("closeSeries", "i+1", indenter)

	requiredPatterns := []string{
		"sourceWindow := make([]float64",
		"for j := 0; j <",
		"sourceWindow[j] =",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing required code structure: %q\nGenerated:\n%s", pattern, code)
		}
	}

	if strings.Count(code, "sourceWindow := make(") != 1 {
		t.Error("Window allocation should occur exactly once")
	}
}

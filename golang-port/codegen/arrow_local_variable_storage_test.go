package codegen

import (
	"strings"
	"testing"
)

func TestArrowLocalVariableStorage_GenerateScalarDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		exprCode string
		want     string
	}{
		{
			name:     "simple assignment",
			varName:  "up",
			exprCode: "change(high)",
			want:     "\tup := change(high)\n",
		},
		{
			name:     "complex expression",
			varName:  "truerange",
			exprCode: "rma(tr, len)",
			want:     "\ttruerange := rma(tr, len)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateScalarDeclaration(tt.varName, tt.exprCode)

			if got != tt.want {
				t.Errorf("GenerateScalarDeclaration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrowLocalVariableStorage_GenerateSeriesStorage(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		want    string
	}{
		{
			name:    "simple variable",
			varName: "up",
			want:    "\tupSeries.Set(up)\n",
		},
		{
			name:    "variable with underscore",
			varName: "true_range",
			want:    "\ttrue_rangeSeries.Set(true_range)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateSeriesStorage(tt.varName)

			if got != tt.want {
				t.Errorf("GenerateSeriesStorage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrowLocalVariableStorage_GenerateDualStorage(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		exprCode string
		want     []string
	}{
		{
			name:     "complete dual storage",
			varName:  "up",
			exprCode: "change(high)",
			want: []string{
				"up := change(high)",
				"upSeries.Set(up)",
			},
		},
		{
			name:     "complex expression dual storage",
			varName:  "truerange",
			exprCode: "rma(tr, len)",
			want: []string{
				"truerange := rma(tr, len)",
				"truerangeSeries.Set(truerange)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateDualStorage(tt.varName, tt.exprCode)

			for _, expectedLine := range tt.want {
				if !strings.Contains(got, expectedLine) {
					t.Errorf("GenerateDualStorage() missing expected line: %q\nGot: %q", expectedLine, got)
				}
			}

			lines := strings.Split(strings.TrimSpace(got), "\n")
			if len(lines) != 2 {
				t.Errorf("GenerateDualStorage() expected 2 lines, got %d", len(lines))
			}
		})
	}
}

func TestArrowLocalVariableStorage_GenerateTupleDualStorage(t *testing.T) {
	tests := []struct {
		name     string
		varNames []string
		exprCode string
		want     []string
	}{
		{
			name:     "two variable tuple",
			varNames: []string{"plus", "minus"},
			exprCode: "dirmov(len)",
			want: []string{
				"temp_plus, temp_minus := dirmov(len)",
				"plus := temp_plus",
				"plusSeries.Set(plus)",
				"minus := temp_minus",
				"minusSeries.Set(minus)",
			},
		},
		{
			name:     "three variable tuple",
			varNames: []string{"adx", "up", "down"},
			exprCode: "calculateADX(period)",
			want: []string{
				"temp_adx, temp_up, temp_down := calculateADX(period)",
				"adx := temp_adx",
				"adxSeries.Set(adx)",
				"up := temp_up",
				"upSeries.Set(up)",
				"down := temp_down",
				"downSeries.Set(down)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage("\t")
			got := storage.GenerateTupleDualStorage(tt.varNames, tt.exprCode)

			for _, expectedLine := range tt.want {
				if !strings.Contains(got, expectedLine) {
					t.Errorf("GenerateTupleDualStorage() missing expected line: %q\nGot: %q", expectedLine, got)
				}
			}

			expectedLineCount := 1 + (len(tt.varNames) * 2)
			gotLines := strings.Split(strings.TrimSpace(got), "\n")
			if len(gotLines) != expectedLineCount {
				t.Errorf("GenerateTupleDualStorage() expected %d lines, got %d", expectedLineCount, len(gotLines))
			}
		})
	}
}

func TestArrowLocalVariableStorage_DualAccessPattern(t *testing.T) {
	t.Run("scalar first then series", func(t *testing.T) {
		storage := NewArrowLocalVariableStorage("\t")
		code := storage.GenerateDualStorage("up", "change(high)")

		lines := strings.Split(strings.TrimSpace(code), "\n")
		if len(lines) != 2 {
			t.Fatalf("Expected 2 lines, got %d", len(lines))
		}

		if !strings.Contains(lines[0], "up := change(high)") {
			t.Errorf("First line should be scalar declaration, got: %s", lines[0])
		}

		if !strings.Contains(lines[1], "upSeries.Set(up)") {
			t.Errorf("Second line should be series storage, got: %s", lines[1])
		}
	})
}

func TestArrowLocalVariableStorage_Indentation(t *testing.T) {
	tests := []struct {
		name   string
		indent string
	}{
		{
			name:   "single tab",
			indent: "\t",
		},
		{
			name:   "two tabs",
			indent: "\t\t",
		},
		{
			name:   "four spaces",
			indent: "    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewArrowLocalVariableStorage(tt.indent)
			code := storage.GenerateScalarDeclaration("test", "value")

			if !strings.HasPrefix(code, tt.indent) {
				t.Errorf("Code should start with indent %q, got: %q", tt.indent, code)
			}
		})
	}
}

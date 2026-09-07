package regression

import (
	"testing"
)

// TestArrayDrawingConstructors_CodegenBuild verifies that Pine scripts declaring
// drawing-type arrays (array.new_line, array.new_label, etc.) survive the full
// codegen -> compile pipeline.  Each case exercises a distinct aspect of the
// drawing array declaration semantics: empty arrays, pre-sized arrays, zero size,
// dynamic-expression size, multiple types, and mixed numeric/drawing arrays.
func TestArrayDrawingConstructors_CodegenBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name string
		pine string
	}{
		{
			name: "empty_label_array",
			pine: "//@version=5\nindicator(\"Label Array\", overlay=true)\nlabels = array.new_label()\nplot(close)\n",
		},
		{
			name: "sized_line_array",
			pine: "//@version=5\nindicator(\"Line Array\", overlay=true)\nlines = array.new_line(10)\nplot(close)\n",
		},
		{
			name: "empty_box_array",
			pine: "//@version=5\nindicator(\"Box Array\", overlay=true)\nboxes = array.new_box()\nplot(close)\n",
		},
		{
			name: "sized_table_array",
			pine: "//@version=5\nindicator(\"Table Array\", overlay=true)\ntables = array.new_table(5)\nplot(close)\n",
		},
		{
			name: "empty_linefill_array",
			pine: "//@version=5\nindicator(\"Linefill Array\", overlay=true)\nfills = array.new_linefill()\nplot(close)\n",
		},
		{
			name: "mixed_drawing_arrays",
			pine: "//@version=5\nindicator(\"Mixed Drawing Arrays\", overlay=true)\nlabels = array.new_label(3)\nlines = array.new_line(5)\nboxes = array.new_box()\ntables = array.new_table(2)\nfills = array.new_linefill(1)\nplot(close)\n",
		},
		{
			name: "drawing_and_numeric_arrays",
			pine: "//@version=5\nindicator(\"Mixed Array Types\", overlay=true)\nprices = array.new_float(10, close)\nlabels = array.new_label(5)\nvolumes = array.new_int(10)\nlines = array.new_line()\nflags = array.new_bool(3)\nboxes = array.new_box(2)\nplot(close)\n",
		},
		{
			name: "zero_size_drawing_arrays",
			pine: "//@version=5\nindicator(\"Zero Size Arrays\", overlay=true)\nlabels = array.new_label(0)\nlines = array.new_line(0)\nboxes = array.new_box(0)\nplot(close)\n",
		},
		{
			name: "dynamic_size_expression",
			pine: "//@version=5\nindicator(\"Dynamic Size\", overlay=true)\nsizeParam = 10\nlabels = array.new_label(sizeParam)\nlines = array.new_line(sizeParam * 2)\nboxes = array.new_box(sizeParam + 5)\nplot(close)\n",
		},
	}

	projectRoot := projectRootFromCwd()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if _, ok := codegenAndBuild(t, tmpDir, tt.name, tt.pine, projectRoot); !ok {
				t.Fail()
			}
		})
	}
}

package regression

import (
	"testing"
)

// TestDrawingNamespaceConstants_CodegenBuild verifies that all 8 drawing-related
// Pine namespaces (extend, line, label, xloc, size, shape, location, hline)
// survive the full codegen -> compile pipeline when their constants appear in
// assignment, switch, and conditional-expression contexts.
func TestDrawingNamespaceConstants_CodegenBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name string
		pine string
	}{
		// --- extend namespace ---
		{
			name: "extend_direct_assignment",
			pine: `//@version=5
indicator("extend constants", overlay=true)
ext1 = extend.none
ext2 = extend.right
ext3 = extend.left
ext4 = extend.both
plot(ext1 + ext2 + ext3 + ext4)
`,
		},
		{
			name: "extend_switch_selection",
			pine: `//@version=5
indicator("extend switch", overlay=true)
mode = input.string("Right", options=["None","Left","Right","Both"])
ext = switch mode
    "None"  => extend.none
    "Left"  => extend.left
    "Right" => extend.right
    "Both"  => extend.both
plot(ext)
`,
		},
		{
			name: "extend_equality_comparison",
			pine: `//@version=5
indicator("extend compare", overlay=true)
ext = extend.right
isBoth = ext == extend.both ? 1.0 : 0.0
plot(isBoth)
`,
		},
		// --- line namespace ---
		{
			name: "line_style_constants",
			pine: `//@version=5
indicator("line style constants", overlay=true)
s1 = line.style_solid
s2 = line.style_dashed
s3 = line.style_dotted
s4 = line.style_arrow_left
s5 = line.style_arrow_right
s6 = line.style_arrow_both
s7 = line.style_cross
plot(s1 + s2 + s3 + s4 + s5 + s6 + s7)
`,
		},
		{
			name: "line_style_switch_selection",
			pine: `//@version=5
indicator("line style switch", overlay=true)
style = input.string("Solid", options=["Solid","Dashed","Dotted"])
ls = switch style
    "Solid"  => line.style_solid
    "Dashed" => line.style_dashed
    "Dotted" => line.style_dotted
plot(ls)
`,
		},
		// --- label namespace ---
		{
			name: "label_style_constants",
			pine: `//@version=5
indicator("label style constants", overlay=true)
s1  = label.style_none
s2  = label.style_xcross
s3  = label.style_cross
s4  = label.style_triangleup
s5  = label.style_triangledown
s6  = label.style_flag
s7  = label.style_circle
s8  = label.style_arrowup
s9  = label.style_arrowdown
s10 = label.style_label_up
plot(s1 + s2 + s3 + s4 + s5 + s6 + s7 + s8 + s9 + s10)
`,
		},
		{
			name: "label_style_lower_bounds",
			pine: `//@version=5
indicator("label style lower bounds", overlay=true)
s11 = label.style_label_down
s12 = label.style_label_left
s13 = label.style_label_right
s14 = label.style_label_lower_left
s15 = label.style_label_lower_right
s16 = label.style_label_upper_left
s17 = label.style_label_upper_right
s18 = label.style_label_center
s19 = label.style_square
s20 = label.style_diamond
plot(s11 + s12 + s13 + s14 + s15 + s16 + s17 + s18 + s19 + s20)
`,
		},
		// --- xloc namespace ---
		{
			name: "xloc_constants",
			pine: `//@version=5
indicator("xloc constants", overlay=true)
x1 = xloc.bar_time
x2 = xloc.bar_index
plot(x1 + x2)
`,
		},
		// --- size namespace ---
		{
			name: "size_constants",
			pine: `//@version=5
indicator("size constants", overlay=true)
sz1 = size.auto
sz2 = size.tiny
sz3 = size.small
sz4 = size.normal
sz5 = size.large
sz6 = size.huge
plot(sz1 + sz2 + sz3 + sz4 + sz5 + sz6)
`,
		},
		{
			name: "size_switch_selection",
			pine: `//@version=5
indicator("size switch", overlay=true)
sizeInput = input.string("Small", options=["Auto","Tiny","Small","Normal","Large","Huge"])
sz = switch sizeInput
    "Auto"   => size.auto
    "Tiny"   => size.tiny
    "Small"  => size.small
    "Normal" => size.normal
    "Large"  => size.large
    "Huge"   => size.huge
plot(sz)
`,
		},
		// --- shape namespace ---
		{
			name: "shape_constants",
			pine: `//@version=5
indicator("shape constants", overlay=true)
sh1  = shape.xcross
sh2  = shape.cross
sh3  = shape.circle
sh4  = shape.triangleup
sh5  = shape.triangledown
sh6  = shape.flag
sh7  = shape.labelup
sh8  = shape.labeldown
sh9  = shape.arrowup
sh10 = shape.arrowdown
sh11 = shape.diamond
sh12 = shape.square
plot(sh1 + sh2 + sh3 + sh4 + sh5 + sh6 + sh7 + sh8 + sh9 + sh10 + sh11 + sh12)
`,
		},
		// --- location namespace ---
		{
			name: "location_constants",
			pine: `//@version=5
indicator("location constants", overlay=true)
l1 = location.abovebar
l2 = location.belowbar
l3 = location.top
l4 = location.bottom
l5 = location.absolute
plot(l1 + l2 + l3 + l4 + l5)
`,
		},
		// --- hline namespace ---
		{
			name: "hline_style_constants",
			pine: `//@version=5
indicator("hline style constants", overlay=true)
h1 = hline.style_solid
h2 = hline.style_dashed
h3 = hline.style_dotted
plot(h1 + h2 + h3)
`,
		},
		// --- cross-namespace: all drawing namespaces in one script ---
		{
			name: "all_drawing_namespaces_combined",
			pine: `//@version=5
indicator("all drawing namespaces", overlay=true)
ext   = extend.right
ls    = line.style_solid
lbl   = label.style_arrowup
xl    = xloc.bar_index
sz    = size.normal
sh    = shape.triangleup
loc   = location.abovebar
hl    = hline.style_dashed
plot(ext + ls + lbl + xl + sz + sh + loc + hl)
`,
		},
		// --- constants used in conditional ternary ---
		{
			name: "drawing_constants_in_ternary",
			pine: `//@version=5
indicator("drawing constants ternary", overlay=true)
bullish = close > open
ext = bullish ? extend.right : extend.left
sz  = bullish ? size.large   : size.small
plot(ext + sz)
`,
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

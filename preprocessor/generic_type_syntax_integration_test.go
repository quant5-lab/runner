package preprocessor

import (
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestGenericTypeSyntaxRewriter_E2EParsing(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldParse bool
	}{
		{
			name: "array_new_generic_parses_after_rewrite",
			source: `//@version=5
indicator("Generic Array", overlay=true)
arr = array.new<float>(10, 0.0)
plot(array.size(arr))`,
			shouldParse: true,
		},
		{
			name: "matrix_new_generic_parses_after_rewrite",
			source: `//@version=5
indicator("Generic Matrix", overlay=true)
m = matrix.new<float>(3, 3)
plot(matrix.rows(m))`,
			shouldParse: true,
		},
		{
			name: "map_new_generic_parses_after_rewrite",
			source: `//@version=5
indicator("Generic Map", overlay=true)
lookup = map.new<string, float>()
map.put(lookup, "key", 1.0)
plot(map.size(lookup))`,
			shouldParse: true,
		},
		{
			name: "mixed_generic_syntax",
			source: `//@version=5
indicator("Mixed Generics", overlay=true)
arr = array.new<float>(5)
m = matrix.new<int>(2, 2)
lookup = map.new<string, bool>()
plot(array.size(arr) + matrix.rows(m) + map.size(lookup))`,
			shouldParse: true,
		},
		{
			name: "generic_syntax_with_comments_preserved",
			source: `//@version=5
indicator("Comments", overlay=true)
// array.new<float> is the new syntax
arr = array.new<float>(10)  // Create array
// Another comment with array.new<int> reference
plot(array.size(arr))`,
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewritten := RewriteGenericTypeSyntax(tt.source)

			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			_, err = p.ParseString("test.pine", rewritten)
			if tt.shouldParse && err != nil {
				t.Errorf("Expected source to parse after rewrite, but got error: %v\nRewritten source:\n%s", err, rewritten)
			}
			if !tt.shouldParse && err == nil {
				t.Error("Expected source to fail parsing, but it parsed successfully")
			}
		})
	}
}

func TestGenericTypeSyntaxRewriter_PreservesNonGenericCode(t *testing.T) {
	source := `//@version=5
indicator("Legacy Syntax", overlay=true)
arr = array.new_float(10, 0.0)
m = matrix.new_float(3, 3, 0.0)
plot(array.size(arr))`

	rewritten := RewriteGenericTypeSyntax(source)

	if rewritten != source {
		t.Errorf("Legacy syntax should not be modified.\nOriginal:\n%s\nRewritten:\n%s", source, rewritten)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, err = p.ParseString("test.pine", rewritten)
	if err != nil {
		t.Errorf("Legacy syntax should still parse: %v", err)
	}
}

func TestGenericTypeSyntaxRewriter_MixedGenericAndLegacy(t *testing.T) {
	source := `//@version=5
indicator("Mixed", overlay=true)
legacy = array.new_float(5)
generic = array.new<int>(10)
plot(array.size(legacy) + array.size(generic))`

	expected := `//@version=5
indicator("Mixed", overlay=true)
legacy = array.new_float(5)
generic = array.new_int(10)
plot(array.size(legacy) + array.size(generic))`

	rewritten := RewriteGenericTypeSyntax(source)
	if rewritten != expected {
		t.Errorf("Mixed syntax rewrite failed.\nExpected:\n%s\nGot:\n%s", expected, rewritten)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, err = p.ParseString("test.pine", rewritten)
	if err != nil {
		t.Errorf("Mixed syntax should parse: %v", err)
	}
}

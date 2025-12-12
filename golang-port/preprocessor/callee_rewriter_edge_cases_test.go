package preprocessor

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* CalleeRewriter edge case tests: boundary conditions, malformed input, invalid states */

func TestCalleeRewriter_MultipleDots(t *testing.T) {
	rewriter := NewCalleeRewriter()

	tests := []struct {
		name          string
		qualifiedName string
		expectRewrite bool
		expectObject  string
		expectProp    string
	}{
		{
			name:          "three dots - takes first two parts",
			qualifiedName: "a.b.c",
			expectRewrite: true,
			expectObject:  "a",
			expectProp:    "b.c",
		},
		{
			name:          "four dots - takes first two parts",
			qualifiedName: "request.security.data.close",
			expectRewrite: true,
			expectObject:  "request",
			expectProp:    "security.data.close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{
				Ident: &funcName,
			}

			rewritten := rewriter.Rewrite(callee, tt.qualifiedName)

			if rewritten != tt.expectRewrite {
				t.Errorf("Expected rewrite=%v, got %v", tt.expectRewrite, rewritten)
			}

			if tt.expectRewrite {
				if callee.MemberAccess == nil {
					t.Fatal("Expected MemberAccess to be created")
				}

				if callee.MemberAccess.Object != tt.expectObject {
					t.Errorf("Expected Object=%q, got %q", tt.expectObject, callee.MemberAccess.Object)
				}

				if callee.MemberAccess.Property != tt.expectProp {
					t.Errorf("Expected Property=%q, got %q", tt.expectProp, callee.MemberAccess.Property)
				}
			}
		})
	}
}

func TestCalleeRewriter_EmptyParts(t *testing.T) {
	rewriter := NewCalleeRewriter()

	tests := []struct {
		name          string
		qualifiedName string
		expectRewrite bool
		expectObject  string
		expectProp    string
	}{
		{
			name:          "leading dot - empty object",
			qualifiedName: ".max",
			expectRewrite: true,
			expectObject:  "",
			expectProp:    "max",
		},
		{
			name:          "trailing dot - empty property",
			qualifiedName: "math.",
			expectRewrite: true,
			expectObject:  "math",
			expectProp:    "",
		},
		{
			name:          "single dot - both empty",
			qualifiedName: ".",
			expectRewrite: true,
			expectObject:  "",
			expectProp:    "",
		},
		{
			name:          "empty string - no dot",
			qualifiedName: "",
			expectRewrite: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{
				Ident: &funcName,
			}

			rewritten := rewriter.Rewrite(callee, tt.qualifiedName)

			if rewritten != tt.expectRewrite {
				t.Errorf("Expected rewrite=%v, got %v", tt.expectRewrite, rewritten)
			}

			if tt.expectRewrite {
				if callee.MemberAccess == nil {
					t.Fatal("Expected MemberAccess to be created")
				}

				if callee.MemberAccess.Object != tt.expectObject {
					t.Errorf("Expected Object=%q, got %q", tt.expectObject, callee.MemberAccess.Object)
				}

				if callee.MemberAccess.Property != tt.expectProp {
					t.Errorf("Expected Property=%q, got %q", tt.expectProp, callee.MemberAccess.Property)
				}
			}
		})
	}
}

func TestCalleeRewriter_WhitespaceHandling(t *testing.T) {
	rewriter := NewCalleeRewriter()

	tests := []struct {
		name          string
		qualifiedName string
		expectRewrite bool
		expectObject  string
		expectProp    string
	}{
		{
			name:          "leading whitespace",
			qualifiedName: " math.max",
			expectRewrite: true,
			expectObject:  " math",
			expectProp:    "max",
		},
		{
			name:          "trailing whitespace",
			qualifiedName: "math.max ",
			expectRewrite: true,
			expectObject:  "math",
			expectProp:    "max ",
		},
		{
			name:          "whitespace around dot",
			qualifiedName: "math . max",
			expectRewrite: true,
			expectObject:  "math ",
			expectProp:    " max",
		},
		{
			name:          "tabs and newlines",
			qualifiedName: "ta.\tsma",
			expectRewrite: true,
			expectObject:  "ta",
			expectProp:    "\tsma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{
				Ident: &funcName,
			}

			rewritten := rewriter.Rewrite(callee, tt.qualifiedName)

			if rewritten != tt.expectRewrite {
				t.Errorf("Expected rewrite=%v, got %v", tt.expectRewrite, rewritten)
			}

			if tt.expectRewrite {
				if callee.MemberAccess == nil {
					t.Fatal("Expected MemberAccess to be created")
				}

				if callee.MemberAccess.Object != tt.expectObject {
					t.Errorf("Expected Object=%q, got %q", tt.expectObject, callee.MemberAccess.Object)
				}

				if callee.MemberAccess.Property != tt.expectProp {
					t.Errorf("Expected Property=%q, got %q", tt.expectProp, callee.MemberAccess.Property)
				}
			}
		})
	}
}

func TestCalleeRewriter_AlreadyMemberAccess(t *testing.T) {
	rewriter := NewCalleeRewriter()

	/* Callee already has MemberAccess (no Ident) - should not rewrite */
	callee := &parser.CallCallee{
		Ident: nil,
		MemberAccess: &parser.MemberAccess{
			Object:   "ta",
			Property: "sma",
		},
	}

	rewritten := rewriter.Rewrite(callee, "request.security")

	if rewritten {
		t.Error("Should not rewrite when Ident is nil (already MemberAccess)")
	}

	/* Original MemberAccess should remain unchanged */
	if callee.MemberAccess.Object != "ta" || callee.MemberAccess.Property != "sma" {
		t.Error("Original MemberAccess should not be modified")
	}
}

func TestCalleeRewriter_NilCallee(t *testing.T) {
	rewriter := NewCalleeRewriter()

	/* Nil callee should not panic */
	rewritten := rewriter.Rewrite(nil, "math.max")

	if rewritten {
		t.Error("Should return false for nil callee")
	}
}

func TestCalleeRewriter_NilIdent(t *testing.T) {
	rewriter := NewCalleeRewriter()

	/* Callee with nil Ident should not panic */
	callee := &parser.CallCallee{
		Ident:        nil,
		MemberAccess: nil,
	}

	rewritten := rewriter.Rewrite(callee, "math.max")

	if rewritten {
		t.Error("Should return false for nil Ident")
	}
}

func TestCalleeRewriter_LongQualifiedNames(t *testing.T) {
	rewriter := NewCalleeRewriter()

	/* Very long namespace/property names (stress test) */
	longObject := strings.Repeat("namespace", 100)
	longProperty := strings.Repeat("property", 100)
	qualifiedName := longObject + "." + longProperty

	funcName := "test"
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	rewritten := rewriter.Rewrite(callee, qualifiedName)

	if !rewritten {
		t.Error("Should handle long qualified names")
	}

	if callee.MemberAccess.Object != longObject {
		t.Error("Object should match long namespace")
	}

	if callee.MemberAccess.Property != longProperty {
		t.Error("Property should match long property")
	}
}

func TestCalleeRewriter_RewriteIfMapped_NilMappings(t *testing.T) {
	rewriter := NewCalleeRewriter()

	funcName := "max"
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	/* Nil mappings should return false, not panic */
	rewritten := rewriter.RewriteIfMapped(callee, "max", nil)

	if rewritten {
		t.Error("Should return false for nil mappings")
	}

	/* Ident should remain unchanged */
	if callee.Ident == nil || *callee.Ident != "max" {
		t.Error("Ident should not be modified when mappings are nil")
	}
}

func TestCalleeRewriter_RewriteIfMapped_EmptyMappings(t *testing.T) {
	rewriter := NewCalleeRewriter()

	funcName := "max"
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	/* Empty mappings should return false */
	emptyMappings := map[string]string{}
	rewritten := rewriter.RewriteIfMapped(callee, "max", emptyMappings)

	if rewritten {
		t.Error("Should return false when function not in mappings")
	}

	if callee.Ident == nil || *callee.Ident != "max" {
		t.Error("Ident should not be modified when function not mapped")
	}
}

func TestCalleeRewriter_RewriteIfMapped_EmptyStringKey(t *testing.T) {
	rewriter := NewCalleeRewriter()

	funcName := ""
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	mappings := map[string]string{
		"": "math.max",
	}

	rewritten := rewriter.RewriteIfMapped(callee, "", mappings)

	if !rewritten {
		t.Error("Should handle empty string as valid mapping key")
	}

	if callee.MemberAccess == nil {
		t.Fatal("Expected MemberAccess to be created")
	}

	if callee.MemberAccess.Object != "math" || callee.MemberAccess.Property != "max" {
		t.Error("Empty string key should map correctly")
	}
}

func TestCalleeRewriter_RewriteIfMapped_InvalidQualifiedName(t *testing.T) {
	rewriter := NewCalleeRewriter()

	funcName := "max"
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	/* Mapping points to invalid qualified name (no dot) */
	mappings := map[string]string{
		"max": "invalidname",
	}

	rewritten := rewriter.RewriteIfMapped(callee, "max", mappings)

	if rewritten {
		t.Error("Should return false when mapped value has no dot")
	}

	/* Ident should remain unchanged when rewrite fails */
	if callee.Ident == nil || *callee.Ident != "max" {
		t.Error("Ident should not be modified when mapped value is invalid")
	}
}

func TestCalleeRewriter_IdempotencyCheck(t *testing.T) {
	rewriter := NewCalleeRewriter()

	funcName := "max"
	callee := &parser.CallCallee{
		Ident: &funcName,
	}

	/* First rewrite: max → math.max */
	rewritten1 := rewriter.Rewrite(callee, "math.max")
	if !rewritten1 {
		t.Fatal("First rewrite should succeed")
	}

	/* Second rewrite attempt: should fail (Ident is now nil) */
	rewritten2 := rewriter.Rewrite(callee, "request.security")
	if rewritten2 {
		t.Error("Second rewrite should fail (idempotent behavior)")
	}

	/* MemberAccess should still be math.max (unchanged) */
	if callee.MemberAccess.Object != "math" || callee.MemberAccess.Property != "max" {
		t.Error("MemberAccess should not change after failed second rewrite")
	}
}

func TestCalleeRewriter_SpecialCharactersInNames(t *testing.T) {
	rewriter := NewCalleeRewriter()

	tests := []struct {
		name          string
		qualifiedName string
		expectObject  string
		expectProp    string
	}{
		{
			name:          "underscores",
			qualifiedName: "my_namespace.my_function",
			expectObject:  "my_namespace",
			expectProp:    "my_function",
		},
		{
			name:          "numbers",
			qualifiedName: "ta2.sma20",
			expectObject:  "ta2",
			expectProp:    "sma20",
		},
		{
			name:          "unicode characters",
			qualifiedName: "математика.макс",
			expectObject:  "математика",
			expectProp:    "макс",
		},
		{
			name:          "special symbols",
			qualifiedName: "ns$.func!",
			expectObject:  "ns$",
			expectProp:    "func!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{
				Ident: &funcName,
			}

			rewritten := rewriter.Rewrite(callee, tt.qualifiedName)

			if !rewritten {
				t.Errorf("Should handle special characters in qualified name")
			}

			if callee.MemberAccess == nil {
				t.Fatal("Expected MemberAccess to be created")
			}

			if callee.MemberAccess.Object != tt.expectObject {
				t.Errorf("Expected Object=%q, got %q", tt.expectObject, callee.MemberAccess.Object)
			}

			if callee.MemberAccess.Property != tt.expectProp {
				t.Errorf("Expected Property=%q, got %q", tt.expectProp, callee.MemberAccess.Property)
			}
		})
	}
}

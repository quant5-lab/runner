package preprocessor

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func assertMemberAccess(t *testing.T, callee *parser.CallCallee, expectObject string, expectProps []string) {
	t.Helper()
	if callee.MemberAccess == nil {
		t.Fatal("MemberAccess not created")
	}
	if callee.MemberAccess.Object != expectObject {
		t.Errorf("Object: expected=%q got=%q", expectObject, callee.MemberAccess.Object)
	}
	if len(callee.MemberAccess.Properties) != len(expectProps) {
		t.Errorf("Properties count: expected=%d got=%d", len(expectProps), len(callee.MemberAccess.Properties))
	}
	for i, expectProp := range expectProps {
		if i >= len(callee.MemberAccess.Properties) {
			break
		}
		if callee.MemberAccess.Properties[i] != expectProp {
			t.Errorf("Properties[%d]: expected=%q got=%q", i, expectProp, callee.MemberAccess.Properties[i])
		}
	}
}

func TestCalleeRewriter_MultipleDots(t *testing.T) {
	rewriter := NewCalleeRewriter()
	tests := []struct {
		name          string
		qualifiedName string
		expectObject  string
		expectProps   []string
	}{
		{"three-level", "a.b.c", "a", []string{"b", "c"}},
		{"four-level", "request.security.data.close", "request", []string{"security", "data", "close"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{Ident: &funcName}
			if !rewriter.Rewrite(callee, tt.qualifiedName) {
				t.Fatal("rewrite failed")
			}
			assertMemberAccess(t, callee, tt.expectObject, tt.expectProps)
		})
	}
}

func TestCalleeRewriter_NilCallee(t *testing.T) {
	rewriter := NewCalleeRewriter()
	if rewriter.Rewrite(nil, "math.max") {
		t.Error("should return false for nil callee")
	}
}

func TestCalleeRewriter_LongNames(t *testing.T) {
	rewriter := NewCalleeRewriter()
	longObj := strings.Repeat("ns", 50)
	longProp := strings.Repeat("prop", 50)
	funcName := "test"
	callee := &parser.CallCallee{Ident: &funcName}
	if !rewriter.Rewrite(callee, longObj+"."+longProp) {
		t.Error("should handle long names")
	}
	assertMemberAccess(t, callee, longObj, []string{longProp})
}

func TestCalleeRewriter_ErrorCases(t *testing.T) {
	rewriter := NewCalleeRewriter()
	tests := []struct {
		name          string
		qualifiedName string
		shouldRewrite bool
		expectObject  string
		expectProps   []string
		note          string
	}{
		{"empty string", "", false, "", nil, "no dots"},
		{"single dot", ".", true, "", []string{""}, "current: creates empty object and property"},
		{"leading dot", ".math.max", true, "", []string{"math", "max"}, "current: creates empty object"},
		{"trailing dot", "math.max.", true, "math", []string{"max", ""}, "current: creates empty property"},
		{"multiple consecutive dots", "a..b", true, "a", []string{"", "b"}, "current: creates empty property"},
		{"single identifier no dot", "simple", false, "", nil, "no dots"},
		{"whitespace only", "   ", false, "", nil, "no dots"},
		{"dot with spaces", "a . b", true, "a ", []string{" b"}, "current: preserves whitespace in identifiers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{Ident: &funcName}
			result := rewriter.Rewrite(callee, tt.qualifiedName)
			if result != tt.shouldRewrite {
				t.Errorf("Rewrite(%q) = %v, want %v (%s)", tt.qualifiedName, result, tt.shouldRewrite, tt.note)
			}
			if tt.shouldRewrite && callee.MemberAccess != nil {
				if callee.MemberAccess.Object != tt.expectObject {
					t.Logf("Note: %s", tt.note)
					t.Errorf("Object: expected=%q got=%q", tt.expectObject, callee.MemberAccess.Object)
				}
				if len(callee.MemberAccess.Properties) != len(tt.expectProps) {
					t.Logf("Note: %s", tt.note)
					t.Errorf("Properties count: expected=%d got=%d", len(tt.expectProps), len(callee.MemberAccess.Properties))
				}
			}
		})
	}
}

func TestCalleeRewriter_SingleIdentifier(t *testing.T) {
	rewriter := NewCalleeRewriter()
	funcName := "test"
	callee := &parser.CallCallee{Ident: &funcName}
	if rewriter.Rewrite(callee, "simple") {
		t.Error("should return false for single identifier without dots")
	}
	if callee.MemberAccess != nil {
		t.Error("MemberAccess should remain nil for single identifier")
	}
}

func TestCalleeRewriter_WhitespaceHandling(t *testing.T) {
	rewriter := NewCalleeRewriter()
	tests := []struct {
		name            string
		qualifiedName   string
		shouldRewrite   bool
		expectObject    string
		expectFirstProp string
		behaviorNote    string
	}{
		{"spaces before dot", "a .b", true, "a ", "b", "whitespace preserved in object, stripped from property by Split"},
		{"spaces after dot", "a. b", true, "a", " b", "whitespace preserved in property"},
		{"spaces around dot", "a . b", true, "a ", " b", "whitespace preserved in both"},
		{"tabs", "a\t.\tb", true, "a\t", "\tb", "whitespace preserved in both"},
		{"newlines", "a\n.\nb", true, "a\n", "\nb", "whitespace preserved in both"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := "test"
			callee := &parser.CallCallee{Ident: &funcName}
			result := rewriter.Rewrite(callee, tt.qualifiedName)
			if result != tt.shouldRewrite {
				t.Errorf("Rewrite(%q) = %v, want %v", tt.qualifiedName, result, tt.shouldRewrite)
			}
			if result && callee.MemberAccess != nil {
				if callee.MemberAccess.Object != tt.expectObject {
					t.Logf("Note: %s", tt.behaviorNote)
					t.Errorf("Object: expected=%q got=%q", tt.expectObject, callee.MemberAccess.Object)
				}
				if len(callee.MemberAccess.Properties) > 0 && callee.MemberAccess.Properties[0] != tt.expectFirstProp {
					t.Logf("Note: %s", tt.behaviorNote)
					t.Errorf("First property: expected=%q got=%q", tt.expectFirstProp, callee.MemberAccess.Properties[0])
				}
			}
		})
	}
}

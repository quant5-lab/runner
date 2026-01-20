package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestBuiltinMemberExpression_TrSupport validates ta.tr MemberExpression support */
func TestBuiltinMemberExpression_TrSupport(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	t.Run("ta.tr MemberExpression", func(t *testing.T) {
		taTrExpr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "tr"},
			Computed: false,
		}

		code, resolved := handler.TryResolveMemberExpression(taTrExpr, false)
		if !resolved {
			t.Fatal("ta.tr should be resolved by TryResolveMemberExpression")
		}

		if !strings.Contains(code, "math.Max") {
			t.Errorf("ta.tr should generate true range calculation, got: %s", code)
		}
	})

	t.Run("bare tr identifier", func(t *testing.T) {
		trIdentifier := &ast.Identifier{Name: "tr"}

		code, resolved := handler.TryResolveIdentifier(trIdentifier, false)
		if !resolved {
			t.Fatal("tr should be resolved by TryResolveIdentifier")
		}

		if !strings.Contains(code, "math.Max") {
			t.Errorf("tr should generate true range calculation, got: %s", code)
		}
	})
}

/* TestBuiltinIdentifier_AllBuiltinsSupported validates all builtin identifiers (non-MemberExpression) */
func TestBuiltinIdentifier_AllBuiltinsSupported(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	builtins := []struct {
		name           string
		identifier     string
		expectedAccess string
	}{
		{"close", "close", ".Close"},
		{"open", "open", ".Open"},
		{"high", "high", ".High"},
		{"low", "low", ".Low"},
		{"volume", "volume", ".Volume"},
		{"tr", "tr", "math.Max"},
	}

	for _, builtin := range builtins {
		t.Run(builtin.name, func(t *testing.T) {
			identifier := &ast.Identifier{Name: builtin.identifier}

			code, resolved := handler.TryResolveIdentifier(identifier, false)
			if !resolved {
				t.Fatalf("%s should be resolved by TryResolveIdentifier", builtin.name)
			}

			if !strings.Contains(code, builtin.expectedAccess) {
				t.Errorf("%s should contain %q, got: %s", builtin.name, builtin.expectedAccess, code)
			}
		})
	}
}

/* TestBuiltinMemberExpression_OtherBuiltinsNotSupported documents current limitation */
func TestBuiltinMemberExpression_OtherBuiltinsNotSupported(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	unsupportedBuiltins := []string{"close", "open", "high", "low", "volume"}

	for _, builtin := range unsupportedBuiltins {
		t.Run("ta."+builtin+" not yet supported", func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: builtin},
				Computed: false,
			}

			_, resolved := handler.TryResolveMemberExpression(expr, false)
			if resolved {
				t.Errorf("ta.%s is unexpectedly supported (test needs update if implemented)", builtin)
			}
		})
	}
}

/* TestBuiltinMemberExpression_TrNestedSubscript validates ta.tr[offset] nested MemberExpression */
func TestBuiltinMemberExpression_TrNestedSubscript(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name           string
		offset         int
		expectedOffset string
	}{
		{"ta.tr[0]", 0, "i-0"},
		{"ta.tr[1]", 1, "i-1"},
		{"ta.tr[2]", 2, "i-2"},
		{"ta.tr[5]", 5, "i-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nested MemberExpression: ta.tr[offset]
			nestedExpr := &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "tr"},
					Computed: false,
				},
				Property: &ast.Literal{Value: tt.offset},
				Computed: true,
			}

			code, resolved := handler.TryResolveMemberExpression(nestedExpr, false)
			if !resolved {
				t.Fatalf("%s should be resolved", tt.name)
			}

			if !strings.Contains(code, tt.expectedOffset) {
				t.Errorf("%s should contain offset %q, got: %s", tt.name, tt.expectedOffset, code)
			}

			// Verify tr calculation is present
			if !strings.Contains(code, "math.Max") {
				t.Errorf("%s should generate true range calculation, got: %s", tt.name, code)
			}
		})
	}
}

/* TestBuiltinMemberExpression_ContextConsistency validates tr works consistently across contexts */
func TestBuiltinMemberExpression_ContextConsistency(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	contexts := []struct {
		name    string
		getCode func() string
	}{
		{
			"current bar",
			func() string { return handler.GenerateCurrentBarAccess("tr") },
		},
		{
			"security context",
			func() string { return handler.GenerateSecurityContextAccess("tr") },
		},
		{
			"historical offset 1",
			func() string { return handler.GenerateHistoricalAccess("tr", 1) },
		},
		{
			"historical offset 2",
			func() string { return handler.GenerateHistoricalAccess("tr", 2) },
		},
	}

	for _, ctx := range contexts {
		t.Run(ctx.name, func(t *testing.T) {
			code := ctx.getCode()

			if code == "" {
				t.Fatalf("tr in %s returned empty code", ctx.name)
			}

			// All contexts should generate inline tr calculation
			if !strings.Contains(code, "math.Max") {
				t.Errorf("tr in %s should generate calculation, got: %s", ctx.name, code)
			}

			// Verify no Series.Get() pattern
			if strings.Contains(code, "trSeries.Get(") || strings.Contains(code, ".Get(tr") {
				t.Errorf("tr in %s should not use Series.Get(), got: %s", ctx.name, code)
			}
		})
	}
}

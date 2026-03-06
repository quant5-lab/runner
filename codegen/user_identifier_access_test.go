package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
Validates context-aware identifier resolution for user-defined variables.

ForwardSeriesBuffer paradigm: top-level scope accesses variables via %sSeries.GetCurrent(),
arrow function scope resolves parameters and locals as scalars, loop-modified variables
as Series.GetCurrent(). Tests validate resolution behavior across all identifier categories
and generation contexts, not specific implementation details.
*/

/* TestResolveUserIdentifierAccess validates canonical identifier resolution across all categories */
func TestResolveUserIdentifierAccess(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(g *generator)
		ident    string
		expected string
	}{
		{
			name:     "top-level scope defaults to series access",
			setup:    func(g *generator) {},
			ident:    "myVar",
			expected: "myVarSeries.GetCurrent()",
		},
		{
			name: "arrow parameter resolves to scalar",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				g.arrowAccessResolver = r
			},
			ident:    "src",
			expected: "src",
		},
		{
			name: "arrow local variable resolves to scalar",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLocalVariable("x")
				g.arrowAccessResolver = r
			},
			ident:    "x",
			expected: "x",
		},
		{
			name: "arrow loop-modified variable resolves to series access",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLoopModified("acc")
				g.arrowAccessResolver = r
			},
			ident:    "acc",
			expected: "accSeries.GetCurrent()",
		},
		{
			name: "arrow unknown identifier falls through to series access",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				g.arrowAccessResolver = r
			},
			ident:    "outerScopeVar",
			expected: "outerScopeVarSeries.GetCurrent()",
		},
		{
			name: "arrow empty resolver falls through to series access",
			setup: func(g *generator) {
				g.arrowAccessResolver = NewArrowSeriesAccessResolver()
			},
			ident:    "anyVar",
			expected: "anyVarSeries.GetCurrent()",
		},
		{
			name: "arrow series parameter resolves to nameSeries.GetCurrent()",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterSeriesParameter("src")
				g.arrowAccessResolver = r
			},
			ident:    "src",
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "arrow series parameter takes priority over scalar parameter",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterSeriesParameter("src")
				g.arrowAccessResolver = r
			},
			ident:    "src",
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "arrow parameter takes priority over local with same name",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("val")
				r.RegisterLocalVariable("val")
				g.arrowAccessResolver = r
			},
			ident:    "val",
			expected: "val",
		},
		{
			name: "arrow context with multiple categories resolves each correctly",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterLocalVariable("diff")
				r.RegisterLoopModified("acc")
				g.arrowAccessResolver = r
			},
			ident:    "diff",
			expected: "diff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			tt.setup(g)

			result := g.resolveUserIdentifierAccess(tt.ident)
			if result != tt.expected {
				t.Errorf("resolveUserIdentifierAccess(%q) = %q, want %q", tt.ident, result, tt.expected)
			}
		})
	}
}

/*
TestExtractSeriesExpression_ContextAwareIdentifierResolution validates identifier extraction

	respects active generation context for all expression types
*/
func TestExtractSeriesExpression_ContextAwareIdentifierResolution(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(g *generator)
		expr     ast.Expression
		expected string
	}{
		{
			name:     "identifier without arrow context uses series access",
			setup:    func(g *generator) {},
			expr:     &ast.Identifier{Name: "src"},
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "parameter identifier in arrow context resolves to scalar",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "src"},
			expected: "src",
		},
		{
			name: "local variable identifier in arrow context resolves to scalar",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLocalVariable("diff")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "diff"},
			expected: "diff",
		},
		{
			name: "series parameter identifier in arrow context resolves to GetCurrent()",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterSeriesParameter("src")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "src"},
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "outer scope identifier in arrow context uses series access",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "sma20"},
			expected: "sma20Series.GetCurrent()",
		},
		{
			name: "loop-modified identifier in arrow context uses series access",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLoopModified("running_sum")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "running_sum"},
			expected: "running_sumSeries.GetCurrent()",
		},
		{
			name:     "literal expression unchanged by context",
			setup:    func(g *generator) {},
			expr:     &ast.Literal{Value: 42.0},
			expected: "42",
		},
		{
			name: "multiple parameters each resolve independently",
			setup: func(g *generator) {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterParameter("len")
				g.arrowAccessResolver = r
			},
			expr:     &ast.Identifier{Name: "len"},
			expected: "len",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			tt.setup(g)

			result := g.extractSeriesExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("extractSeriesExpression() = %q, want %q", result, tt.expected)
			}
		})
	}
}

/*
TestCallDelegation_IdentifierResolutionScope validates that identifier resolution context

	propagates correctly through call handler delegation and is properly scoped
*/
func TestCallDelegation_IdentifierResolutionScope(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *ArrowSeriesAccessResolver
		call     *ast.CallExpression
		expected string
	}{
		{
			name: "math.max with two parameters",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterParameter("threshold")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "max"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Identifier{Name: "threshold"},
				},
			},
			expected: "math.Max(src, threshold)",
		},
		{
			name: "math.max with parameter and literal",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "max"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Literal{Value: 0.0},
				},
			},
			expected: "math.Max(src, 0)",
		},
		{
			name: "math.min with parameter and outer scope series",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "min"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "src"},
					&ast.Identifier{Name: "sma20"},
				},
			},
			expected: "math.Min(src, sma20Series.GetCurrent())",
		},
		{
			name: "math.abs with local variable",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLocalVariable("diff")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "diff"},
				},
			},
			expected: "math.Abs(diff)",
		},
		{
			name: "math.abs with loop-modified variable",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterLoopModified("running_sum")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "running_sum"},
				},
			},
			expected: "math.Abs(running_sumSeries.GetCurrent())",
		},
		{
			name: "math.abs with binary expression of parameters",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterParameter("offset")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "src"},
						Operator: "-",
						Right:    &ast.Identifier{Name: "offset"},
					},
				},
			},
			expected: "math.Abs((src - offset))",
		},
		{
			name: "math.pow with all outer scope identifiers",
			setup: func() *ArrowSeriesAccessResolver {
				return NewArrowSeriesAccessResolver()
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "pow"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "baseVal"},
					&ast.Identifier{Name: "exponent"},
				},
			},
			expected: "math.Pow(baseValSeries.GetCurrent(), exponentSeries.GetCurrent())",
		},
		{
			name: "math.max with mixed categories: parameter, local, loop-modified",
			setup: func() *ArrowSeriesAccessResolver {
				r := NewArrowSeriesAccessResolver()
				r.RegisterParameter("src")
				r.RegisterLocalVariable("scaled")
				return r
			},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "max"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "scaled"},
					&ast.Identifier{Name: "src"},
				},
			},
			expected: "math.Max(scaled, src)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.callRouter = NewCallExpressionRouter()
			resolver := tt.setup()
			exprGen := NewArrowExpressionGeneratorImpl(g, resolver)

			code, err := exprGen.Generate(tt.call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			if code != tt.expected {
				t.Errorf("Generate() = %q, want %q", code, tt.expected)
			}
		})
	}
}

/*
TestCallDelegation_ResolverLifecycle validates that arrow context is scoped

	to the delegated call and does not leak into subsequent operations
*/
func TestCallDelegation_ResolverLifecycle(t *testing.T) {
	g := newTestGenerator()
	g.callRouter = NewCallExpressionRouter()

	resolver := NewArrowSeriesAccessResolver()
	resolver.RegisterParameter("src")
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "math"},
			Property: &ast.Identifier{Name: "abs"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "src"},
		},
	}

	_, err := exprGen.Generate(call)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if g.arrowAccessResolver != nil {
		t.Error("arrow context leaked after delegated call completed")
	}

	/* After call returns, same identifier must revert to top-level series access */
	afterResult := g.extractSeriesExpression(&ast.Identifier{Name: "src"})
	expected := "srcSeries.GetCurrent()"
	if afterResult != expected {
		t.Errorf("post-delegation extractSeriesExpression(%q) = %q, want %q", "src", afterResult, expected)
	}
}

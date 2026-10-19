package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* --------------------------------------------------------------------------
 * TestOuterScopeCapture_GoParamName
 *
 * GoParamName returns the Go identifier used both in function signatures and
 * at call sites for non-bool scalar captures. The naming convention encodes
 * kind so generated code remains unambiguous when multiple captures coexist.
 * -------------------------------------------------------------------------*/

func TestOuterScopeCapture_GoParamName(t *testing.T) {
	tests := []struct {
		name     string
		capture  OuterScopeCapture
		wantName string
	}{
		// Series captures gain a "Series" suffix so generated code unambiguously
		// refers to the *series.Series pointer rather than a bare float64.
		{
			name:     "series float appends Series suffix",
			capture:  OuterScopeCapture{Name: "price", Kind: OuterScopeCaptureSeriesFloat},
			wantName: "priceSeries",
		},
		// ArraySeries captures gain "ArraySeries" to distinguish from plain series.
		{
			name:     "array series appends ArraySeries suffix",
			capture:  OuterScopeCapture{Name: "levels", Kind: OuterScopeCaptureArraySeries},
			wantName: "levelsArraySeries",
		},
		// StringArraySeries captures gain "StringArraySeries".
		{
			name:     "string array series appends StringArraySeries suffix",
			capture:  OuterScopeCapture{Name: "labels", Kind: OuterScopeCaptureStringArraySeries},
			wantName: "labelsStringArraySeries",
		},
		// Scalar and string captures use the bare Pine name — they are already
		// primitive Go types and require no disambiguation suffix.
		{
			name:     "scalar uses bare name",
			capture:  OuterScopeCapture{Name: "length", Kind: OuterScopeCaptureScalar},
			wantName: "length",
		},
		{
			name:     "string uses bare name",
			capture:  OuterScopeCapture{Name: "session", Kind: OuterScopeCaptureString},
			wantName: "session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.capture.GoParamName(); got != tt.wantName {
				t.Errorf("GoParamName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestOuterScopeCapture_GoParamType
 *
 * GoParamType returns the Go type string used in function signatures.
 * -------------------------------------------------------------------------*/

func TestOuterScopeCapture_GoParamType(t *testing.T) {
	tests := []struct {
		name     string
		capture  OuterScopeCapture
		wantType string
	}{
		{"series float → *series.Series", OuterScopeCapture{Kind: OuterScopeCaptureSeriesFloat}, "*series.Series"},
		{"array series → *series.ArraySeries", OuterScopeCapture{Kind: OuterScopeCaptureArraySeries}, "*series.ArraySeries"},
		{"string array series → *series.StringArraySeries", OuterScopeCapture{Kind: OuterScopeCaptureStringArraySeries}, "*series.StringArraySeries"},
		{"string → string", OuterScopeCapture{Kind: OuterScopeCaptureString}, "string"},
		{"scalar → float64", OuterScopeCapture{Kind: OuterScopeCaptureScalar}, "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.capture.GoParamType(); got != tt.wantType {
				t.Errorf("GoParamType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestOuterScopeCapture_GoCallSiteExpression
 *
 * GoCallSiteExpression returns the Go expression to emit at a call site.
 * For bool scalar constants, the bare name is a Go untyped bool and would
 * cause a compile-time type mismatch against a float64 parameter, so an
 * IIFE conversion is required. Every other capture kind returns GoParamName().
 * -------------------------------------------------------------------------*/

func TestOuterScopeCapture_GoCallSiteExpression_BoolScalarWrappedInIIFE(t *testing.T) {
	tests := []struct {
		name      string
		constName string
		constVal  bool
	}{
		{"true constant", "FilterOn", true},
		{"false constant", "FilterOff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cap := OuterScopeCapture{Name: tt.constName, Kind: OuterScopeCaptureScalar}
			constants := map[string]interface{}{tt.constName: tt.constVal}

			got := cap.GoCallSiteExpression(constants)

			if !strings.HasPrefix(got, "func() float64 {") {
				t.Errorf("expected IIFE prefix, got: %s", got)
			}
			if !strings.Contains(got, tt.constName) {
				t.Errorf("IIFE must reference constant name %q, got: %s", tt.constName, got)
			}
			if !strings.Contains(got, "return 1.0") || !strings.Contains(got, "return 0.0") {
				t.Errorf("IIFE must return 1.0/0.0, got: %s", got)
			}
		})
	}
}

func TestOuterScopeCapture_GoCallSiteExpression_NonBoolFallsBackToParamName(t *testing.T) {
	tests := []struct {
		name      string
		capture   OuterScopeCapture
		constants map[string]interface{}
	}{
		{
			name:      "float scalar constant uses bare name",
			capture:   OuterScopeCapture{Name: "length", Kind: OuterScopeCaptureScalar},
			constants: map[string]interface{}{"length": 14.0},
		},
		{
			name:      "string constant uses bare name",
			capture:   OuterScopeCapture{Name: "sess", Kind: OuterScopeCaptureString},
			constants: map[string]interface{}{"sess": "0900-1700"},
		},
		{
			name:      "scalar not in constants map uses bare name",
			capture:   OuterScopeCapture{Name: "period", Kind: OuterScopeCaptureScalar},
			constants: map[string]interface{}{},
		},
		{
			name:      "series float uses Series-suffixed name",
			capture:   OuterScopeCapture{Name: "price", Kind: OuterScopeCaptureSeriesFloat},
			constants: map[string]interface{}{},
		},
		{
			name:      "array series uses ArraySeries-suffixed name",
			capture:   OuterScopeCapture{Name: "levels", Kind: OuterScopeCaptureArraySeries},
			constants: map[string]interface{}{},
		},
		{
			name:      "nil constants map is safe for non-scalar",
			capture:   OuterScopeCapture{Name: "src", Kind: OuterScopeCaptureSeriesFloat},
			constants: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.capture.GoCallSiteExpression(tt.constants)
			want := tt.capture.GoParamName()
			if got != want {
				t.Errorf("GoCallSiteExpression() = %q, want %q (GoParamName)", got, want)
			}
		})
	}
}

// TestOuterScopeCapture_GoCallSiteExpression_NilConstantsMapSafeForScalar verifies
// that passing nil for constants does not panic for a scalar capture that has no entry.
func TestOuterScopeCapture_GoCallSiteExpression_NilConstantsMapSafeForScalar(t *testing.T) {
	cap := OuterScopeCapture{Name: "x", Kind: OuterScopeCaptureScalar}
	got := cap.GoCallSiteExpression(nil)
	if got != "x" {
		t.Errorf("got %q, want %q", got, "x")
	}
}

/* --------------------------------------------------------------------------
 * TestKindForConstant – internal classification of constant values
 * -------------------------------------------------------------------------*/

func TestKindForConstant(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		want OuterScopeCaptureKind
	}{
		{"float64 → Scalar", 14.0, OuterScopeCaptureScalar},
		{"int → Scalar", 42, OuterScopeCaptureScalar},
		{"bool true → Scalar", true, OuterScopeCaptureScalar},
		{"bool false → Scalar", false, OuterScopeCaptureScalar},
		{"arbitrary string → String", "0900-1700", OuterScopeCaptureString},
		{"empty string → String", "", OuterScopeCaptureString},
		{"input.source → SeriesFloat", "input.source", OuterScopeCaptureSeriesFloat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kindForConstant(tt.val); got != tt.want {
				t.Errorf("kindForConstant(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestKindForVariable – mapping from generator variable type tags to capture kinds
 * -------------------------------------------------------------------------*/

func TestKindForVariable(t *testing.T) {
	tests := []struct {
		varType  string
		wantKind OuterScopeCaptureKind
		wantOK   bool
	}{
		{"float", OuterScopeCaptureSeriesFloat, true},
		{"float64", OuterScopeCaptureSeriesFloat, true},
		{"bool", OuterScopeCaptureSeriesFloat, true},
		{"string", OuterScopeCaptureString, true},
		{"array_series_float", OuterScopeCaptureArraySeries, true},
		{"array_series_string", OuterScopeCaptureStringArraySeries, true},
		// Types resolved by other mechanisms must not be captured as parameters.
		{"function", 0, false},
		{"color", 0, false},
		{"", 0, false},
		{"unknown_type", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.varType, func(t *testing.T) {
			gotKind, gotOK := kindForVariable(tt.varType)
			if gotOK != tt.wantOK {
				t.Errorf("kindForVariable(%q) ok = %v, want %v", tt.varType, gotOK, tt.wantOK)
			}
			if tt.wantOK && gotKind != tt.wantKind {
				t.Errorf("kindForVariable(%q) kind = %v, want %v", tt.varType, gotKind, tt.wantKind)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestArrowCaptureRegistry – store and retrieve captures per function name
 * -------------------------------------------------------------------------*/

func TestArrowCaptureRegistry(t *testing.T) {
	t.Run("registered captures are retrievable", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		caps := []OuterScopeCapture{
			{Name: "price", Kind: OuterScopeCaptureSeriesFloat},
			{Name: "length", Kind: OuterScopeCaptureScalar},
		}
		r.Register("myFn", caps)

		got := r.Get("myFn")
		if len(got) != 2 {
			t.Fatalf("Get returned %d captures, want 2", len(got))
		}
		if got[0].Name != "price" || got[1].Name != "length" {
			t.Errorf("capture names mismatch: %v", got)
		}
	})

	t.Run("unknown function returns nil", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		if got := r.Get("nonExistent"); got != nil {
			t.Errorf("expected nil for unregistered function, got %v", got)
		}
	})

	t.Run("multiple functions stored independently", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		r.Register("f1", []OuterScopeCapture{{Name: "a", Kind: OuterScopeCaptureScalar}})
		r.Register("f2", []OuterScopeCapture{{Name: "b", Kind: OuterScopeCaptureString}})

		if r.Get("f1")[0].Name != "a" {
			t.Error("f1 captures corrupted by f2 registration")
		}
		if r.Get("f2")[0].Name != "b" {
			t.Error("f2 captures incorrect")
		}
	})

	t.Run("empty capture list registered and returned", func(t *testing.T) {
		r := NewArrowCaptureRegistry()
		r.Register("pure", []OuterScopeCapture{})
		got := r.Get("pure")
		if got == nil {
			t.Error("expected non-nil empty slice for registered function with no captures")
		}
		if len(got) != 0 {
			t.Errorf("expected 0 captures, got %d", len(got))
		}
	})
}

/* --------------------------------------------------------------------------
 * TestOuterScopeCaptureAnalyzer_Analyze
 *
 * Analyze must capture every identifier the arrow body reads from the outer
 * scope — across all expression and statement node types — without capturing
 * params, locals, non-computed property field names, or the same name twice.
 * -------------------------------------------------------------------------*/

func TestOuterScopeCaptureAnalyzer_Analyze(t *testing.T) {
	ident := func(name string) *ast.Identifier { return &ast.Identifier{Name: name} }
	lit := func(v float64) *ast.Literal { return &ast.Literal{Value: v} }
	exprStmt := func(e ast.Expression) ast.Node {
		return &ast.ExpressionStatement{Expression: e}
	}

	outerConsts := map[string]interface{}{"offset": 2.0, "showLabel": false}
	outerVars := map[string]string{"arr": "float", "price": "float", "foo": "float", "bar": "float"}

	tests := []struct {
		name         string
		params       map[string]bool
		locals       map[string]bool
		constants    map[string]interface{}
		variables    map[string]string
		body         []ast.Node
		wantCaptured []string
		wantAbsent   []string
	}{
		{
			name:         "BinaryExpression captures both operands",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.BinaryExpression{Operator: "+", Left: ident("offset"), Right: ident("price")})},
			wantCaptured: []string{"offset", "price"},
		},
		{
			name:         "UnaryExpression captures argument",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.UnaryExpression{Operator: "-", Argument: ident("price")})},
			wantCaptured: []string{"price"},
		},
		{
			name:         "LogicalExpression captures both sides",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.LogicalExpression{Operator: "and", Left: ident("showLabel"), Right: ident("price")})},
			wantCaptured: []string{"showLabel", "price"},
		},
		{
			name:      "ConditionalExpression captures test consequent and alternate",
			constants: outerConsts,
			variables: outerVars,
			body: []ast.Node{exprStmt(&ast.ConditionalExpression{
				Test: ident("showLabel"), Consequent: ident("price"), Alternate: lit(0),
			})},
			wantCaptured: []string{"showLabel", "price"},
		},
		{
			name:         "CallExpression captures all arguments",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.CallExpression{Callee: ident("ta.sma"), Arguments: []ast.Expression{ident("price"), ident("offset")}})},
			wantCaptured: []string{"price", "offset"},
		},
		{
			name:         "computed MemberExpression captures object and index",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.MemberExpression{Object: ident("arr"), Property: ident("offset"), Computed: true})},
			wantCaptured: []string{"arr", "offset"},
		},
		{
			name:         "non-computed MemberExpression captures object but not field name",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(&ast.MemberExpression{Object: ident("foo"), Property: ident("bar"), Computed: false})},
			wantCaptured: []string{"foo"},
			wantAbsent:   []string{"bar"},
		},
		{
			name:      "IfStatement captures test and all branch identifiers",
			constants: outerConsts,
			variables: outerVars,
			body: []ast.Node{&ast.IfStatement{
				Test:       ident("showLabel"),
				Consequent: []ast.Node{exprStmt(ident("price"))},
				Alternate:  []ast.Node{exprStmt(ident("arr"))},
			}},
			wantCaptured: []string{"showLabel", "price", "arr"},
		},
		{
			name:      "VariableDeclaration initializer is scanned",
			constants: outerConsts,
			variables: outerVars,
			body: []ast.Node{&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{ID: &ast.Identifier{Name: "x"}, Init: ident("price")},
				},
			}},
			wantCaptured: []string{"price"},
		},
		{
			name:      "nested expression captures deeply nested identifiers",
			constants: outerConsts,
			variables: outerVars,
			body: []ast.Node{exprStmt(&ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.BinaryExpression{Operator: "+", Left: ident("price"), Right: ident("offset")},
				Right:    ident("arr"),
			})},
			wantCaptured: []string{"price", "offset", "arr"},
		},
		{
			name:         "repeated references deduplicated to single capture",
			constants:    outerConsts,
			variables:    outerVars,
			body:         []ast.Node{exprStmt(ident("price")), exprStmt(ident("price")), exprStmt(ident("price"))},
			wantCaptured: []string{"price"},
		},
		{
			name:       "param shadows outer variable and is not captured",
			params:     map[string]bool{"price": true},
			constants:  outerConsts,
			variables:  outerVars,
			body:       []ast.Node{exprStmt(ident("price"))},
			wantAbsent: []string{"price"},
		},
		{
			name:       "local shadows outer variable and is not captured",
			locals:     map[string]bool{"price": true},
			constants:  outerConsts,
			variables:  outerVars,
			body:       []ast.Node{exprStmt(ident("price"))},
			wantAbsent: []string{"price"},
		},
		{
			name:       "unknown identifier absent from constants and variables is ignored",
			constants:  map[string]interface{}{},
			variables:  map[string]string{},
			body:       []ast.Node{exprStmt(ident("unknownVar"))},
			wantAbsent: []string{"unknownVar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			if params == nil {
				params = map[string]bool{}
			}
			locals := tt.locals
			if locals == nil {
				locals = map[string]bool{}
			}

			analyzer := NewOuterScopeCaptureAnalyzer(params, locals, tt.constants, tt.variables)
			captures := analyzer.Analyze(tt.body)

			capturedCount := make(map[string]int)
			for _, c := range captures {
				capturedCount[c.Name]++
			}

			for _, want := range tt.wantCaptured {
				if capturedCount[want] == 0 {
					t.Errorf("expected %q to be captured; captures: %v", want, captures)
				}
			}
			for _, absent := range tt.wantAbsent {
				if capturedCount[absent] > 0 {
					t.Errorf("expected %q NOT to be captured; captures: %v", absent, captures)
				}
			}
			for name, count := range capturedCount {
				if count > 1 {
					t.Errorf("identifier %q captured %d times; must be deduplicated", name, count)
				}
			}
		})
	}
}

package codegen

import "testing"

/*
TestArrowSeriesAccessResolver_Registration validates that all identifier categories register
correctly and are queryable via inspection methods. These tests verify the resolver's registration
API orthogonally from its resolution behavior.
*/
func TestArrowSeriesAccessResolver_Registration(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *ArrowSeriesAccessResolver)
		ident    string
		isParam  bool // IsParameter
		isSeries bool // IsSeriesParameter
		isLocal  bool // IsLocalVariable
	}{
		{
			name: "scalar parameter: IsParameter true, IsSeriesParameter false, IsLocalVariable false",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterParameter("src")
			},
			ident: "src", isParam: true, isSeries: false, isLocal: false,
		},
		{
			name: "series parameter: IsParameter true, IsSeriesParameter true, IsLocalVariable false",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterSeriesParameter("data")
			},
			ident: "data", isParam: true, isSeries: true, isLocal: false,
		},
		{
			name: "local variable: IsParameter false, IsSeriesParameter false, IsLocalVariable true",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLocalVariable("diff")
			},
			ident: "diff", isParam: false, isSeries: false, isLocal: true,
		},
		{
			name: "loop-modified variable: all inspection methods false",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLoopModified("acc")
			},
			ident: "acc", isParam: false, isSeries: false, isLocal: false,
		},
		{
			name: "unregistered identifier: all inspection methods false",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterParameter("other")
			},
			ident: "unknown", isParam: false, isSeries: false, isLocal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewArrowSeriesAccessResolver()
			tt.setup(r)

			if got := r.IsParameter(tt.ident); got != tt.isParam {
				t.Errorf("IsParameter(%q) = %v, want %v", tt.ident, got, tt.isParam)
			}
			if got := r.IsSeriesParameter(tt.ident); got != tt.isSeries {
				t.Errorf("IsSeriesParameter(%q) = %v, want %v", tt.ident, got, tt.isSeries)
			}
			if got := r.IsLocalVariable(tt.ident); got != tt.isLocal {
				t.Errorf("IsLocalVariable(%q) = %v, want %v", tt.ident, got, tt.isLocal)
			}
		})
	}
}

/*
TestArrowSeriesAccessResolver_ResolveAccess validates that all identifier categories return
the correct Go access expression and the found/not-found signal.
*/
func TestArrowSeriesAccessResolver_ResolveAccess(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(r *ArrowSeriesAccessResolver)
		ident        string
		wantCode     string
		wantResolved bool
	}{
		{
			name: "scalar parameter resolves to bare name",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterParameter("src")
			},
			ident: "src", wantCode: "src", wantResolved: true,
		},
		{
			name: "series parameter resolves to nameSeries.GetCurrent()",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterSeriesParameter("data")
			},
			ident: "data", wantCode: "dataSeries.GetCurrent()", wantResolved: true,
		},
		{
			name: "local variable resolves to bare name",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLocalVariable("diff")
			},
			ident: "diff", wantCode: "diff", wantResolved: true,
		},
		{
			name: "loop-modified variable resolves to nameSeries.GetCurrent()",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLoopModified("acc")
			},
			ident: "acc", wantCode: "accSeries.GetCurrent()", wantResolved: true,
		},
		{
			name: "unregistered identifier not resolved",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterParameter("other")
			},
			ident: "outerScopeVar", wantCode: "", wantResolved: false,
		},
		{
			name:  "empty resolver not resolved",
			setup: func(r *ArrowSeriesAccessResolver) {},
			ident: "anyVar", wantCode: "", wantResolved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewArrowSeriesAccessResolver()
			tt.setup(r)

			code, resolved := r.ResolveAccess(tt.ident)
			if resolved != tt.wantResolved {
				t.Errorf("ResolveAccess(%q) resolved = %v, want %v", tt.ident, resolved, tt.wantResolved)
			}
			if code != tt.wantCode {
				t.Errorf("ResolveAccess(%q) code = %q, want %q", tt.ident, code, tt.wantCode)
			}
		})
	}
}

/*
TestArrowSeriesAccessResolver_ResolutionPriority validates resolution priority ordering:
series parameters > scalar parameters > loop-modified > local variables.
Each test registers the same identifier in multiple categories to exercise precedence.
*/
func TestArrowSeriesAccessResolver_ResolutionPriority(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *ArrowSeriesAccessResolver)
		ident    string
		wantCode string
	}{
		{
			name: "series parameter takes priority over scalar parameter",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterParameter("x")
				r.RegisterSeriesParameter("x")
			},
			ident: "x", wantCode: "xSeries.GetCurrent()",
		},
		{
			name: "series parameter takes priority over local variable",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLocalVariable("x")
				r.RegisterSeriesParameter("x")
			},
			ident: "x", wantCode: "xSeries.GetCurrent()",
		},
		{
			name: "series parameter takes priority over loop-modified",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLoopModified("x")
				r.RegisterSeriesParameter("x")
			},
			ident: "x", wantCode: "xSeries.GetCurrent()",
		},
		{
			name: "scalar parameter takes priority over local variable",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLocalVariable("x")
				r.RegisterParameter("x")
			},
			ident: "x", wantCode: "x",
		},
		{
			name: "scalar parameter takes priority over loop-modified",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLoopModified("x")
				r.RegisterParameter("x")
			},
			ident: "x", wantCode: "x",
		},
		{
			name: "loop-modified takes priority over local variable",
			setup: func(r *ArrowSeriesAccessResolver) {
				r.RegisterLocalVariable("x")
				r.RegisterLoopModified("x")
			},
			ident: "x", wantCode: "xSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewArrowSeriesAccessResolver()
			tt.setup(r)

			code, resolved := r.ResolveAccess(tt.ident)
			if !resolved {
				t.Errorf("ResolveAccess(%q) = not resolved, want resolved", tt.ident)
			}
			if code != tt.wantCode {
				t.Errorf("ResolveAccess(%q) = %q, want %q", tt.ident, code, tt.wantCode)
			}
		})
	}
}

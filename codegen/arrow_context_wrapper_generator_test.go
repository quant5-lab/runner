package codegen

import (
	"strings"
	"testing"
)

func TestArrowContextWrapperGenerator_GenerateWrapper(t *testing.T) {
	tests := []struct {
		name       string
		scope      ArrowContextScope
		wantInCode []string
	}{
		{
			name: "single return value",
			scope: ArrowContextScope{
				FunctionName:   "dirmov",
				ContextVarName: "arrowCtx_dirmov",
				ResultVarNames: []string{"result"},
				ArgumentList:   "18.0",
			},
			wantInCode: []string{
				"arrowCtx_dirmov := context.NewArrowContext(ctx)",
				"result := dirmov(arrowCtx_dirmov, 18.0)",
				"arrowCtx_dirmov.AdvanceAll()",
			},
		},
		{
			name: "multiple return values",
			scope: ArrowContextScope{
				FunctionName:   "adx",
				ContextVarName: "arrowCtx_adx",
				ResultVarNames: []string{"ADX", "up", "down"},
				ArgumentList:   "18.0, 16.0",
			},
			wantInCode: []string{
				"arrowCtx_adx := context.NewArrowContext(ctx)",
				"ADX, up, down := adx(arrowCtx_adx, 18.0, 16.0)",
				"arrowCtx_adx.AdvanceAll()",
			},
		},
		{
			name: "no return values",
			scope: ArrowContextScope{
				FunctionName:   "helper",
				ContextVarName: "arrowCtx_helper",
				ResultVarNames: []string{},
				ArgumentList:   "10.0",
			},
			wantInCode: []string{
				"arrowCtx_helper := context.NewArrowContext(ctx)",
				"_ := helper(arrowCtx_helper, 10.0)",
				"arrowCtx_helper.AdvanceAll()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewArrowContextWrapperGenerator("\t")
			code := generator.GenerateWrapper(tt.scope)

			for _, expected := range tt.wantInCode {
				if !strings.Contains(code, expected) {
					t.Errorf("Generated code missing expected pattern %q\nGot:\n%s", expected, code)
				}
			}
		})
	}
}

func TestArrowContextWrapperGenerator_buildResultAssignment(t *testing.T) {
	tests := []struct {
		name     string
		varNames []string
		want     string
	}{
		{
			name:     "empty",
			varNames: []string{},
			want:     "_",
		},
		{
			name:     "single variable",
			varNames: []string{"result"},
			want:     "result",
		},
		{
			name:     "two variables",
			varNames: []string{"a", "b"},
			want:     "a, b",
		},
		{
			name:     "three variables",
			varNames: []string{"ADX", "up", "down"},
			want:     "ADX, up, down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewArrowContextWrapperGenerator("\t")
			got := generator.buildResultAssignment(tt.varNames)

			if got != tt.want {
				t.Errorf("buildResultAssignment() = %q, want %q", got, tt.want)
			}
		})
	}
}

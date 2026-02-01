package codegen

import "github.com/quant5-lab/runner/ast"

type TAResolvedCall struct {
	FunctionName         string
	SeriesArguments      []ast.Expression
	ScalarArguments      []ast.Expression
	RequiresOHLC         bool
	DefaultSourceApplied bool
	DefaultSourceName    string
}

func NewTAResolvedCall(functionName string) *TAResolvedCall {
	return &TAResolvedCall{
		FunctionName:    functionName,
		SeriesArguments: make([]ast.Expression, 0),
		ScalarArguments: make([]ast.Expression, 0),
	}
}

func (r *TAResolvedCall) AddSeriesArgument(expr ast.Expression) {
	r.SeriesArguments = append(r.SeriesArguments, expr)
}

func (r *TAResolvedCall) AddScalarArgument(expr ast.Expression) {
	r.ScalarArguments = append(r.ScalarArguments, expr)
}

func (r *TAResolvedCall) SetOHLCRequired() {
	r.RequiresOHLC = true
}

func (r *TAResolvedCall) ApplyDefaultSource(sourceName string) {
	r.DefaultSourceApplied = true
	r.DefaultSourceName = sourceName
	r.SeriesArguments = append([]ast.Expression{&ast.Identifier{Name: sourceName}}, r.SeriesArguments...)
}

func (r *TAResolvedCall) TotalArgumentCount() int {
	return len(r.SeriesArguments) + len(r.ScalarArguments)
}

func (r *TAResolvedCall) HasSeries() bool {
	return len(r.SeriesArguments) > 0
}

func (r *TAResolvedCall) HasScalars() bool {
	return len(r.ScalarArguments) > 0
}

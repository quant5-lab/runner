package context

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/series"
)

/*
ArrowContext provides isolated Series storage for arrow function local variables requiring historical access.

	Created once per call site, lazily initializes Series, advances cursors via AdvanceAll() after each bar.
*/
type ArrowContext struct {
	Context     *Context
	LocalSeries map[string]*series.Series
	capacity    int
}

/* NewArrowContext wraps Context with isolated Series map for arrow function local variables */
func NewArrowContext(ctx *Context) *ArrowContext {
	return &ArrowContext{
		Context:     ctx,
		LocalSeries: make(map[string]*series.Series),
		capacity:    len(ctx.Data),
	}
}

/* GetOrCreateSeries returns existing or creates new Series for variable name (lazy init) */
func (ac *ArrowContext) GetOrCreateSeries(name string) *series.Series {
	if s, exists := ac.LocalSeries[name]; exists {
		return s
	}

	s := series.NewSeries(ac.capacity)
	ac.LocalSeries[name] = s
	return s
}

/* AdvanceAll moves all local Series cursors forward by one bar */
func (ac *ArrowContext) AdvanceAll() {
	for _, s := range ac.LocalSeries {
		s.Next()
	}
}

/* GetSeries retrieves existing Series, returns error if not found */
func (ac *ArrowContext) GetSeries(name string) (*series.Series, error) {
	s, exists := ac.LocalSeries[name]
	if !exists {
		return nil, fmt.Errorf("arrow context: Series %q not found", name)
	}
	return s, nil
}

/* Reset moves all local Series cursors to specified position */
func (ac *ArrowContext) Reset(position int) {
	for _, s := range ac.LocalSeries {
		s.Reset(position)
	}
}

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
	Context            *Context
	LocalSeries        map[string]*series.Series
	SecurityContexts   map[string]*Context
	SecurityBarMappers map[string]BarIndexMapper
	ConcreteBarMappers map[string]interface{}
	SecurityEvaluators map[string]interface{}
	capacity           int
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

/* SetSecurityContext registers a security context by cache key */
func (ac *ArrowContext) SetSecurityContext(key string, ctx *Context) {
	if ac.SecurityContexts == nil {
		ac.SecurityContexts = make(map[string]*Context)
	}
	ac.SecurityContexts[key] = ctx
}

/* SetBarMapper registers a bar index mapper by cache key */
func (ac *ArrowContext) SetBarMapper(key string, mapper BarIndexMapper) {
	if ac.SecurityBarMappers == nil {
		ac.SecurityBarMappers = make(map[string]BarIndexMapper)
	}
	ac.SecurityBarMappers[key] = mapper
}

/* SetConcreteBarMapper stores the concrete bar mapper (e.g. *request.SecurityBarMapper) for arrow security eval */
func (ac *ArrowContext) SetConcreteBarMapper(key string, mapper interface{}) {
	if ac.ConcreteBarMappers == nil {
		ac.ConcreteBarMappers = make(map[string]interface{})
	}
	ac.ConcreteBarMappers[key] = mapper
}

/* GetOrCreateSecurityEvaluator returns the cached evaluator or nil */
func (ac *ArrowContext) GetOrCreateSecurityEvaluators() map[string]interface{} {
	if ac.SecurityEvaluators == nil {
		ac.SecurityEvaluators = make(map[string]interface{})
	}
	return ac.SecurityEvaluators
}

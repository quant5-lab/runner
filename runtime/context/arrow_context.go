package context

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/series"
)

// ArrowContext provides isolated Series storage for a single arrow function invocation scope.
// Created once per call site before the bar loop; persists across all bars.
// Nested UDF calls use child contexts (GetOrCreateChildContext) so each level of nesting
// has its own isolated Series namespace that also survives across bars.
type ArrowContext struct {
	Context            *Context
	LocalSeries        map[string]*series.Series
	SecurityContexts   map[string]*Context
	SecurityBarMappers map[string]BarIndexMapper
	ConcreteBarMappers map[string]interface{}
	SecurityEvaluators map[string]interface{}
	children           map[string]*ArrowContext
	capacity           int
}

func NewArrowContext(ctx *Context) *ArrowContext {
	return &ArrowContext{
		Context:     ctx,
		LocalSeries: make(map[string]*series.Series),
		capacity:    len(ctx.Data),
	}
}

// GetOrCreateSeries returns the existing series for name or lazily creates one.
func (ac *ArrowContext) GetOrCreateSeries(name string) *series.Series {
	if s, exists := ac.LocalSeries[name]; exists {
		return s
	}
	s := series.NewSeries(ac.capacity)
	ac.LocalSeries[name] = s
	return s
}

// GetOrCreateChildContext returns the persistent child context for the given key,
// creating it on first access.  Each nested UDF call site must use a unique key
// (conventionally "funcName_N" where N is the call-site index within the parent body)
// so that sibling calls to the same function have independent Series namespaces.
func (ac *ArrowContext) GetOrCreateChildContext(key string) *ArrowContext {
	if ac.children == nil {
		ac.children = make(map[string]*ArrowContext)
	}
	if child, exists := ac.children[key]; exists {
		return child
	}
	child := &ArrowContext{
		Context:     ac.Context,
		LocalSeries: make(map[string]*series.Series),
		capacity:    ac.capacity,
	}
	ac.children[key] = child
	return child
}

// AdvanceAll advances cursors for all local series and recurses into child contexts.
// Called once per bar after all UDF bodies have executed.
func (ac *ArrowContext) AdvanceAll() {
	for _, s := range ac.LocalSeries {
		s.Next()
	}
	for _, child := range ac.children {
		child.AdvanceAll()
	}
}

func (ac *ArrowContext) GetSeries(name string) (*series.Series, error) {
	s, exists := ac.LocalSeries[name]
	if !exists {
		return nil, fmt.Errorf("arrow context: Series %q not found", name)
	}
	return s, nil
}

func (ac *ArrowContext) Reset(position int) {
	for _, s := range ac.LocalSeries {
		s.Reset(position)
	}
	for _, child := range ac.children {
		child.Reset(position)
	}
}

func (ac *ArrowContext) SetSecurityContext(key string, ctx *Context) {
	if ac.SecurityContexts == nil {
		ac.SecurityContexts = make(map[string]*Context)
	}
	ac.SecurityContexts[key] = ctx
}

func (ac *ArrowContext) SetBarMapper(key string, mapper BarIndexMapper) {
	if ac.SecurityBarMappers == nil {
		ac.SecurityBarMappers = make(map[string]BarIndexMapper)
	}
	ac.SecurityBarMappers[key] = mapper
}

func (ac *ArrowContext) SetConcreteBarMapper(key string, mapper interface{}) {
	if ac.ConcreteBarMappers == nil {
		ac.ConcreteBarMappers = make(map[string]interface{})
	}
	ac.ConcreteBarMappers[key] = mapper
}

func (ac *ArrowContext) GetOrCreateSecurityEvaluators() map[string]interface{} {
	if ac.SecurityEvaluators == nil {
		ac.SecurityEvaluators = make(map[string]interface{})
	}
	return ac.SecurityEvaluators
}

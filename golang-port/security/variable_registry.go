package security

import "github.com/quant5-lab/runner/runtime/series"

type VariableRegistry struct {
	series map[string]*series.Series
}

func NewVariableRegistry() *VariableRegistry {
	return &VariableRegistry{
		series: make(map[string]*series.Series),
	}
}

func (r *VariableRegistry) Register(name string, s *series.Series) {
	r.series[name] = s
}

func (r *VariableRegistry) Get(name string) (*series.Series, bool) {
	s, ok := r.series[name]
	return s, ok
}

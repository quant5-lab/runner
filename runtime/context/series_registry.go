package context

import "github.com/quant5-lab/runner/runtime/series"

type SeriesRegistry interface {
	Get(name string) (*series.Series, bool)
	Set(name string, series *series.Series)
}

type MapBasedRegistry struct {
	storage map[string]*series.Series
}

func NewMapBasedRegistry() *MapBasedRegistry {
	return &MapBasedRegistry{
		storage: make(map[string]*series.Series),
	}
}

func (r *MapBasedRegistry) Get(name string) (*series.Series, bool) {
	series, found := r.storage[name]
	return series, found
}

func (r *MapBasedRegistry) Set(name string, series *series.Series) {
	r.storage[name] = series
}

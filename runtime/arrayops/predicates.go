package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type Predicates struct{}

func NewPredicates() *Predicates {
	return &Predicates{}
}

func (p *Predicates) Every(arr *series.ArraySeries, offset int) bool {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return true
	}

	for _, v := range slice {
		if v == 0.0 {
			return false
		}
	}
	return true
}

func (p *Predicates) Some(arr *series.ArraySeries, offset int) bool {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return false
	}

	for _, v := range slice {
		if v != 0.0 {
			return true
		}
	}
	return false
}

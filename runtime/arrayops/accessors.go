package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type Accessor struct{}

func NewAccessor() *Accessor {
	return &Accessor{}
}

func (a *Accessor) First(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return 0.0
	}
	return slice[0]
}

func (a *Accessor) Last(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return 0.0
	}
	return slice[len(slice)-1]
}

func (a *Accessor) Includes(s *series.ArraySeries, offset int, value float64) bool {
	slice := s.Get(offset)
	if slice == nil {
		return false
	}
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (a *Accessor) IndexOf(s *series.ArraySeries, offset int, value float64, startIndex int) int {
	slice := s.Get(offset)
	if slice == nil {
		return -1
	}
	if startIndex < 0 {
		startIndex = 0
	}
	for i := startIndex; i < len(slice); i++ {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

func (a *Accessor) LastIndexOf(s *series.ArraySeries, offset int, value float64, startIndex int) int {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return -1
	}
	if startIndex < 0 || startIndex >= len(slice) {
		startIndex = len(slice) - 1
	}
	for i := startIndex; i >= 0; i-- {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

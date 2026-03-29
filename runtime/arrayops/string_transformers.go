package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type StringTransformers struct{}

func NewStringTransformers() *StringTransformers {
	return &StringTransformers{}
}

func (t *StringTransformers) Reverse(s *series.StringArraySeries) {
	current := s.GetCurrent()
	if len(current) <= 1 {
		return
	}
	next := make([]string, len(current))
	for i, v := range current {
		next[len(current)-1-i] = v
	}
	s.Set(next)
}

func (t *StringTransformers) Concat(s1 *series.StringArraySeries, offset1 int, s2 *series.StringArraySeries, offset2 int) []string {
	slice1 := s1.Get(offset1)
	slice2 := s2.Get(offset2)
	result := make([]string, len(slice1)+len(slice2))
	copy(result, slice1)
	copy(result[len(slice1):], slice2)
	return result
}

func (t *StringTransformers) Slice(s *series.StringArraySeries, offset int, indexFrom, indexTo int) []string {
	current := s.Get(offset)
	if len(current) == 0 {
		return []string{}
	}
	if indexFrom < 0 {
		indexFrom = 0
	}
	if indexTo < 0 || indexTo > len(current) {
		indexTo = len(current)
	}
	if indexFrom >= indexTo {
		return []string{}
	}
	result := make([]string, indexTo-indexFrom)
	copy(result, current[indexFrom:indexTo])
	return result
}

func (t *StringTransformers) Copy(s *series.StringArraySeries, offset int) []string {
	current := s.Get(offset)
	if current == nil {
		return nil
	}
	result := make([]string, len(current))
	copy(result, current)
	return result
}

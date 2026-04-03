package arrayops

import (
	"sort"

	"github.com/quant5-lab/runner/runtime/series"
)

type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) Copy(s *series.ArraySeries, offset int) []float64 {
	slice := s.Get(offset)
	if slice == nil {
		return nil
	}
	copied := make([]float64, len(slice))
	copy(copied, slice)
	return copied
}

func (t *Transformer) Slice(s *series.ArraySeries, offset, indexFrom, indexTo int) []float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return nil
	}
	if indexFrom < 0 {
		indexFrom = 0
	}
	if indexTo < 0 || indexTo > len(slice) {
		indexTo = len(slice)
	}
	if indexFrom >= indexTo {
		return []float64{}
	}
	result := make([]float64, indexTo-indexFrom)
	copy(result, slice[indexFrom:indexTo])
	return result
}

func (t *Transformer) Reverse(s *series.ArraySeries) {
	current := s.GetCurrent()
	if len(current) <= 1 {
		return
	}
	reversed := make([]float64, len(current))
	for i := 0; i < len(current); i++ {
		reversed[i] = current[len(current)-1-i]
	}
	s.Set(reversed)
}

func (t *Transformer) Sort(s *series.ArraySeries, order string) {
	current := s.GetCurrent()
	if len(current) <= 1 {
		return
	}
	sorted := make([]float64, len(current))
	copy(sorted, current)
	if order == "order.descending" {
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })
	} else {
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	}
	s.Set(sorted)
}

func (t *Transformer) SortIndices(s *series.ArraySeries, offset int, order string) []float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return []float64{}
	}
	indices := make([]int, len(slice))
	for i := range indices {
		indices[i] = i
	}
	if order == "order.descending" {
		sort.Slice(indices, func(i, j int) bool {
			return slice[indices[i]] > slice[indices[j]]
		})
	} else {
		sort.Slice(indices, func(i, j int) bool {
			return slice[indices[i]] < slice[indices[j]]
		})
	}
	result := make([]float64, len(indices))
	for i, idx := range indices {
		result[i] = float64(idx)
	}
	return result
}

func (t *Transformer) Concat(s1 *series.ArraySeries, offset1 int, s2 *series.ArraySeries, offset2 int) []float64 {
	slice1 := s1.Get(offset1)
	slice2 := s2.Get(offset2)
	if slice1 == nil && slice2 == nil {
		return nil
	}
	if slice1 == nil {
		return t.Copy(s2, offset2)
	}
	if slice2 == nil {
		return t.Copy(s1, offset1)
	}
	result := make([]float64, len(slice1)+len(slice2))
	copy(result, slice1)
	copy(result[len(slice1):], slice2)
	return result
}

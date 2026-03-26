package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type Mutator struct{}

func NewMutator() *Mutator {
	return &Mutator{}
}

func (m *Mutator) Push(s *series.ArraySeries, value float64) {
	current := s.GetCurrent()
	if current == nil {
		s.Set([]float64{value})
		return
	}
	next := make([]float64, len(current)+1)
	copy(next, current)
	next[len(current)] = value
	s.Set(next)
}

func (m *Mutator) Pop(s *series.ArraySeries) float64 {
	current := s.GetCurrent()
	if len(current) == 0 {
		s.Set(nil)
		return 0.0
	}
	popped := current[len(current)-1]
	next := make([]float64, len(current)-1)
	copy(next, current[:len(current)-1])
	s.Set(next)
	return popped
}

func (m *Mutator) Shift(s *series.ArraySeries) float64 {
	current := s.GetCurrent()
	if len(current) == 0 {
		s.Set(nil)
		return 0.0
	}
	shifted := current[0]
	next := make([]float64, len(current)-1)
	copy(next, current[1:])
	s.Set(next)
	return shifted
}

func (m *Mutator) Unshift(s *series.ArraySeries, value float64) {
	current := s.GetCurrent()
	if current == nil {
		s.Set([]float64{value})
		return
	}
	next := make([]float64, len(current)+1)
	next[0] = value
	copy(next[1:], current)
	s.Set(next)
}

func (m *Mutator) SetElement(s *series.ArraySeries, index int, value float64) {
	current := s.GetCurrent()
	if current == nil || index < 0 || index >= len(current) {
		return
	}
	next := make([]float64, len(current))
	copy(next, current)
	next[index] = value
	s.Set(next)
}

func (m *Mutator) Insert(s *series.ArraySeries, index int, value float64) {
	current := s.GetCurrent()
	if current == nil {
		if index == 0 {
			s.Set([]float64{value})
		}
		return
	}
	if index < 0 || index > len(current) {
		return
	}
	next := make([]float64, len(current)+1)
	copy(next[:index], current[:index])
	next[index] = value
	copy(next[index+1:], current[index:])
	s.Set(next)
}

func (m *Mutator) Remove(s *series.ArraySeries, index int) float64 {
	current := s.GetCurrent()
	if current == nil || index < 0 || index >= len(current) {
		return 0.0
	}
	removed := current[index]
	next := make([]float64, len(current)-1)
	copy(next[:index], current[:index])
	copy(next[index:], current[index+1:])
	s.Set(next)
	return removed
}

func (m *Mutator) Clear(s *series.ArraySeries) {
	s.Set([]float64{})
}

func (m *Mutator) Fill(s *series.ArraySeries, value float64, indexFrom, indexTo int) {
	current := s.GetCurrent()
	if len(current) == 0 {
		return
	}
	if indexFrom < 0 {
		indexFrom = 0
	}
	if indexTo < 0 || indexTo > len(current) {
		indexTo = len(current)
	}
	if indexFrom >= indexTo {
		return
	}
	next := make([]float64, len(current))
	copy(next, current)
	for i := indexFrom; i < indexTo; i++ {
		next[i] = value
	}
	s.Set(next)
}

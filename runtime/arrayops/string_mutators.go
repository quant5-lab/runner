package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type StringMutator struct{}

func NewStringMutator() *StringMutator {
	return &StringMutator{}
}

func (m *StringMutator) Push(s *series.StringArraySeries, value string) {
	current := s.GetCurrent()
	if current == nil {
		s.Set([]string{value})
		return
	}
	next := make([]string, len(current)+1)
	copy(next, current)
	next[len(current)] = value
	s.Set(next)
}

func (m *StringMutator) Pop(s *series.StringArraySeries) string {
	current := s.GetCurrent()
	if len(current) == 0 {
		s.Set(nil)
		return ""
	}
	popped := current[len(current)-1]
	next := make([]string, len(current)-1)
	copy(next, current[:len(current)-1])
	s.Set(next)
	return popped
}

func (m *StringMutator) Shift(s *series.StringArraySeries) string {
	current := s.GetCurrent()
	if len(current) == 0 {
		s.Set(nil)
		return ""
	}
	shifted := current[0]
	next := make([]string, len(current)-1)
	copy(next, current[1:])
	s.Set(next)
	return shifted
}

func (m *StringMutator) Unshift(s *series.StringArraySeries, value string) {
	current := s.GetCurrent()
	if current == nil {
		s.Set([]string{value})
		return
	}
	next := make([]string, len(current)+1)
	next[0] = value
	copy(next[1:], current)
	s.Set(next)
}

func (m *StringMutator) SetElement(s *series.StringArraySeries, index int, value string) {
	current := s.GetCurrent()
	if current == nil || index < 0 || index >= len(current) {
		return
	}
	next := make([]string, len(current))
	copy(next, current)
	next[index] = value
	s.Set(next)
}

func (m *StringMutator) Insert(s *series.StringArraySeries, index int, value string) {
	current := s.GetCurrent()
	if current == nil {
		if index == 0 {
			s.Set([]string{value})
		}
		return
	}
	if index < 0 || index > len(current) {
		return
	}
	next := make([]string, len(current)+1)
	copy(next[:index], current[:index])
	next[index] = value
	copy(next[index+1:], current[index:])
	s.Set(next)
}

func (m *StringMutator) Remove(s *series.StringArraySeries, index int) string {
	current := s.GetCurrent()
	if current == nil || index < 0 || index >= len(current) {
		return ""
	}
	removed := current[index]
	next := make([]string, len(current)-1)
	copy(next[:index], current[:index])
	copy(next[index:], current[index+1:])
	s.Set(next)
	return removed
}

func (m *StringMutator) Clear(s *series.StringArraySeries) {
	s.Set([]string{})
}

func (m *StringMutator) Fill(s *series.StringArraySeries, value string, indexFrom, indexTo int) {
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
	next := make([]string, len(current))
	copy(next, current)
	for i := indexFrom; i < indexTo; i++ {
		next[i] = value
	}
	s.Set(next)
}

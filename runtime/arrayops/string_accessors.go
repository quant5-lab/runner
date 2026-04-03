package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type StringAccessor struct{}

func NewStringAccessor() *StringAccessor {
	return &StringAccessor{}
}

func (a *StringAccessor) First(s *series.StringArraySeries, offset int) string {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return ""
	}
	return slice[0]
}

func (a *StringAccessor) Last(s *series.StringArraySeries, offset int) string {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return ""
	}
	return slice[len(slice)-1]
}

func (a *StringAccessor) Includes(s *series.StringArraySeries, offset int, value string) bool {
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

func (a *StringAccessor) IndexOf(s *series.StringArraySeries, offset int, value string, startIndex int) int {
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

func (a *StringAccessor) LastIndexOf(s *series.StringArraySeries, offset int, value string, startIndex int) int {
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

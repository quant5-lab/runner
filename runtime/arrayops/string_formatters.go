package arrayops

import (
	"strings"

	"github.com/quant5-lab/runner/runtime/series"
)

type StringFormatters struct{}

func NewStringFormatters() *StringFormatters {
	return &StringFormatters{}
}

func (f *StringFormatters) Join(arr *series.StringArraySeries, offset int, separator string) string {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return ""
	}
	return strings.Join(slice, separator)
}

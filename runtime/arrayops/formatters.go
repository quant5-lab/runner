package arrayops

import (
	"strconv"
	"strings"

	"github.com/quant5-lab/runner/runtime/series"
)

type Formatters struct{}

func NewFormatters() *Formatters {
	return &Formatters{}
}

func (f *Formatters) Join(arr *series.ArraySeries, offset int, separator string) string {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return ""
	}

	parts := make([]string, len(slice))
	for i, v := range slice {
		parts[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	return strings.Join(parts, separator)
}

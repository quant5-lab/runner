package arrayops

import "github.com/quant5-lab/runner/runtime/series"

type Search struct{}

func NewSearch() *Search {
	return &Search{}
}

func (s *Search) BinarySearch(arr *series.ArraySeries, offset int, value float64) int {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return -1
	}

	left, right := 0, len(slice)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2

		if slice[mid] == value {
			return mid
		}

		if slice[mid] < value {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

func (s *Search) BinarySearchLeftmost(arr *series.ArraySeries, offset int, value float64) int {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return -1
	}

	left, right := 0, len(slice)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2

		if slice[mid] == value {
			result = mid
			right = mid - 1
		} else if slice[mid] < value {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

func (s *Search) BinarySearchRightmost(arr *series.ArraySeries, offset int, value float64) int {
	slice := arr.Get(offset)
	if len(slice) == 0 {
		return -1
	}

	left, right := 0, len(slice)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2

		if slice[mid] == value {
			result = mid
			left = mid + 1
		} else if slice[mid] < value {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

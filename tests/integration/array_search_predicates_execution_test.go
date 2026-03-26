//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestArrayBinarySearchCodegen(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Binary Search Test", overlay=false)

prices = array.from(1, 3, 5, 5, 5, 7, 9)
found = array.binary_search(prices, 5.0)
leftmost = array.binary_search_leftmost(prices, 5.0)
rightmost = array.binary_search_rightmost(prices, 5.0)
notFound = array.binary_search(prices, 4.0)

plot(found, "Found")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "binary-search-codegen", pineScript)

	t.Logf("Generated code:\n%s", goCode)

	requiredCode := []string{
		"arrayops.NewSearch().BinarySearch(pricesArraySeries, 0, 5)",
		"arrayops.NewSearch().BinarySearchLeftmost(pricesArraySeries, 0, 5)",
		"arrayops.NewSearch().BinarySearchRightmost(pricesArraySeries, 0, 5)",
		"arrayops.NewSearch().BinarySearch(pricesArraySeries, 0, 4)",
	}

	for _, exp := range requiredCode {
		if !contains(goCode, exp) {
			t.Errorf("Missing codegen: %s", exp)
		}
	}

	exec.ExecuteScript(t, "binary-search-exec", pineScript)
}

func TestArrayPredicatesCodegen(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Predicates Test", overlay=false)

bools = array.from(1, 1, 0, 1)
allTrue = array.from(1, 1, 1)
allFalse = array.from(0, 0, 0)

everyResult = array.every(allTrue)
someResult = array.some(bools)
everyFalse = array.every(allFalse)
someFalse = array.some(allFalse)

plot(everyResult ? 1 : 0, "Every")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "predicates-codegen", pineScript)

	requiredCode := []string{
		"arrayops.NewPredicates().Every(allTrueArraySeries, 0)",
		"arrayops.NewPredicates().Some(boolsArraySeries, 0)",
		"arrayops.NewPredicates().Every(allFalseArraySeries, 0)",
		"arrayops.NewPredicates().Some(allFalseArraySeries, 0)",
	}

	for _, exp := range requiredCode {
		if !contains(goCode, exp) {
			t.Errorf("Missing codegen: %s", exp)
		}
	}

	exec.ExecuteScript(t, "predicates-exec", pineScript)
}

func TestArrayJoinCodegen(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Join Test", overlay=false)

nums = array.from(10, 20, 30)
result1 = array.join(nums, ", ")
result2 = array.join(nums, " | ")
result3 = array.join(nums)

plot(1, "Dummy")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "join-codegen", pineScript)

	requiredCode := []string{
		`arrayops.NewFormatters().Join(numsArraySeries, 0, ", ")`,
		`arrayops.NewFormatters().Join(numsArraySeries, 0, " | ")`,
		`arrayops.NewFormatters().Join(numsArraySeries, 0, ", ")`,
	}

	for _, exp := range requiredCode {
		if !contains(goCode, exp) {
			t.Errorf("Missing codegen: %s", exp)
		}
	}

	exec.ExecuteScript(t, "join-exec", pineScript)
}

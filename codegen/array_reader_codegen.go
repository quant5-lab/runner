package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ArrayReaderCodegen struct{}

func NewArrayReaderCodegen() *ArrayReaderCodegen {
	return &ArrayReaderCodegen{}
}

func (h *ArrayReaderCodegen) CanHandle(funcName string) bool {
	classifier := NewArrayConstructorClassifier()
	return classifier.IsReadOnlyMethod(funcName)
}

func (h *ArrayReaderCodegen) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "array.first":
		return h.generateFirst(g, call)
	case "array.last":
		return h.generateLast(g, call)
	case "array.includes":
		return h.generateIncludes(g, call)
	case "array.indexof":
		return h.generateIndexOf(g, call)
	case "array.lastindexof":
		return h.generateLastIndexOf(g, call)
	case "array.slice":
		return h.generateSlice(g, call)
	case "array.copy":
		return h.generateCopy(g, call)
	case "array.sort_indices":
		return h.generateSortIndices(g, call)
	case "array.sum":
		return h.generateSum(g, call)
	case "array.avg":
		return h.generateAvg(g, call)
	case "array.min":
		return h.generateMin(g, call)
	case "array.max":
		return h.generateMax(g, call)
	case "array.median":
		return h.generateMedian(g, call)
	case "array.mode":
		return h.generateMode(g, call)
	case "array.stdev":
		return h.generateStdev(g, call)
	case "array.variance":
		return h.generateVariance(g, call)
	case "array.range":
		return h.generateRange(g, call)
	case "array.percentile_linear_interpolation":
		return h.generatePercentileLinear(g, call)
	case "array.percentile_nearest_rank":
		return h.generatePercentileNearest(g, call)
	case "array.percentrank":
		return h.generatePercentRank(g, call)
	case "array.covariance":
		return h.generateCovariance(g, call)
	case "array.standardize":
		return h.generateStandardize(g, call)
	case "array.abs":
		return h.generateAbs(g, call)
	case "array.binary_search":
		return h.generateBinarySearch(g, call, "BinarySearch")
	case "array.binary_search_leftmost":
		return h.generateBinarySearch(g, call, "BinarySearchLeftmost")
	case "array.binary_search_rightmost":
		return h.generateBinarySearch(g, call, "BinarySearchRightmost")
	case "array.every":
		return h.generatePredicate(g, call, "array.every", "Every")
	case "array.some":
		return h.generatePredicate(g, call, "array.some", "Some")
	case "array.join":
		return h.generateJoin(g, call)
	}

	return "", nil
}

func (h *ArrayReaderCodegen) generateFirst(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.first requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.first: %w", err)
	}

	return fmt.Sprintf("%s.First(%s%s, %d)",
		elemType.AccessorConstructor(), arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateLast(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.last requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.last: %w", err)
	}

	return fmt.Sprintf("%s.Last(%s%s, %d)",
		elemType.AccessorConstructor(), arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateIncludes(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.includes requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.includes: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.includes: value: %w", err)
	}

	return fmt.Sprintf("func() float64 { if %s.Includes(%s%s, %d, %s) { return 1.0 }; return 0.0 }()",
		elemType.AccessorConstructor(), arrayVar, elemType.VariableSuffix(), offset, valueCode), nil
}

func (h *ArrayReaderCodegen) generateIndexOf(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
		return "", fmt.Errorf("array.indexof requires 2-3 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.indexof: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.indexof: value: %w", err)
	}

	startIndex := "0"
	if len(call.Arguments) == 3 {
		startCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
		if err != nil {
			return "", fmt.Errorf("array.indexof: startIndex: %w", err)
		}
		startIndex = fmt.Sprintf("int(%s)", startCode)
	}

	return fmt.Sprintf("float64(%s.IndexOf(%s%s, %d, %s, %s))",
		elemType.AccessorConstructor(), arrayVar, elemType.VariableSuffix(), offset, valueCode, startIndex), nil
}

func (h *ArrayReaderCodegen) generateLastIndexOf(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
		return "", fmt.Errorf("array.lastindexof requires 2-3 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.lastindexof: %w", err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.lastindexof: value: %w", err)
	}

	startIndex := "-1"
	if len(call.Arguments) == 3 {
		startCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
		if err != nil {
			return "", fmt.Errorf("array.lastindexof: startIndex: %w", err)
		}
		startIndex = fmt.Sprintf("int(%s)", startCode)
	}

	return fmt.Sprintf("float64(%s.LastIndexOf(%s%s, %d, %s, %s))",
		elemType.AccessorConstructor(), arrayVar, elemType.VariableSuffix(), offset, valueCode, startIndex), nil
}

func (h *ArrayReaderCodegen) generateSlice(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 3 {
		return "", fmt.Errorf("array.slice requires 1-3 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.slice: %w", err)
	}

	indexFrom := "0"
	indexTo := "-1"

	if len(call.Arguments) >= 2 {
		fromCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("array.slice: indexFrom: %w", err)
		}
		indexFrom = fmt.Sprintf("int(%s)", fromCode)
	}

	if len(call.Arguments) == 3 {
		toCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
		if err != nil {
			return "", fmt.Errorf("array.slice: indexTo: %w", err)
		}
		indexTo = fmt.Sprintf("int(%s)", toCode)
	}

	return fmt.Sprintf("%s.Slice(%s%s, %d, %s, %s)",
		elemType.TransformerConstructor(), arrayVar, elemType.VariableSuffix(), offset, indexFrom, indexTo), nil
}

func (h *ArrayReaderCodegen) generateCopy(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.copy requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.copy: %w", err)
	}

	return fmt.Sprintf("%s.Copy(%s%s, %d)",
		elemType.TransformerConstructor(), arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateSortIndices(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
		return "", fmt.Errorf("array.sort_indices requires 1-2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.sort_indices: %w", err)
	}

	order := `"order.ascending"`
	if len(call.Arguments) == 2 {
		orderCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("array.sort_indices: order: %w", err)
		}
		order = orderCode
	}

	return fmt.Sprintf("%s.SortIndices(%s%s, %d, %s)",
		elemType.TransformerConstructor(), arrayVar, elemType.VariableSuffix(), offset, order), nil
}

func (h *ArrayReaderCodegen) generateSum(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.sum", "Sum")
}

func (h *ArrayReaderCodegen) generateAvg(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.avg", "Avg")
}

func (h *ArrayReaderCodegen) generateMin(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.min", "Min")
}

func (h *ArrayReaderCodegen) generateMax(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.max", "Max")
}

func (h *ArrayReaderCodegen) generateMedian(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.median", "Median")
}

func (h *ArrayReaderCodegen) generateMode(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.mode", "Mode")
}

func (h *ArrayReaderCodegen) generateRange(g *generator, call *ast.CallExpression) (string, error) {
	return h.generateSimpleStatistic(g, call, "array.range", "Range")
}

func (h *ArrayReaderCodegen) generateStdev(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
		return "", fmt.Errorf("array.stdev requires 1-2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.stdev: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.stdev requires numeric array, got %s array", elemType.GoType())
	}

	biased := "false"
	if len(call.Arguments) == 2 {
		biasedCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("array.stdev: biased: %w", err)
		}
		biased = fmt.Sprintf("%s != 0.0", biasedCode)
	}

	return fmt.Sprintf("%s.Stdev(%s%s, %d, %s)",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset, biased), nil
}

func (h *ArrayReaderCodegen) generateVariance(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
		return "", fmt.Errorf("array.variance requires 1-2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.variance: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.variance requires numeric array, got %s array", elemType.GoType())
	}

	biased := "false"
	if len(call.Arguments) == 2 {
		biasedCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("array.variance: biased: %w", err)
		}
		biased = fmt.Sprintf("%s != 0.0", biasedCode)
	}

	return fmt.Sprintf("%s.Variance(%s%s, %d, %s)",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset, biased), nil
}

func (h *ArrayReaderCodegen) generatePercentileLinear(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.percentile_linear_interpolation requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.percentile_linear_interpolation: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.percentile_linear_interpolation requires numeric array, got %s array", elemType.GoType())
	}

	percentileCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.percentile_linear_interpolation: percentile: %w", err)
	}

	return fmt.Sprintf("%s.Percentile(%s%s, %d, %s, \"linear\")",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset, percentileCode), nil
}

func (h *ArrayReaderCodegen) generatePercentileNearest(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.percentile_nearest_rank requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.percentile_nearest_rank: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.percentile_nearest_rank requires numeric array, got %s array", elemType.GoType())
	}

	percentileCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.percentile_nearest_rank: percentile: %w", err)
	}

	return fmt.Sprintf("%s.Percentile(%s%s, %d, %s, \"nearest_rank\")",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset, percentileCode), nil
}

func (h *ArrayReaderCodegen) generatePercentRank(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.percentrank requires 2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.percentrank: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.percentrank requires numeric array, got %s array", elemType.GoType())
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.percentrank: value: %w", err)
	}

	return fmt.Sprintf("%s.PercentRank(%s%s, %d, %s)",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset, valueCode), nil
}

func (h *ArrayReaderCodegen) generateCovariance(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
		return "", fmt.Errorf("array.covariance requires 2-3 arguments, got %d", len(call.Arguments))
	}

	array1Var, offset1, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.covariance: first array: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.covariance requires numeric array, got %s array", elemType.GoType())
	}

	array2Var, offset2, _, err := h.extractArrayAccess(g, call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.covariance: second array: %w", err)
	}

	biased := "false"
	if len(call.Arguments) == 3 {
		biasedCode, err := g.generateArrowFunctionExpression(call.Arguments[2])
		if err != nil {
			return "", fmt.Errorf("array.covariance: biased: %w", err)
		}
		biased = fmt.Sprintf("%s != 0.0", biasedCode)
	}

	return fmt.Sprintf("%s.Covariance(%s%s, %d, %s%s, %d, %s)",
		elemType.StatisticsConstructor(),
		array1Var, elemType.VariableSuffix(), offset1,
		array2Var, elemType.VariableSuffix(), offset2,
		biased), nil
}

func (h *ArrayReaderCodegen) generateStandardize(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.standardize requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.standardize: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.standardize requires numeric array, got %s array", elemType.GoType())
	}

	return fmt.Sprintf("%s.Standardize(%s%s, %d)",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateAbs(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("array.abs requires 1 argument, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.abs: %w", err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("array.abs requires numeric array, got %s array", elemType.GoType())
	}

	return fmt.Sprintf("%s.Abs(%s%s, %d)",
		elemType.StatisticsConstructor(), arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateSimpleStatistic(g *generator, call *ast.CallExpression, funcName, method string) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("%s requires 1 argument, got %d", funcName, len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("%s: %w", funcName, err)
	}

	if !elemType.SupportsStatistics() {
		return "", fmt.Errorf("%s requires numeric array, got %s array", funcName, elemType.GoType())
	}

	return fmt.Sprintf("%s.%s(%s%s, %d)",
		elemType.StatisticsConstructor(), method, arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateBinarySearch(g *generator, call *ast.CallExpression, method string) (string, error) {
	if len(call.Arguments) != 2 {
		return "", fmt.Errorf("array.%s requires 2 arguments, got %d", method, len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.%s: %w", method, err)
	}

	valueCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.%s: value: %w", method, err)
	}

	return fmt.Sprintf("float64(%s.%s(%s%s, %d, %s))",
		elemType.SearchConstructor(), method, arrayVar, elemType.VariableSuffix(), offset, valueCode), nil
}

func (h *ArrayReaderCodegen) generatePredicate(g *generator, call *ast.CallExpression, funcName, method string) (string, error) {
	if len(call.Arguments) != 1 {
		return "", fmt.Errorf("%s requires 1 argument, got %d", funcName, len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("%s: %w", funcName, err)
	}

	return fmt.Sprintf("func() float64 { if %s.%s(%s%s, %d) { return 1.0 }; return 0.0 }()",
		elemType.PredicatesConstructor(), method, arrayVar, elemType.VariableSuffix(), offset), nil
}

func (h *ArrayReaderCodegen) generateJoin(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
		return "", fmt.Errorf("array.join requires 1-2 arguments, got %d", len(call.Arguments))
	}

	arrayVar, offset, elemType, err := h.extractArrayAccess(g, call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.join: %w", err)
	}

	separator := `", "`
	if len(call.Arguments) == 2 {
		sepCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("array.join: separator: %w", err)
		}
		separator = sepCode
	}

	return fmt.Sprintf("%s.Join(%s%s, %d, %s)",
		elemType.FormattersConstructor(), arrayVar, elemType.VariableSuffix(), offset, separator), nil
}

func (h *ArrayReaderCodegen) extractArrayAccess(g *generator, arg ast.Expression) (arrayVar string, offset int, elemType ArrayElementType, err error) {
	switch e := arg.(type) {
	case *ast.Identifier:
		var ok bool
		elemType, ok = g.lookupArrayElementType(e.Name)
		if !ok {
			err = fmt.Errorf("variable %q is not an array", e.Name)
			return
		}
		arrayVar = e.Name
		offset = 0
		return

	case *ast.MemberExpression:
		if !e.Computed {
			break
		}
		id, ok := e.Object.(*ast.Identifier)
		if !ok {
			break
		}
		elemType, ok = g.lookupArrayElementType(id.Name)
		if !ok {
			break
		}
		arrayVar = id.Name
		if lit, ok := e.Property.(*ast.Literal); ok {
			switch v := lit.Value.(type) {
			case int:
				offset = v
				return
			case float64:
				offset = int(v)
				return
			}
		}
	}

	err = fmt.Errorf("expected array variable or array[n]")
	return
}

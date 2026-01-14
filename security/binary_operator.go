package security

import (
	"fmt"
	"math"
)

func applyBinaryOperator(operator string, left, right float64) (float64, error) {
	switch operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0.0 {
			return math.NaN(), nil
		}
		return left / right, nil
	case "%":
		if right == 0.0 {
			return math.NaN(), nil
		}
		return math.Mod(left, right), nil
	case ">":
		if left > right {
			return 1.0, nil
		}
		return 0.0, nil
	case ">=":
		if left >= right {
			return 1.0, nil
		}
		return 0.0, nil
	case "<":
		if left < right {
			return 1.0, nil
		}
		return 0.0, nil
	case "<=":
		if left <= right {
			return 1.0, nil
		}
		return 0.0, nil
	case "==":
		if math.Abs(left-right) < 1e-10 {
			return 1.0, nil
		}
		return 0.0, nil
	case "!=":
		if math.Abs(left-right) >= 1e-10 {
			return 1.0, nil
		}
		return 0.0, nil
	case "and":
		if left != 0.0 && right != 0.0 {
			return 1.0, nil
		}
		return 0.0, nil
	case "or":
		if left != 0.0 || right != 0.0 {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0.0, fmt.Errorf("unsupported binary operator: %s", operator)
	}
}

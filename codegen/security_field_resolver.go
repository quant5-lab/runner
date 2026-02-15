package codegen

import "fmt"

func SecurityBarFieldExpression(fieldName, barAccess string) (string, bool) {
	switch fieldName {
	case "close":
		return barAccess + ".Close", true
	case "open":
		return barAccess + ".Open", true
	case "high":
		return barAccess + ".High", true
	case "low":
		return barAccess + ".Low", true
	case "volume":
		return barAccess + ".Volume", true
	case "ohlc4":
		return fmt.Sprintf("(%s.Open + %s.High + %s.Low + %s.Close) / 4", barAccess, barAccess, barAccess, barAccess), true
	case "hlc3":
		return fmt.Sprintf("(%s.High + %s.Low + %s.Close) / 3", barAccess, barAccess, barAccess), true
	case "hl2":
		return fmt.Sprintf("(%s.High + %s.Low) / 2", barAccess, barAccess), true
	case "hlcc4":
		return fmt.Sprintf("(%s.High + %s.Low + %s.Close + %s.Close) / 4", barAccess, barAccess, barAccess, barAccess), true
	}
	return "", false
}

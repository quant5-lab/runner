package codegen

import "fmt"

type DerivedPriceFormulaGenerator struct{}

func NewDerivedPriceFormulaGenerator() *DerivedPriceFormulaGenerator {
	return &DerivedPriceFormulaGenerator{}
}

func (g *DerivedPriceFormulaGenerator) Generate(priceName string, highAccess, lowAccess, closeAccess, openAccess string) string {
	switch priceName {
	case "hl2":
		return fmt.Sprintf("((%s + %s) / 2)", highAccess, lowAccess)
	case "hlc3":
		return fmt.Sprintf("((%s + %s + %s) / 3)", highAccess, lowAccess, closeAccess)
	case "ohlc4":
		return fmt.Sprintf("((%s + %s + %s + %s) / 4)", openAccess, highAccess, lowAccess, closeAccess)
	case "hlcc4":
		return fmt.Sprintf("((%s + %s + %s + %s) / 4)", highAccess, lowAccess, closeAccess, closeAccess)
	default:
		return ""
	}
}

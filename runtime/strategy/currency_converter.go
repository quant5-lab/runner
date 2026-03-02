package strategy

type CurrencyConverter struct{}

func NewCurrencyConverter() *CurrencyConverter {
	return &CurrencyConverter{}
}

func (c *CurrencyConverter) ToAccount(value float64) float64 {
	return value
}

func (c *CurrencyConverter) ToSymbol(value float64) float64 {
	return value
}

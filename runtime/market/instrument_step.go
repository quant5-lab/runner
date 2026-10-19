package market

// Binance lot sizes are symbol-specific and unavailable offline; 0 signals the
// caller to skip rounding.
func InstrumentQtyStep(exchange Exchange) float64 {
	switch exchange {
	case ExchangeBinance:
		return 0
	default:
		return 1
	}
}

func ResolveQtyStep(exchange Exchange, metadata SourceMetadata) float64 {
	if metadata.QtyStep > 0 {
		return metadata.QtyStep
	}
	return InstrumentQtyStep(exchange)
}

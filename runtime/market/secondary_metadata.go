package market

// SecondaryContextDefaults carries the primary-context values used as a
// last-resort fallback when both the secondary exchange and its fixture-level
// metadata carry no timezone or reference session of their own.
type SecondaryContextDefaults struct {
	Timezone         string
	ReferenceSession string
}

// CompleteSecondaryMetadata returns a copy of raw with any missing Timezone
// or ReferenceSession filled in.
//
// Resolution priority (highest wins):
//  1. Value explicitly present in raw — honoured unchanged.
//  2. Exchange default derived from symbol or raw.Exchange when the exchange
//     is recognisable (MOEX, NYSE, Binance, …).
//  3. Corresponding field from defaults — only when the exchange is
//     ExchangeUnknown, covering same-symbol and opaque-fixture cases.
func CompleteSecondaryMetadata(symbol string, raw SourceMetadata, defaults SecondaryContextDefaults) SourceMetadata {
	exchange := ResolveExchangeWithMetadata(symbol, raw.Exchange)
	result := raw
	result.Timezone = secondaryTimezone(raw.Timezone, exchange, defaults.Timezone)
	result.ReferenceSession = secondaryReferenceSession(raw.ReferenceSession, exchange, defaults.ReferenceSession)
	return result
}

func secondaryTimezone(declared string, exchange Exchange, primaryFallback string) string {
	if declared != "" {
		return declared
	}
	if exchange != ExchangeUnknown {
		return DefaultTimezone(exchange)
	}
	return primaryFallback
}

func secondaryReferenceSession(declared string, exchange Exchange, primaryFallback string) string {
	if declared != "" {
		return declared
	}
	if exchange != ExchangeUnknown {
		return string(DefaultReferenceSession(exchange))
	}
	return primaryFallback
}

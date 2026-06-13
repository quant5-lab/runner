package market

import "github.com/quant5-lab/runner/runtime/context"

func FilterBars(symbol, timeframe, timezone string, bars []context.OHLCV) []context.OHLCV {
	normalized, _ := NormalizeBars(symbol, timeframe, timezone, bars)
	return normalized
}

func NormalizeBars(symbol, timeframe, timezone string, bars []context.OHLCV) ([]context.OHLCV, string) {
	normalized, profile := NormalizeBarsWithMetadata(symbol, timeframe, SourceMetadata{Timezone: timezone}, bars)
	return normalized, profile.Timezone
}

func NormalizeBarsWithReferenceSession(symbol, timeframe, timezone, referenceSession string, bars []context.OHLCV) ([]context.OHLCV, string, ReferenceSession) {
	normalized, profile := NormalizeBarsWithMetadata(symbol, timeframe, SourceMetadata{Timezone: timezone, ReferenceSession: referenceSession}, bars)
	return normalized, profile.Timezone, profile.ReferenceSession
}

func NormalizeBarsWithMetadata(symbol, timeframe string, metadata SourceMetadata, bars []context.OHLCV) ([]context.OHLCV, Profile) {
	profile := ResolveProfileWithMetadata(symbol, metadata)
	normalized, _ := NormalizeBarsForProfile(profile, timeframe, bars)
	return normalized, profile
}

func NormalizeBarsWithMetadataE(symbol, timeframe string, metadata SourceMetadata, bars []context.OHLCV) ([]context.OHLCV, Profile, error) {
	profile, err := ResolveProfileWithMetadataE(symbol, metadata)
	if err != nil {
		return nil, Profile{}, err
	}
	normalized, _ := NormalizeBarsForProfile(profile, timeframe, bars)
	return normalized, profile, nil
}

func NormalizeBarsForProfile(profile Profile, timeframe string, bars []context.OHLCV) ([]context.OHLCV, ReferenceSession) {
	if profile.Calendar == nil || len(bars) == 0 {
		return bars, profile.ReferenceSession
	}

	kept := make([]context.OHLCV, 0, len(bars))
	for _, bar := range bars {
		if profile.Calendar.Accepts(bar, timeframe, profile.Timezone) {
			kept = append(kept, bar)
		}
	}
	return kept, profile.ReferenceSession
}

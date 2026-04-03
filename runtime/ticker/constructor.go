package ticker

import "fmt"

func New(prefix, symbol string, session SessionType, adjustment AdjustmentType, backAdjustment BackAdjustmentType, settlement SettlementType) string {
	return TickerID{
		Prefix:         prefix,
		Symbol:         symbol,
		Session:        session,
		Adjustment:     adjustment,
		BackAdjustment: backAdjustment,
		Settlement:     settlement,
	}.Encode()
}

func Modify(tickerid string, session SessionType, adjustment AdjustmentType, backAdjustment BackAdjustmentType, settlement SettlementType) string {
	id := DecodeTickerID(tickerid)
	if session != "" {
		id.Session = session
	}
	if adjustment != "" {
		id.Adjustment = adjustment
	}
	if backAdjustment != "" {
		id.BackAdjustment = backAdjustment
	}
	if settlement != "" {
		id.Settlement = settlement
	}
	return id.Encode()
}

func Inherit(fromTickerid, symbol string) string {
	source := DecodeTickerID(fromTickerid)
	return TickerID{
		Symbol:         symbol,
		Session:        source.Session,
		Adjustment:     source.Adjustment,
		BackAdjustment: source.BackAdjustment,
		Settlement:     source.Settlement,
	}.Encode()
}

func PointFigure(symbol, source, style string, param, reversal float64) string {
	return fmt.Sprintf("%s:%s:%s:%s:%.2f:%.2f", ModifierPointFig, symbol, source, style, param, reversal)
}

func Standard(tickerid string) string {
	return ExtractBaseSymbol(tickerid)
}

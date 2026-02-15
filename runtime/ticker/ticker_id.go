package ticker

import "strings"

const modifierSeparator = "|"

type TickerID struct {
	Prefix         string
	Symbol         string
	Session        SessionType
	Adjustment     AdjustmentType
	BackAdjustment BackAdjustmentType
	Settlement     SettlementType
}

func (t TickerID) Encode() string {
	base := t.Symbol
	if t.Prefix != "" {
		base = t.Prefix + ":" + t.Symbol
	}

	var modifiers []string
	if t.Session != "" {
		modifiers = append(modifiers, "s="+string(t.Session))
	}
	if t.Adjustment != "" {
		modifiers = append(modifiers, "a="+string(t.Adjustment))
	}
	if t.BackAdjustment != "" {
		modifiers = append(modifiers, "ba="+string(t.BackAdjustment))
	}
	if t.Settlement != "" {
		modifiers = append(modifiers, "sc="+string(t.Settlement))
	}

	if len(modifiers) == 0 {
		return base
	}
	return base + modifierSeparator + strings.Join(modifiers, modifierSeparator)
}

func DecodeTickerID(encoded string) TickerID {
	base, modifiers := splitModifiers(encoded)

	var prefix, symbol string
	parts := strings.SplitN(base, ":", 2)
	if len(parts) == 2 {
		prefix = parts[0]
		symbol = parts[1]
	} else {
		symbol = base
	}

	id := TickerID{Prefix: prefix, Symbol: symbol}
	for _, mod := range modifiers {
		kv := strings.SplitN(mod, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "s":
			id.Session = SessionType(kv[1])
		case "a":
			id.Adjustment = AdjustmentType(kv[1])
		case "ba":
			id.BackAdjustment = BackAdjustmentType(kv[1])
		case "sc":
			id.Settlement = SettlementType(kv[1])
		}
	}
	return id
}

func StripModifiers(encoded string) string {
	base, _ := splitModifiers(encoded)
	return base
}

func splitModifiers(encoded string) (base string, modifiers []string) {
	parts := strings.Split(encoded, modifierSeparator)
	return parts[0], parts[1:]
}

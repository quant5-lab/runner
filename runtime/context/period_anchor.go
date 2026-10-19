package context

// PeriodAnchor describes how period boundaries are tiled for a symbol.
//
// Intraday tiling (e.g. 4h) originates from SessionOpenMinute within each
// calendar day in Timezone, rather than from UTC midnight.  Calendar-scale
// tiling (D, W, M) uses midnight of the given Timezone rather than UTC midnight.
//
// The zero value (empty Timezone, zero SessionOpenMinute) means UTC with a
// session that opens at midnight — identical to pre-anchor UTC arithmetic.
type PeriodAnchor struct {
	// IANA timezone name, e.g. "Europe/Moscow". Empty string means UTC.
	Timezone string

	// Minutes since local midnight when the exchange session opens.
	// 0 means the session originates at midnight (no offset).
	// Example: MOEX regular session opens at 07:00 MSK → 420.
	SessionOpenMinute int
}

package market

import (
	"strings"
	"time"
)

type Exchange string

const (
	ExchangeUnknown Exchange = ""
	ExchangeMOEX    Exchange = "MOEX"
	ExchangeBinance Exchange = "BINANCE"
)

type ReferenceSession string

const (
	ReferenceSessionAlwaysOpen ReferenceSession = "always-open"
	ReferenceSessionRegular    ReferenceSession = "regular"
)

type SourceMetadata struct {
	Exchange           string   `json:"exchange,omitempty"`
	Timezone           string   `json:"timezone,omitempty"`
	ReferenceSession   string   `json:"referenceSession,omitempty"`
	SessionSource      string   `json:"sessionSource,omitempty"`
	CalendarID         string   `json:"calendarId,omitempty"`
	ClosedWeekdays     []string `json:"closedWeekdays,omitempty"`
	OpenDates          []string `json:"openDates,omitempty"`
	ClosedDates        []string `json:"closedDates,omitempty"`
	IncludedDates      []string `json:"includedDates,omitempty"`
	IncludedTimestamps []int64  `json:"includedTimestamps,omitempty"`
}

type Profile struct {
	Exchange         Exchange
	Timezone         string
	Calendar         Calendar
	ReferenceSession ReferenceSession
	SessionSource    string
	CalendarID       string
}

func ResolveProfile(symbol, timezone string) Profile {
	return ResolveProfileWithMetadata(symbol, SourceMetadata{Timezone: timezone})
}

func ResolveProfileWithReferenceSession(symbol, timezone, referenceSession string) Profile {
	return ResolveProfileWithMetadata(symbol, SourceMetadata{Timezone: timezone, ReferenceSession: referenceSession})
}

func ResolveProfileWithMetadata(symbol string, metadata SourceMetadata) Profile {
	exchange := ResolveExchangeWithMetadata(symbol, metadata.Exchange)
	timezone := strings.TrimSpace(metadata.Timezone)
	if timezone == "" {
		timezone = DefaultTimezone(exchange)
	}
	session := ParseReferenceSession(metadata.ReferenceSession)

	return Profile{
		Exchange:         exchange,
		Timezone:         timezone,
		Calendar:         CalendarFor(exchange, session, metadata),
		ReferenceSession: session,
		SessionSource:    strings.TrimSpace(metadata.SessionSource),
		CalendarID:       strings.TrimSpace(metadata.CalendarID),
	}
}

func ResolveExchange(symbol string) Exchange {
	return ResolveExchangeWithMetadata(symbol, "")
}

func ResolveExchangeWithMetadata(symbol, exchange string) Exchange {
	if parsed := ParseExchange(exchange); parsed != ExchangeUnknown {
		return parsed
	}

	name := strings.ToUpper(strings.TrimSpace(symbol))
	name = strings.TrimPrefix(name, "MOEX:")
	name = strings.TrimPrefix(name, "BINANCE:")

	if strings.Contains(strings.ToUpper(symbol), "BINANCE:") || strings.HasSuffix(name, "USDT") {
		return ExchangeBinance
	}

	switch name {
	case "SBERP", "SBER", "GAZP", "LKOH", "YNDX", "CNRU":
		return ExchangeMOEX
	default:
		return ExchangeUnknown
	}
}

func ParseExchange(value string) Exchange {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "MOEX", "MISX":
		return ExchangeMOEX
	case "BINANCE":
		return ExchangeBinance
	default:
		return ExchangeUnknown
	}
}

func DefaultTimezone(exchange Exchange) string {
	switch exchange {
	case ExchangeMOEX:
		return "Europe/Moscow"
	default:
		return "UTC"
	}
}

func ParseReferenceSession(value string) ReferenceSession {
	switch ReferenceSession(strings.ToLower(strings.TrimSpace(value))) {
	case ReferenceSessionRegular:
		return ReferenceSessionRegular
	default:
		return ReferenceSessionAlwaysOpen
	}
}

func CalendarFor(exchange Exchange, referenceSession ReferenceSession, metadata SourceMetadata) Calendar {
	if exact := NewTimestampSet(metadata.IncludedTimestamps...); !exact.Empty() {
		return ExactTimestampCalendar{Allowed: exact}
	}
	if dates := NewDateSet(metadata.IncludedDates...); !dates.Empty() {
		return DateWhitelistCalendar{Allowed: dates}
	}
	if referenceSession != ReferenceSessionRegular {
		return AlwaysOpenCalendar{}
	}
	return RegularCalendarForExchange(exchange, metadata)
}

func RegularCalendarForExchange(exchange Exchange, metadata SourceMetadata) Calendar {
	closed := regularClosedWeekdays(exchange)
	if override := WeekdaySetFromNames(metadata.ClosedWeekdays...); !override.Empty() {
		closed = override
	}
	openDates := NewDateSet(metadata.OpenDates...)
	closedDates := NewDateSet(metadata.ClosedDates...)
	if closed.Empty() && openDates.Empty() && closedDates.Empty() {
		return AlwaysOpenCalendar{}
	}
	return NewRegularWeekdayCalendar(closed, openDates, closedDates)
}

func regularClosedWeekdays(exchange Exchange) WeekdaySet {
	switch exchange {
	case ExchangeMOEX:
		return NewWeekdaySet(time.Saturday, time.Sunday)
	default:
		return WeekdaySet{}
	}
}

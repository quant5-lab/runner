package market

import (
	"fmt"
	"strings"
	"time"
)

type Exchange string

const (
	ExchangeUnknown Exchange = ""
	ExchangeMOEX    Exchange = "MOEX"
	ExchangeBinance Exchange = "BINANCE"
	ExchangeNYSE    Exchange = "NYSE"
)

type ReferenceSession string

const (
	ReferenceSessionAlwaysOpen ReferenceSession = "always-open"
	ReferenceSessionRegular    ReferenceSession = "regular"
)

type SourceMetadata struct {
	Exchange             string            `json:"exchange,omitempty"`
	Timezone             string            `json:"timezone,omitempty"`
	ReferenceSession     string            `json:"referenceSession,omitempty"`
	SessionSource        string            `json:"sessionSource,omitempty"`
	CalendarID           string            `json:"calendarId,omitempty"`
	SessionWindow        string            `json:"sessionWindow,omitempty"`
	WeekdaySessionWindow string            `json:"weekdaySessionWindow,omitempty"`
	WeekendSessionWindow string            `json:"weekendSessionWindow,omitempty"`
	DateSessionWindows   map[string]string `json:"dateSessionWindows,omitempty"`
	ClosedWeekdays       []string          `json:"closedWeekdays,omitempty"`
	OpenDates            []string          `json:"openDates,omitempty"`
	ClosedDates          []string          `json:"closedDates,omitempty"`
	IncludedDates        []string          `json:"includedDates,omitempty"`
	IncludedTimestamps   []int64           `json:"includedTimestamps,omitempty"`
	QtyStep              float64           `json:"qtyStep,omitempty"`
}

type Profile struct {
	Exchange         Exchange
	Timezone         string
	Calendar         Calendar
	ReferenceSession ReferenceSession
	SessionSource    string
	CalendarID       string
	QtyStep          float64
	// SessionOpenMinute is minutes since local midnight when the exchange primary
	// session opens.  Zero means midnight — correct for UTC-aligned and always-open markets.
	SessionOpenMinute int
}

func ResolveProfile(symbol, timezone string) Profile {
	return ResolveProfileWithMetadata(symbol, SourceMetadata{Timezone: timezone})
}

func ResolveProfileWithReferenceSession(symbol, timezone, referenceSession string) Profile {
	return ResolveProfileWithMetadata(symbol, SourceMetadata{Timezone: timezone, ReferenceSession: referenceSession})
}

func ResolveProfileWithMetadata(symbol string, metadata SourceMetadata) Profile {
	profile, err := ResolveProfileWithMetadataE(symbol, metadata)
	if err == nil {
		return profile
	}
	exchange := ResolveExchangeWithMetadata(symbol, metadata.Exchange)
	timezone := strings.TrimSpace(metadata.Timezone)
	if timezone == "" {
		timezone = DefaultTimezone(exchange)
	}
	session := resolveReferenceSession(metadata, exchange)
	return Profile{
		Exchange:          exchange,
		Timezone:          timezone,
		Calendar:          CalendarFor(exchange, session, metadata),
		ReferenceSession:  session,
		SessionSource:     strings.TrimSpace(metadata.SessionSource),
		CalendarID:        strings.TrimSpace(metadata.CalendarID),
		QtyStep:           ResolveQtyStep(exchange, metadata),
		SessionOpenMinute: sessionOpenMinuteFor(exchange, session),
	}
}

func ResolveProfileWithMetadataE(symbol string, metadata SourceMetadata) (Profile, error) {
	exchange := ResolveExchangeWithMetadata(symbol, metadata.Exchange)
	timezone := strings.TrimSpace(metadata.Timezone)
	if timezone == "" {
		timezone = DefaultTimezone(exchange)
	}
	session := resolveReferenceSession(metadata, exchange)

	calendar, err := CalendarForE(exchange, session, metadata)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		Exchange:          exchange,
		Timezone:          timezone,
		Calendar:          calendar,
		ReferenceSession:  session,
		SessionSource:     strings.TrimSpace(metadata.SessionSource),
		CalendarID:        strings.TrimSpace(metadata.CalendarID),
		QtyStep:           ResolveQtyStep(exchange, metadata),
		SessionOpenMinute: sessionOpenMinuteFor(exchange, session),
	}, nil
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
	case "NYSE", "NASDAQ", "AMEX", "XNAS", "XNYS":
		return ExchangeNYSE
	default:
		return ExchangeUnknown
	}
}

func DefaultTimezone(exchange Exchange) string {
	switch exchange {
	case ExchangeMOEX:
		return "Europe/Moscow"
	case ExchangeNYSE:
		return "America/New_York"
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

func DefaultReferenceSession(exchange Exchange) ReferenceSession {
	switch exchange {
	case ExchangeMOEX, ExchangeNYSE:
		return ReferenceSessionRegular
	default:
		return ReferenceSessionAlwaysOpen
	}
}

func resolveReferenceSession(metadata SourceMetadata, exchange Exchange) ReferenceSession {
	if metadata.ReferenceSession != "" {
		return ParseReferenceSession(metadata.ReferenceSession)
	}
	return DefaultReferenceSession(exchange)
}

func CalendarFor(exchange Exchange, referenceSession ReferenceSession, metadata SourceMetadata) Calendar {
	calendar, err := CalendarForE(exchange, referenceSession, metadata)
	if err == nil {
		return calendar
	}
	return RegularCalendarForExchange(exchange, metadata)
}

func CalendarForE(exchange Exchange, referenceSession ReferenceSession, metadata SourceMetadata) (Calendar, error) {
	if exact := NewTimestampSet(metadata.IncludedTimestamps...); !exact.Empty() {
		return ExactTimestampCalendar{Allowed: exact}, nil
	}
	if dates := NewDateSet(metadata.IncludedDates...); !dates.Empty() {
		return DateWhitelistCalendar{Allowed: dates}, nil
	}
	if referenceSession != ReferenceSessionRegular {
		return AlwaysOpenCalendar{}, nil
	}
	return RegularCalendarForExchangeE(exchange, metadata)
}

func RegularCalendarForExchange(exchange Exchange, metadata SourceMetadata) Calendar {
	calendar, err := RegularCalendarForExchangeE(exchange, metadata)
	if err == nil {
		return calendar
	}
	closed := regularClosedWeekdays(exchange)
	if override := WeekdaySetFromNames(metadata.ClosedWeekdays...); !override.Empty() {
		closed = override
	}
	return NewRegularSessionCalendarWithSchedule(closed, NewDateSet(metadata.OpenDates...), NewDateSet(metadata.ClosedDates...), regularWeekSchedule(exchange))
}

func RegularCalendarForExchangeE(exchange Exchange, metadata SourceMetadata) (Calendar, error) {
	closed := regularClosedWeekdays(exchange)
	if override := WeekdaySetFromNames(metadata.ClosedWeekdays...); !override.Empty() {
		closed = override
	}
	openDates := NewDateSet(metadata.OpenDates...)
	closedDates := NewDateSet(metadata.ClosedDates...)
	schedule, err := resolveWeekSchedule(exchange, metadata)
	if err != nil {
		return nil, err
	}
	dateWindows, err := NewDateSessionWindows(metadata.DateSessionWindows)
	if err != nil {
		return nil, err
	}
	if closed.Empty() && openDates.Empty() && closedDates.Empty() && schedule.IsUnbounded() && dateWindows.Empty() {
		return AlwaysOpenCalendar{}, nil
	}
	return NewRegularSessionCalendar(closed, openDates, closedDates, schedule, dateWindows), nil
}

func regularClosedWeekdays(exchange Exchange) WeekdaySet {
	switch exchange {
	case ExchangeNYSE:
		return NewWeekdaySet(time.Saturday, time.Sunday)
	default:
		return WeekdaySet{}
	}
}

// regularWeekSchedule returns the canonical weekday and weekend session windows
// for exchanges with known fixed schedules, as observed in TradingView reference data.
// All session times are in the exchange's local timezone (see DefaultTimezone).
func regularWeekSchedule(exchange Exchange) WeekSchedule {
	switch exchange {
	case ExchangeMOEX:
		// Weekday [07:00, 23:50) MSK, weekend [10:00, 19:00) MSK — TV bar universe.
		return WeekSchedule{
			Weekday: NewSessionWindow(ClockTimeAt(7, 0), ClockTimeAt(23, 50)),
			Weekend: NewSessionWindow(ClockTimeAt(10, 0), ClockTimeAt(19, 0)),
		}
	case ExchangeNYSE:
		// Regular market session [09:30, 16:00) ET, closed weekends.
		return WeekSchedule{
			Weekday: NewSessionWindow(ClockTimeAt(9, 30), ClockTimeAt(16, 0)),
		}
	default:
		return WeekSchedule{}
	}
}

func resolveWeekSchedule(exchange Exchange, metadata SourceMetadata) (WeekSchedule, error) {
	schedule, ok, err := weekScheduleFromMetadata(metadata)
	if err != nil {
		return WeekSchedule{}, err
	}
	if ok {
		return schedule, nil
	}
	return regularWeekSchedule(exchange), nil
}

func weekScheduleFromMetadata(metadata SourceMetadata) (WeekSchedule, bool, error) {
	if metadata.SessionWindow != "" {
		parsed, err := ParseSessionWindow(metadata.SessionWindow)
		if err != nil {
			return WeekSchedule{}, false, fmt.Errorf("sessionWindow: %w", err)
		}
		return UniformWeekSchedule(parsed), true, nil
	}

	var schedule WeekSchedule
	var hasOverride bool
	if metadata.WeekdaySessionWindow != "" {
		parsed, err := ParseSessionWindow(metadata.WeekdaySessionWindow)
		if err != nil {
			return WeekSchedule{}, false, fmt.Errorf("weekdaySessionWindow: %w", err)
		}
		schedule.Weekday = parsed
		hasOverride = true
	}
	if metadata.WeekendSessionWindow != "" {
		parsed, err := ParseSessionWindow(metadata.WeekendSessionWindow)
		if err != nil {
			return WeekSchedule{}, false, fmt.Errorf("weekendSessionWindow: %w", err)
		}
		schedule.Weekend = parsed
		hasOverride = true
	}
	return schedule, hasOverride, nil
}

// sessionOpenMinuteFor returns the exchange session's weekday open as minutes
// since local midnight.  When the session is always-open or the exchange has no
// known schedule, zero is returned (midnight origin — no tiling offset).
func sessionOpenMinuteFor(exchange Exchange, session ReferenceSession) int {
	if session != ReferenceSessionRegular {
		return 0
	}
	return regularWeekSchedule(exchange).Weekday.Start.TotalMinutes()
}

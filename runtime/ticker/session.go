package ticker

type SessionType string

const (
	SessionRegular  SessionType = "regular"
	SessionExtended SessionType = "extended"
)

type AdjustmentType string

const (
	AdjustmentNone      AdjustmentType = "none"
	AdjustmentSplits    AdjustmentType = "splits"
	AdjustmentDividends AdjustmentType = "dividends"
)

type BackAdjustmentType string

const (
	BackAdjustmentInherit BackAdjustmentType = "inherit"
	BackAdjustmentOn      BackAdjustmentType = "on"
	BackAdjustmentOff     BackAdjustmentType = "off"
)

type SettlementType string

const (
	SettlementInherit SettlementType = "inherit"
	SettlementOn      SettlementType = "on"
	SettlementOff     SettlementType = "off"
)

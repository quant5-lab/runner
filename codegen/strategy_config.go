package codegen

const (
	defaultInitialCapital  = 10000.0
	defaultQtyValue        = 1.0
	defaultPyramiding      = 0
	defaultCommissionValue = 0.0
)

type StrategyConfig struct {
	Name                 string
	InitialCapital       float64
	DefaultQtyValue      float64
	DefaultQtyType       string
	Pyramiding           int
	CommissionType       string
	CommissionValue      float64
	ProcessOrdersOnClose bool
}

func NewStrategyConfig() *StrategyConfig {
	return &StrategyConfig{
		Name:            "Generated Strategy",
		InitialCapital:  defaultInitialCapital,
		DefaultQtyValue: defaultQtyValue,
		Pyramiding:      defaultPyramiding,
		CommissionValue: defaultCommissionValue,
	}
}

func (c *StrategyConfig) MergeFrom(other *StrategyConfig) {
	if other == nil {
		return
	}
	if other.Name != "" && other.Name != "Generated Strategy" {
		c.Name = other.Name
	}
	if other.InitialCapital > 0 {
		c.InitialCapital = other.InitialCapital
	}
	if other.DefaultQtyValue > 0 {
		c.DefaultQtyValue = other.DefaultQtyValue
	}
	if other.DefaultQtyType != "" {
		c.DefaultQtyType = other.DefaultQtyType
	}
	if other.Pyramiding >= 0 {
		c.Pyramiding = other.Pyramiding
	}
	if other.CommissionType != "" {
		c.CommissionType = other.CommissionType
	}
	if other.CommissionValue > 0 {
		c.CommissionValue = other.CommissionValue
	}
	if other.ProcessOrdersOnClose {
		c.ProcessOrdersOnClose = true
	}
}

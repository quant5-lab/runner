package codegen

type SeriesLifecycle interface {
	HasUsage() bool
	NeedsTimezone() bool
	GenerateDeclarations(indent string) string
	GenerateInitializations(indent string) string
	GenerateBarPopulation(indent, iterVar string) string
	GenerateAdvancement(indent, iterVar string) string
	GenerateRegistrations(indent string) string
	GenerateSuppressUnused(indent string) string
	GenerateSymbolTableRegistrations(symbolTable SymbolTable)
}

type CompositeSeriesLifecycle struct {
	children []SeriesLifecycle
}

func NewCompositeSeriesLifecycle(children ...SeriesLifecycle) *CompositeSeriesLifecycle {
	return &CompositeSeriesLifecycle{children: children}
}

func (c *CompositeSeriesLifecycle) HasUsage() bool {
	if c == nil {
		return false
	}
	for _, child := range c.children {
		if child.HasUsage() {
			return true
		}
	}
	return false
}

func (c *CompositeSeriesLifecycle) NeedsTimezone() bool {
	if c == nil {
		return false
	}
	for _, child := range c.children {
		if child.NeedsTimezone() {
			return true
		}
	}
	return false
}

func (c *CompositeSeriesLifecycle) GenerateDeclarations(indent string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateDeclarations(indent)
	})
}

func (c *CompositeSeriesLifecycle) GenerateInitializations(indent string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateInitializations(indent)
	})
}

func (c *CompositeSeriesLifecycle) GenerateBarPopulation(indent, iterVar string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateBarPopulation(indent, iterVar)
	})
}

func (c *CompositeSeriesLifecycle) GenerateAdvancement(indent, iterVar string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateAdvancement(indent, iterVar)
	})
}

func (c *CompositeSeriesLifecycle) GenerateRegistrations(indent string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateRegistrations(indent)
	})
}

func (c *CompositeSeriesLifecycle) GenerateSuppressUnused(indent string) string {
	return c.collect(func(child SeriesLifecycle) string {
		return child.GenerateSuppressUnused(indent)
	})
}

func (c *CompositeSeriesLifecycle) GenerateSymbolTableRegistrations(symbolTable SymbolTable) {
	if c == nil {
		return
	}
	for _, child := range c.children {
		child.GenerateSymbolTableRegistrations(symbolTable)
	}
}

func (c *CompositeSeriesLifecycle) GenerateTimezoneSetup(indent string) string {
	if !c.NeedsTimezone() {
		return ""
	}
	return indent + "exchangeLoc, err := time.LoadLocation(ctx.Timezone)\n" +
		indent + `if err != nil { panic("invalid timezone: " + ctx.Timezone) }` + "\n"
}

func (c *CompositeSeriesLifecycle) collect(fn func(SeriesLifecycle) string) string {
	if c == nil {
		return ""
	}
	code := ""
	for _, child := range c.children {
		code += fn(child)
	}
	return code
}

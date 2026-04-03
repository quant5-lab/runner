package output

/* PlotPoint represents a single plot data point */
type PlotPoint struct {
	Time    int64
	Value   float64
	Options map[string]interface{}
}

/* PlotSeries represents a named series of plot points */
type PlotSeries struct {
	Title string
	Data  []PlotPoint
}

/* PlotCollector interface for collecting plot data */
type PlotCollector interface {
	Add(title string, time int64, value float64, options map[string]interface{})
	GetSeries() []PlotSeries
}

/* Collector implements PlotCollector */
type Collector struct {
	series map[string]*PlotSeries
}

/* NewCollector creates a new plot collector */
func NewCollector() *Collector {
	return &Collector{
		series: make(map[string]*PlotSeries),
	}
}

/* Add adds a plot point to the named series */
func (c *Collector) Add(title string, time int64, value float64, options map[string]interface{}) {
	if _, exists := c.series[title]; !exists {
		c.series[title] = &PlotSeries{
			Title: title,
			Data:  make([]PlotPoint, 0),
		}
	}

	c.series[title].Data = append(c.series[title].Data, PlotPoint{
		Time:    time,
		Value:   value,
		Options: options,
	})
}

/* GetSeries returns all plot series */
func (c *Collector) GetSeries() []PlotSeries {
	result := make([]PlotSeries, 0, len(c.series))
	for _, series := range c.series {
		result = append(result, *series)
	}
	return result
}

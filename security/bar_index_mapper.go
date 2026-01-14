package security

type BarRange struct {
	DailyBarIndex    int
	StartHourlyIndex int
	EndHourlyIndex   int
}

type SecurityBarMapperInterface interface {
	FindDailyBarIndex(hourlyIndex int, lookahead bool) int
	GetRanges() []BarRange
}

type BarIndexMapper struct {
	secToMainMap map[int]int
}

func NewBarIndexMapper() *BarIndexMapper {
	return &BarIndexMapper{
		secToMainMap: make(map[int]int),
	}
}

func (m *BarIndexMapper) SetMapping(secBarIdx, mainBarIdx int) {
	m.secToMainMap[secBarIdx] = mainBarIdx
}

func (m *BarIndexMapper) GetMainBarIndexForSecurityBar(secBarIdx int) int {
	if mainIdx, ok := m.secToMainMap[secBarIdx]; ok {
		return mainIdx
	}
	return -1
}

func (m *BarIndexMapper) PopulateFromSecurityMapper(
	secMapper SecurityBarMapperInterface,
	mainBarCount int,
) {
	ranges := secMapper.GetRanges()
	for _, r := range ranges {
		if r.StartHourlyIndex >= 0 {
			m.SetMapping(r.DailyBarIndex, r.StartHourlyIndex)
		}
	}
}

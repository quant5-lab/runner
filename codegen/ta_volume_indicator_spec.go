package codegen

import (
	"fmt"
	"strings"
)

// VolumeIndicatorSpec is the single source of truth for one ta.* volume variable:
// its detection key, generated series name, and bar-level code generation.
type VolumeIndicatorSpec struct {
	MemberKey  string // "ta.obv"
	SeriesName string // "ta_obvSeries"
	populateFn func(seriesName, indent, iterVar string) string
}

func (s *VolumeIndicatorSpec) PropName() string {
	return strings.TrimPrefix(s.MemberKey, "ta.")
}

func (s *VolumeIndicatorSpec) PopulateBarCode(indent, iterVar string) string {
	return s.populateFn(s.SeriesName, indent, iterVar)
}

// volumeIndicatorByPropName is the O(1) lookup table built once at init time.
var volumeIndicatorByPropName = buildVolumeIndicatorIndex()

func buildVolumeIndicatorIndex() map[string]*VolumeIndicatorSpec {
	m := make(map[string]*VolumeIndicatorSpec, len(allVolumeIndicatorSpecs))
	for _, spec := range allVolumeIndicatorSpecs {
		m[spec.PropName()] = spec
	}
	return m
}

// LookupVolumeIndicator returns the spec for a ta.* property name (e.g. "obv").
func LookupVolumeIndicator(propName string) (*VolumeIndicatorSpec, bool) {
	spec, ok := volumeIndicatorByPropName[propName]
	return spec, ok
}

// VolumeIndicatorMemberKeys returns all "ta.X" detection keys for usage scanning.
func VolumeIndicatorMemberKeys() []string {
	keys := make([]string, len(allVolumeIndicatorSpecs))
	for i, spec := range allVolumeIndicatorSpecs {
		keys[i] = spec.MemberKey
	}
	return keys
}

// allVolumeIndicatorSpecs defines the Pine ta.* volume built-in variables.
var allVolumeIndicatorSpecs = []*VolumeIndicatorSpec{
	{MemberKey: "ta.obv", SeriesName: "ta_obvSeries", populateFn: genOBVPopulate},
	{MemberKey: "ta.accdist", SeriesName: "ta_accdistSeries", populateFn: genAccdistPopulate},
	{MemberKey: "ta.pvt", SeriesName: "ta_pvtSeries", populateFn: genPVTPopulate},
	{MemberKey: "ta.iii", SeriesName: "ta_iiiSeries", populateFn: genIIIPopulate},
	{MemberKey: "ta.wvad", SeriesName: "ta_wvadSeries", populateFn: genWVADPopulate},
	{MemberKey: "ta.nvi", SeriesName: "ta_nviSeries", populateFn: genNVIPopulate},
	{MemberKey: "ta.pvi", SeriesName: "ta_pviSeries", populateFn: genPVIPopulate},
	{MemberKey: "ta.wad", SeriesName: "ta_wadSeries", populateFn: genWADPopulate},
}

// genOBVPopulate — On Balance Volume:
//
//	obv = if close > close[1]: obv[1] + volume
//	      if close < close[1]: obv[1] - volume
//	      else:                obv[1]
func genOBVPopulate(sn, indent, iv string) string {
	t, tt, ttt := tabs(indent, 1), tabs(indent, 2), tabs(indent, 3)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(fmt.Sprintf(t+"if %s == 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"%s.Set(0.0)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"_obv_prev := %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_obv_prev) { _obv_prev = 0.0 }\n")
	b.WriteString(fmt.Sprintf(tt+"_obv_prevClose := ctx.Data[%s-1].Close\n", iv))
	b.WriteString(tt + "if bar.Close > _obv_prevClose {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_obv_prev + bar.Volume)\n", sn))
	b.WriteString(tt + "} else if bar.Close < _obv_prevClose {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_obv_prev - bar.Volume)\n", sn))
	b.WriteString(tt + "} else {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_obv_prev)\n", sn))
	b.WriteString(tt + "}\n")
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genAccdistPopulate — Accumulation/Distribution:
//
//	clv = ((close - low) - (high - close)) / (high - low)
//	accdist = cum(clv * volume)
func genAccdistPopulate(sn, indent, iv string) string {
	t, tt := tabs(indent, 1), tabs(indent, 2)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(t + "_accdist_prev := 0.0\n")
	b.WriteString(fmt.Sprintf(t+"if %s > 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"_accdist_prev = %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_accdist_prev) { _accdist_prev = 0.0 }\n")
	b.WriteString(t + "}\n")
	b.WriteString(t + "_accdist_hl := bar.High - bar.Low\n")
	b.WriteString(t + "if _accdist_hl == 0 {\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_accdist_prev)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(tt + "_accdist_clv := ((bar.Close - bar.Low) - (bar.High - bar.Close)) / _accdist_hl\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_accdist_prev + _accdist_clv*bar.Volume)\n", sn))
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genPVTPopulate — Price Volume Trend:
//
//	pvt = cum((change(close) / close[1]) * volume)
func genPVTPopulate(sn, indent, iv string) string {
	t, tt, ttt := tabs(indent, 1), tabs(indent, 2), tabs(indent, 3)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(t + "_pvt_prev := 0.0\n")
	b.WriteString(fmt.Sprintf(t+"if %s == 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"%s.Set(0.0)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"_pvt_prev = %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_pvt_prev) { _pvt_prev = 0.0 }\n")
	b.WriteString(fmt.Sprintf(tt+"_pvt_prevClose := ctx.Data[%s-1].Close\n", iv))
	b.WriteString(tt + "if _pvt_prevClose == 0 {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_pvt_prev)\n", sn))
	b.WriteString(tt + "} else {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_pvt_prev + (bar.Close-_pvt_prevClose)/_pvt_prevClose*bar.Volume)\n", sn))
	b.WriteString(tt + "}\n")
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genIIIPopulate — Intraday Intensity Index:
//
//	iii = cum((2*close - high - low) / ((high - low) * volume))
func genIIIPopulate(sn, indent, iv string) string {
	t, tt := tabs(indent, 1), tabs(indent, 2)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(t + "_iii_prev := 0.0\n")
	b.WriteString(fmt.Sprintf(t+"if %s > 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"_iii_prev = %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_iii_prev) { _iii_prev = 0.0 }\n")
	b.WriteString(t + "}\n")
	b.WriteString(t + "_iii_hl := bar.High - bar.Low\n")
	b.WriteString(t + "if _iii_hl == 0 || bar.Volume == 0 {\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_iii_prev)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_iii_prev + (2*bar.Close-bar.High-bar.Low)/(_iii_hl*bar.Volume))\n", sn))
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genWVADPopulate — Williams Variable Accumulation/Distribution:
//
//	wvad = cum((close - open) / (high - low) * volume)
func genWVADPopulate(sn, indent, iv string) string {
	t, tt := tabs(indent, 1), tabs(indent, 2)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(t + "_wvad_prev := 0.0\n")
	b.WriteString(fmt.Sprintf(t+"if %s > 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"_wvad_prev = %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_wvad_prev) { _wvad_prev = 0.0 }\n")
	b.WriteString(t + "}\n")
	b.WriteString(t + "_wvad_hl := bar.High - bar.Low\n")
	b.WriteString(t + "if _wvad_hl == 0 {\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_wvad_prev)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"%s.Set(_wvad_prev + (bar.Close-bar.Open)/_wvad_hl*bar.Volume)\n", sn))
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genNVIPopulate — Negative Volume Index (seed 1000):
//
//	nvi = if volume < volume[1]: nvi[1] * (1 + change(close)/close[1])
//	      else:                  nvi[1]
func genNVIPopulate(sn, indent, iv string) string {
	t, tt, ttt := tabs(indent, 1), tabs(indent, 2), tabs(indent, 3)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(fmt.Sprintf(t+"if %s == 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"%s.Set(1000.0)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"_nvi_prev := %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_nvi_prev) { _nvi_prev = 1000.0 }\n")
	b.WriteString(fmt.Sprintf(tt+"_nvi_prevBar := ctx.Data[%s-1]\n", iv))
	b.WriteString(tt + "if bar.Volume < _nvi_prevBar.Volume && _nvi_prevBar.Close != 0 {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_nvi_prev * (1.0 + (bar.Close-_nvi_prevBar.Close)/_nvi_prevBar.Close))\n", sn))
	b.WriteString(tt + "} else {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_nvi_prev)\n", sn))
	b.WriteString(tt + "}\n")
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genPVIPopulate — Positive Volume Index (seed 1000):
//
//	pvi = if volume > volume[1]: pvi[1] * (1 + change(close)/close[1])
//	      else:                  pvi[1]
func genPVIPopulate(sn, indent, iv string) string {
	t, tt, ttt := tabs(indent, 1), tabs(indent, 2), tabs(indent, 3)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(fmt.Sprintf(t+"if %s == 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"%s.Set(1000.0)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"_pvi_prev := %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_pvi_prev) { _pvi_prev = 1000.0 }\n")
	b.WriteString(fmt.Sprintf(tt+"_pvi_prevBar := ctx.Data[%s-1]\n", iv))
	b.WriteString(tt + "if bar.Volume > _pvi_prevBar.Volume && _pvi_prevBar.Close != 0 {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_pvi_prev * (1.0 + (bar.Close-_pvi_prevBar.Close)/_pvi_prevBar.Close))\n", sn))
	b.WriteString(tt + "} else {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_pvi_prev)\n", sn))
	b.WriteString(tt + "}\n")
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

// genWADPopulate — Williams Accumulation/Distribution:
//
//	trueHigh = max(high, close[1])
//	trueLow  = min(low,  close[1])
//	wad = if close > close[1]: wad[1] + close - trueLow
//	      if close < close[1]: wad[1] + close - trueHigh
//	      else:                wad[1]
func genWADPopulate(sn, indent, iv string) string {
	t, tt, ttt := tabs(indent, 1), tabs(indent, 2), tabs(indent, 3)
	var b strings.Builder
	b.WriteString(indent + "{\n")
	b.WriteString(fmt.Sprintf(t+"if %s == 0 {\n", iv))
	b.WriteString(fmt.Sprintf(tt+"%s.Set(0.0)\n", sn))
	b.WriteString(t + "} else {\n")
	b.WriteString(fmt.Sprintf(tt+"_wad_prev := %s.Get(1)\n", sn))
	b.WriteString(tt + "if math.IsNaN(_wad_prev) { _wad_prev = 0.0 }\n")
	b.WriteString(fmt.Sprintf(tt+"_wad_prevClose := ctx.Data[%s-1].Close\n", iv))
	b.WriteString(tt + "_wad_trueHigh := math.Max(bar.High, _wad_prevClose)\n")
	b.WriteString(tt + "_wad_trueLow := math.Min(bar.Low, _wad_prevClose)\n")
	b.WriteString(tt + "if bar.Close > _wad_prevClose {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_wad_prev + bar.Close - _wad_trueLow)\n", sn))
	b.WriteString(tt + "} else if bar.Close < _wad_prevClose {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_wad_prev + bar.Close - _wad_trueHigh)\n", sn))
	b.WriteString(tt + "} else {\n")
	b.WriteString(fmt.Sprintf(ttt+"%s.Set(_wad_prev)\n", sn))
	b.WriteString(tt + "}\n")
	b.WriteString(t + "}\n")
	b.WriteString(indent + "}\n")
	return b.String()
}

func tabs(base string, n int) string {
	return base + strings.Repeat("\t", n)
}

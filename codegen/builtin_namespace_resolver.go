package codegen

import "fmt"

type NamespaceResolution struct {
	Code   string
	GoType GoValueType
}

type BuiltinNamespaceResolver struct {
	dispatchers map[string]func(string) (NamespaceResolution, bool)
}

func NewBuiltinNamespaceResolver() *BuiltinNamespaceResolver {
	r := &BuiltinNamespaceResolver{}
	r.dispatchers = map[string]func(string) (NamespaceResolution, bool){
		"barstate":  r.resolveBarState,
		"timeframe": r.resolveTimeframe,
		"syminfo":   r.resolveSyminfo,
		"dayofweek": r.resolveDayOfWeek,
		"session":   r.resolveSession,
		"chart":     r.resolveChart,
		"dividends": r.resolveDividends,
		"earnings":  r.resolveEarnings,
		"math":      r.resolveMath,
		"strategy":  r.resolveStrategy,
		"extend":    r.resolveExtend,
		"line":      r.resolveLineStyle,
		"label":     r.resolveLabelStyle,
		"xloc":      r.resolveXloc,
		"size":      r.resolveSize,
		"shape":     r.resolveShape,
		"location":  r.resolveLocation,
		"hline":     r.resolveHlineStyle,
	}
	return r
}

func (r *BuiltinNamespaceResolver) Resolve(obj, prop string) (NamespaceResolution, bool) {
	if dispatcher, exists := r.dispatchers[obj]; exists {
		return dispatcher(prop)
	}
	return NamespaceResolution{}, false
}

func (r *BuiltinNamespaceResolver) IsNamespace(obj string) bool {
	_, exists := r.dispatchers[obj]
	return exists
}

/* Overrides builtins referencing outer-scope variables; delegates ctx-compatible cases to Resolve */
func (r *BuiltinNamespaceResolver) ResolveForArrow(obj, prop string) (NamespaceResolution, bool) {
	switch obj {
	case "syminfo":
		return r.resolveSyminfoForArrow(prop)
	case "session":
		return r.resolveSessionForArrow(prop)
	default:
		return r.Resolve(obj, prop)
	}
}

func (r *BuiltinNamespaceResolver) resolveSyminfoForArrow(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "tickerid", "ticker":
		return NamespaceResolution{Code: "ctx.Symbol", GoType: GoString}, true
	case "description":
		return NamespaceResolution{Code: "ctx.Symbol", GoType: GoString}, true
	default:
		return r.resolveSyminfo(prop)
	}
}

func (r *BuiltinNamespaceResolver) resolveSessionForArrow(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismarket":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "ispremarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "ispostmarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isfirstbar":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsFirstBarSeriesName), GoType: GoBool}, true
	case "islastbar":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsLastBarSeriesName), GoType: GoBool}, true
	case "isfirstbar_regular":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsFirstBarRegularSeriesName), GoType: GoBool}, true
	case "islastbar_regular":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsLastBarRegularSeriesName), GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func arrowSessionBoolLookup(seriesName string) string {
	return fmt.Sprintf(
		`func() bool { if s, ok := ctx.LookupSeries(%q); ok { return s.GetCurrent() == 1.0 }; return false }()`,
		seriesName)
}

func (r *BuiltinNamespaceResolver) resolveBarState(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "isfirst":
		return NamespaceResolution{Code: "(ctx.BarIndex == 0)", GoType: GoBool}, true
	case "islast":
		return NamespaceResolution{Code: "(ctx.BarIndex == len(ctx.Data)-1)", GoType: GoBool}, true
	case "ishistory":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "isrealtime":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isnew":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "isconfirmed":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "islastconfirmedhistory":
		return NamespaceResolution{Code: "(ctx.BarIndex == len(ctx.Data)-1)", GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveTimeframe(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismonthly":
		return NamespaceResolution{Code: "ctx.IsMonthly", GoType: GoBool}, true
	case "isdaily":
		return NamespaceResolution{Code: "ctx.IsDaily", GoType: GoBool}, true
	case "isweekly":
		return NamespaceResolution{Code: "ctx.IsWeekly", GoType: GoBool}, true
	case "isintraday":
		return NamespaceResolution{Code: "ctx.IsIntraday", GoType: GoBool}, true
	case "isdwm":
		return NamespaceResolution{Code: "(ctx.IsDaily || ctx.IsWeekly || ctx.IsMonthly)", GoType: GoBool}, true
	case "isminutes":
		return NamespaceResolution{Code: "ctx.IsIntraday", GoType: GoBool}, true
	case "isseconds":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isticks":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "multiplier":
		return NamespaceResolution{Code: "float64(context.TimeframeMultiplier(ctx.Timeframe))"}, true
	case "period":
		return NamespaceResolution{Code: "ctx.Timeframe", GoType: GoString}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSyminfo(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "tickerid", "ticker":
		return NamespaceResolution{Code: "syminfo_tickerid", GoType: GoString}, true
	case "timezone":
		return NamespaceResolution{Code: "ctx.Timezone", GoType: GoString}, true
	case "type":
		return NamespaceResolution{Code: `"stock"`, GoType: GoString}, true
	case "prefix":
		return NamespaceResolution{Code: `""`, GoType: GoString}, true
	case "session":
		return NamespaceResolution{Code: `"regular"`, GoType: GoString}, true
	case "currency":
		return NamespaceResolution{Code: `"USD"`, GoType: GoString}, true
	case "basecurrency":
		return NamespaceResolution{Code: `""`, GoType: GoString}, true
	case "description":
		return NamespaceResolution{Code: "syminfo_tickerid", GoType: GoString}, true
	case "pointvalue":
		return NamespaceResolution{Code: "1.0"}, true
	case "mintick":
		return NamespaceResolution{Code: "0.01"}, true
	case "volumetype":
		return NamespaceResolution{Code: `"base"`, GoType: GoString}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveDayOfWeek(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "sunday":
		return NamespaceResolution{Code: "1.0"}, true
	case "monday":
		return NamespaceResolution{Code: "2.0"}, true
	case "tuesday":
		return NamespaceResolution{Code: "3.0"}, true
	case "wednesday":
		return NamespaceResolution{Code: "4.0"}, true
	case "thursday":
		return NamespaceResolution{Code: "5.0"}, true
	case "friday":
		return NamespaceResolution{Code: "6.0"}, true
	case "saturday":
		return NamespaceResolution{Code: "7.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSession(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismarket":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "ispremarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "ispostmarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isfirstbar":
		return NamespaceResolution{Code: SessionIsFirstBarSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "islastbar":
		return NamespaceResolution{Code: SessionIsLastBarSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "isfirstbar_regular":
		return NamespaceResolution{Code: SessionIsFirstBarRegularSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "islastbar_regular":
		return NamespaceResolution{Code: SessionIsLastBarRegularSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveChart(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "is_standard":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "is_heikinashi":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_kagi":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_linebreak":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_pnf":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_range":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_renko":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "bg_color":
		return NamespaceResolution{Code: `"#FFFFFF"`, GoType: GoString}, true
	case "fg_color":
		return NamespaceResolution{Code: `"#000000"`, GoType: GoString}, true
	case "left_visible_bar_time":
		return NamespaceResolution{Code: "float64(ctx.Data[0].Time * 1000)"}, true
	case "right_visible_bar_time":
		return NamespaceResolution{Code: "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveDividends(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "future_amount", "future_ex_date", "future_pay_date":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveEarnings(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "future_eps", "future_period_end_time", "future_revenue", "future_time":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveMath(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "pi":
		return NamespaceResolution{Code: "math.Pi"}, true
	case "e":
		return NamespaceResolution{Code: "math.E"}, true
	case "phi":
		return NamespaceResolution{Code: "1.618033988749895"}, true
	case "rphi":
		return NamespaceResolution{Code: "0.618033988749895"}, true
	default:
		return NamespaceResolution{}, false
	}
}

/* resolveStrategy handles strategy.* flat constants; strategy.entry/close are call sites, not namespaced constants. */
func (r *BuiltinNamespaceResolver) resolveStrategy(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "cash":
		return NamespaceResolution{Code: `"cash"`, GoType: GoString}, true
	case "fixed":
		return NamespaceResolution{Code: `"fixed"`, GoType: GoString}, true
	case "percent_of_equity":
		return NamespaceResolution{Code: `"percent_of_equity"`, GoType: GoString}, true
	case "long":
		return NamespaceResolution{Code: `"long"`, GoType: GoString}, true
	case "short":
		return NamespaceResolution{Code: `"short"`, GoType: GoString}, true
	case "both":
		return NamespaceResolution{Code: `"both"`, GoType: GoString}, true
	case "account_currency":
		return NamespaceResolution{Code: `"USD"`, GoType: GoString}, true
	case "margin_liquidation_price":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveExtend(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "none":
		return NamespaceResolution{Code: "0.0"}, true
	case "right":
		return NamespaceResolution{Code: "1.0"}, true
	case "left":
		return NamespaceResolution{Code: "2.0"}, true
	case "both":
		return NamespaceResolution{Code: "3.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveLineStyle(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "style_solid":
		return NamespaceResolution{Code: "0.0"}, true
	case "style_dashed":
		return NamespaceResolution{Code: "1.0"}, true
	case "style_dotted":
		return NamespaceResolution{Code: "2.0"}, true
	case "style_arrow_left":
		return NamespaceResolution{Code: "3.0"}, true
	case "style_arrow_right":
		return NamespaceResolution{Code: "4.0"}, true
	case "style_arrow_both":
		return NamespaceResolution{Code: "5.0"}, true
	case "style_cross":
		return NamespaceResolution{Code: "6.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveLabelStyle(prop string) (NamespaceResolution, bool) {
	styles := map[string]float64{
		"style_none":              0,
		"style_xcross":            1,
		"style_cross":             2,
		"style_triangleup":        3,
		"style_triangledown":      4,
		"style_flag":              5,
		"style_circle":            6,
		"style_arrowup":           7,
		"style_arrowdown":         8,
		"style_label_up":          9,
		"style_label_down":        10,
		"style_label_left":        11,
		"style_label_right":       12,
		"style_label_lower_left":  13,
		"style_label_lower_right": 14,
		"style_label_upper_left":  15,
		"style_label_upper_right": 16,
		"style_label_center":      17,
		"style_square":            18,
		"style_diamond":           19,
	}
	if v, ok := styles[prop]; ok {
		return NamespaceResolution{Code: fmt.Sprintf("%.1f", v)}, true
	}
	return NamespaceResolution{}, false
}

func (r *BuiltinNamespaceResolver) resolveXloc(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "bar_time":
		return NamespaceResolution{Code: "0.0"}, true
	case "bar_index":
		return NamespaceResolution{Code: "1.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSize(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "auto":
		return NamespaceResolution{Code: "0.0"}, true
	case "tiny":
		return NamespaceResolution{Code: "1.0"}, true
	case "small":
		return NamespaceResolution{Code: "2.0"}, true
	case "normal":
		return NamespaceResolution{Code: "3.0"}, true
	case "large":
		return NamespaceResolution{Code: "4.0"}, true
	case "huge":
		return NamespaceResolution{Code: "5.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveShape(prop string) (NamespaceResolution, bool) {
	shapes := map[string]float64{
		"xcross":       0,
		"cross":        1,
		"circle":       2,
		"triangleup":   3,
		"triangledown": 4,
		"flag":         5,
		"labelup":      6,
		"labeldown":    7,
		"arrowup":      8,
		"arrowdown":    9,
		"diamond":      10,
		"square":       11,
	}
	if v, ok := shapes[prop]; ok {
		return NamespaceResolution{Code: fmt.Sprintf("%.1f", v)}, true
	}
	return NamespaceResolution{}, false
}

func (r *BuiltinNamespaceResolver) resolveLocation(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "abovebar":
		return NamespaceResolution{Code: "0.0"}, true
	case "belowbar":
		return NamespaceResolution{Code: "1.0"}, true
	case "top":
		return NamespaceResolution{Code: "2.0"}, true
	case "bottom":
		return NamespaceResolution{Code: "3.0"}, true
	case "absolute":
		return NamespaceResolution{Code: "4.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveHlineStyle(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "style_solid":
		return NamespaceResolution{Code: "0.0"}, true
	case "style_dashed":
		return NamespaceResolution{Code: "1.0"}, true
	case "style_dotted":
		return NamespaceResolution{Code: "2.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

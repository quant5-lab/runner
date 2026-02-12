package codegen

type NamespaceResolution struct {
	Code   string
	IsBool bool
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

func (r *BuiltinNamespaceResolver) resolveBarState(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "isfirst":
		return NamespaceResolution{Code: "(ctx.BarIndex == 0)", IsBool: true}, true
	case "islast":
		return NamespaceResolution{Code: "(ctx.BarIndex == len(ctx.Data)-1)", IsBool: true}, true
	case "ishistory":
		return NamespaceResolution{Code: "true", IsBool: true}, true
	case "isrealtime":
		return NamespaceResolution{Code: "false", IsBool: true}, true
	case "isnew":
		return NamespaceResolution{Code: "true", IsBool: true}, true
	case "isconfirmed":
		return NamespaceResolution{Code: "true", IsBool: true}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveTimeframe(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismonthly":
		return NamespaceResolution{Code: "ctx.IsMonthly", IsBool: true}, true
	case "isdaily":
		return NamespaceResolution{Code: "ctx.IsDaily", IsBool: true}, true
	case "isweekly":
		return NamespaceResolution{Code: "ctx.IsWeekly", IsBool: true}, true
	case "isintraday":
		return NamespaceResolution{Code: "ctx.IsIntraday", IsBool: true}, true
	case "period":
		return NamespaceResolution{Code: "ctx.Timeframe"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSyminfo(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "tickerid", "ticker":
		return NamespaceResolution{Code: "syminfo_tickerid"}, true
	case "timezone":
		return NamespaceResolution{Code: "ctx.Timezone"}, true
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

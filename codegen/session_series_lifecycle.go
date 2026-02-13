package codegen

import "fmt"

/* Regular variants (isfirstbar_regular/islastbar_regular) produce identical values
 * since our data feeds only contain regular session bars. */
type SessionSeriesLifecycle struct {
	hasIsfirstbar        bool
	hasIslastbar         bool
	hasIsfirstbarRegular bool
	hasIslastbarRegular  bool
}

func NewSessionSeriesLifecycle(hasIsfirstbar, hasIslastbar, hasIsfirstbarRegular, hasIslastbarRegular bool) *SessionSeriesLifecycle {
	return &SessionSeriesLifecycle{
		hasIsfirstbar:        hasIsfirstbar,
		hasIslastbar:         hasIslastbar,
		hasIsfirstbarRegular: hasIsfirstbarRegular,
		hasIslastbarRegular:  hasIslastbarRegular,
	}
}

func (l *SessionSeriesLifecycle) HasUsage() bool {
	return l != nil && (l.hasIsfirstbar || l.hasIslastbar || l.hasIsfirstbarRegular || l.hasIslastbarRegular)
}

func (l *SessionSeriesLifecycle) NeedsTimezone() bool {
	return l.HasUsage()
}

func (l *SessionSeriesLifecycle) GenerateDeclarations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasIsfirstbar {
		code += indent + "var session_isfirstbarSeries *series.Series\n"
	}
	if l.hasIslastbar {
		code += indent + "var session_islastbarSeries *series.Series\n"
	}
	if l.hasIsfirstbarRegular {
		code += indent + "var session_isfirstbar_regularSeries *series.Series\n"
	}
	if l.hasIslastbarRegular {
		code += indent + "var session_islastbar_regularSeries *series.Series\n"
	}
	return code
}

func (l *SessionSeriesLifecycle) GenerateInitializations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasIsfirstbar {
		code += indent + "session_isfirstbarSeries = series.NewSeries(len(ctx.Data))\n"
	}
	if l.hasIslastbar {
		code += indent + "session_islastbarSeries = series.NewSeries(len(ctx.Data))\n"
	}
	if l.hasIsfirstbarRegular {
		code += indent + "session_isfirstbar_regularSeries = series.NewSeries(len(ctx.Data))\n"
	}
	if l.hasIslastbarRegular {
		code += indent + "session_islastbar_regularSeries = series.NewSeries(len(ctx.Data))\n"
	}
	return code
}

/* Values stored as 1.0/0.0 since FSB only holds float64 */
func (l *SessionSeriesLifecycle) GenerateBarPopulation(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}

	needsFirst := l.hasIsfirstbar || l.hasIsfirstbarRegular
	needsLast := l.hasIslastbar || l.hasIslastbarRegular

	code := indent + "func() {\n"
	code += indent + "\tcurrDay := time.Unix(bar.Time, 0).In(exchangeLoc).YearDay()\n"
	code += indent + "\tcurrYear := time.Unix(bar.Time, 0).In(exchangeLoc).Year()\n"

	if needsFirst {
		code += indent + fmt.Sprintf("\tvar sessionIsFirst float64\n")
		code += indent + fmt.Sprintf("\tif %s == 0 {\n", iterVar)
		code += indent + "\t\tsessionIsFirst = 1.0\n"
		code += indent + "\t} else {\n"
		code += indent + fmt.Sprintf("\t\tprevDay := time.Unix(ctx.Data[%s-1].Time, 0).In(exchangeLoc).YearDay()\n", iterVar)
		code += indent + fmt.Sprintf("\t\tprevYear := time.Unix(ctx.Data[%s-1].Time, 0).In(exchangeLoc).Year()\n", iterVar)
		code += indent + "\t\tif currDay != prevDay || currYear != prevYear { sessionIsFirst = 1.0 }\n"
		code += indent + "\t}\n"
		if l.hasIsfirstbar {
			code += indent + "\tsession_isfirstbarSeries.Set(sessionIsFirst)\n"
		}
		if l.hasIsfirstbarRegular {
			code += indent + "\tsession_isfirstbar_regularSeries.Set(sessionIsFirst)\n"
		}
	}

	if needsLast {
		code += indent + fmt.Sprintf("\tvar sessionIsLast float64\n")
		code += indent + fmt.Sprintf("\tif %s >= barCount-1 {\n", iterVar)
		code += indent + "\t\tsessionIsLast = 1.0\n"
		code += indent + "\t} else {\n"
		code += indent + fmt.Sprintf("\t\tnextDay := time.Unix(ctx.Data[%s+1].Time, 0).In(exchangeLoc).YearDay()\n", iterVar)
		code += indent + fmt.Sprintf("\t\tnextYear := time.Unix(ctx.Data[%s+1].Time, 0).In(exchangeLoc).Year()\n", iterVar)
		code += indent + "\t\tif currDay != nextDay || currYear != nextYear { sessionIsLast = 1.0 }\n"
		code += indent + "\t}\n"
		if l.hasIslastbar {
			code += indent + "\tsession_islastbarSeries.Set(sessionIsLast)\n"
		}
		if l.hasIslastbarRegular {
			code += indent + "\tsession_islastbar_regularSeries.Set(sessionIsLast)\n"
		}
	}

	code += indent + "}()\n"
	return code
}

func (l *SessionSeriesLifecycle) GenerateAdvancement(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasIsfirstbar {
		code += indent + fmt.Sprintf("if %s < barCount-1 { session_isfirstbarSeries.Next() }\n", iterVar)
	}
	if l.hasIslastbar {
		code += indent + fmt.Sprintf("if %s < barCount-1 { session_islastbarSeries.Next() }\n", iterVar)
	}
	if l.hasIsfirstbarRegular {
		code += indent + fmt.Sprintf("if %s < barCount-1 { session_isfirstbar_regularSeries.Next() }\n", iterVar)
	}
	if l.hasIslastbarRegular {
		code += indent + fmt.Sprintf("if %s < barCount-1 { session_islastbar_regularSeries.Next() }\n", iterVar)
	}
	return code
}

func (l *SessionSeriesLifecycle) GenerateRegistrations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasIsfirstbar {
		code += indent + `ctx.RegisterSeries("session_isfirstbarSeries", session_isfirstbarSeries)` + "\n"
	}
	if l.hasIslastbar {
		code += indent + `ctx.RegisterSeries("session_islastbarSeries", session_islastbarSeries)` + "\n"
	}
	if l.hasIsfirstbarRegular {
		code += indent + `ctx.RegisterSeries("session_isfirstbar_regularSeries", session_isfirstbar_regularSeries)` + "\n"
	}
	if l.hasIslastbarRegular {
		code += indent + `ctx.RegisterSeries("session_islastbar_regularSeries", session_islastbar_regularSeries)` + "\n"
	}
	return code
}

func (l *SessionSeriesLifecycle) GenerateSuppressUnused(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	if l.hasIsfirstbar {
		code += indent + "_ = session_isfirstbarSeries\n"
	}
	if l.hasIslastbar {
		code += indent + "_ = session_islastbarSeries\n"
	}
	if l.hasIsfirstbarRegular {
		code += indent + "_ = session_isfirstbar_regularSeries\n"
	}
	if l.hasIslastbarRegular {
		code += indent + "_ = session_islastbar_regularSeries\n"
	}
	return code
}

func (l *SessionSeriesLifecycle) GenerateSymbolTableRegistrations(symbolTable SymbolTable) {
	if l == nil {
		return
	}
	if symbolTable == nil {
		return
	}
	if l.hasIsfirstbar {
		symbolTable.Register("session.isfirstbar", VariableTypeSeries)
	}
	if l.hasIslastbar {
		symbolTable.Register("session.islastbar", VariableTypeSeries)
	}
	if l.hasIsfirstbarRegular {
		symbolTable.Register("session.isfirstbar_regular", VariableTypeSeries)
	}
	if l.hasIslastbarRegular {
		symbolTable.Register("session.islastbar_regular", VariableTypeSeries)
	}
}

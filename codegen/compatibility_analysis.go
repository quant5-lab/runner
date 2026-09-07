package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type compatibilityDiagnostic struct {
	FeatureID    string
	Source       string
	Location     ast.SourceLocation
	Phase        string
	Impact       string
	Substitution string
	Sinks        []string
}

type compatibilityFeature struct {
	ID       string
	Source   string
	Location ast.SourceLocation
}

type compatibilityFeatureSet map[string]compatibilityFeature

type compatibilityAnalyzer struct {
	program       *ast.Program
	gaps          map[string]bool
	udfs          map[string]*ast.ArrowFunctionExpression
	boolConstants map[string]bool
	boolConflicts map[string]bool
	tainted       map[string]compatibilityFeatureSet
}

func analyzeCompatibility(program *ast.Program, featureGaps []string) []compatibilityDiagnostic {
	if program == nil {
		return nil
	}
	analyzer := newCompatibilityAnalyzer(program, featureGaps)
	return analyzer.Analyze()
}

func newCompatibilityAnalyzer(program *ast.Program, featureGaps []string) *compatibilityAnalyzer {
	gaps := make(map[string]bool, len(featureGaps))
	for _, name := range featureGaps {
		if isDependencyTrackedGap(name) {
			gaps[name] = true
		}
	}
	return &compatibilityAnalyzer{
		program:       program,
		gaps:          gaps,
		udfs:          collectCompatibilityUDFs(program),
		boolConstants: map[string]bool{},
		boolConflicts: map[string]bool{},
		tainted:       map[string]compatibilityFeatureSet{},
	}
}

func collectCompatibilityUDFs(program *ast.Program) map[string]*ast.ArrowFunctionExpression {
	udfs := map[string]*ast.ArrowFunctionExpression{}
	for _, node := range program.Body {
		decl, ok := node.(*ast.VariableDeclaration)
		if !ok {
			continue
		}
		for _, d := range decl.Declarations {
			id, ok := d.ID.(*ast.Identifier)
			if !ok {
				continue
			}
			arrow, ok := d.Init.(*ast.ArrowFunctionExpression)
			if ok {
				udfs[id.Name] = arrow
			}
		}
	}
	return udfs
}

func (a *compatibilityAnalyzer) Analyze() []compatibilityDiagnostic {
	a.expandBoolConstants()
	a.expandVariableTaint()

	diagnosticsByKey := map[string]compatibilityDiagnostic{}
	a.collectStatementDiagnostics(a.program.Body, diagnosticsByKey)
	a.collectBacktestDiagnostics(a.program.Body, diagnosticsByKey)

	out := make([]compatibilityDiagnostic, 0, len(diagnosticsByKey))
	for _, diag := range diagnosticsByKey {
		sort.Strings(diag.Sinks)
		out = append(out, diag)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FeatureID != out[j].FeatureID {
			return out[i].FeatureID < out[j].FeatureID
		}
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return strings.Join(out[i].Sinks, ",") < strings.Join(out[j].Sinks, ",")
	})
	return out
}

func (a *compatibilityAnalyzer) expandBoolConstants() {
	for changed := true; changed; {
		changed = false
		walkControlFlow(a.program.Body, func(node ast.Node) {
			decl, ok := node.(*ast.VariableDeclaration)
			if !ok {
				return
			}
			for _, d := range decl.Declarations {
				value, ok := a.boolConstantInExpression(d.Init)
				if !ok {
					continue
				}
				for _, name := range extractPatternNames(d.ID) {
					if a.boolConflicts[name] {
						continue
					}
					existing, exists := a.boolConstants[name]
					if exists && existing == value {
						continue
					}
					if exists {
						delete(a.boolConstants, name)
						a.boolConflicts[name] = true
						changed = true
						continue
					}
					a.boolConstants[name] = value
					changed = true
				}
			}
		})
	}
}

func isDependencyTrackedGap(name string) bool {
	switch name {
	case "input":
		return false
	default:
		return true
	}
}

func (a *compatibilityAnalyzer) expandVariableTaint() {
	for changed := true; changed; {
		changed = false
		walkControlFlow(a.program.Body, func(node ast.Node) {
			decl, ok := node.(*ast.VariableDeclaration)
			if !ok {
				return
			}
			for _, d := range decl.Declarations {
				features := a.featuresInExpression(d.Init, a.tainted, "generator.variable_init_unknown", nil)
				if len(features) == 0 {
					continue
				}
				for _, name := range extractPatternNames(d.ID) {
					if a.mergeVariableTaint(name, features) {
						changed = true
					}
				}
			}
		})
	}
}

func (a *compatibilityAnalyzer) mergeVariableTaint(name string, features compatibilityFeatureSet) bool {
	existing := a.tainted[name]
	if existing == nil {
		a.tainted[name] = cloneFeatureSet(features)
		return true
	}
	before := len(existing)
	mergeFeatureSets(existing, features)
	return len(existing) != before
}

func (a *compatibilityAnalyzer) collectStatementDiagnostics(nodes []ast.Node, out map[string]compatibilityDiagnostic) {
	for _, node := range nodes {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			call, ok := s.Expression.(*ast.CallExpression)
			if ok {
				a.recordUnsupportedStatementCall(call, out)
			}
		case *ast.IfStatement:
			a.collectStatementDiagnostics(s.Consequent, out)
			a.collectStatementDiagnostics(s.Alternate, out)
		case *ast.ForStatement:
			a.collectStatementDiagnostics(s.Body, out)
		case *ast.ForInStatement:
			a.collectStatementDiagnostics(s.Body, out)
		case *ast.WhileStatement:
			a.collectStatementDiagnostics(s.Body, out)
		}
	}
}

func (a *compatibilityAnalyzer) recordUnsupportedStatementCall(call *ast.CallExpression, out map[string]compatibilityDiagnostic) {
	name := extractCallFunctionName(call)
	if !a.gaps[name] || isBacktestCriticalSink(name) {
		return
	}
	addDiagnostic(out, compatibilityFeature{
		ID:       name,
		Source:   "generator.statement_unknown",
		Location: call.Location,
	}, "observable-non-backtest", "")
}

func (a *compatibilityAnalyzer) collectBacktestDiagnostics(nodes []ast.Node, out map[string]compatibilityDiagnostic) {
	for _, node := range nodes {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			if call, ok := s.Expression.(*ast.CallExpression); ok {
				a.recordCallSinkDiagnostic(call, nil, out)
			}
		case *ast.VariableDeclaration:
			for _, d := range s.Declarations {
				if arrow, ok := d.Init.(*ast.ArrowFunctionExpression); ok {
					a.collectBacktestDiagnostics(arrow.Body, out)
				}
			}
		case *ast.IfStatement:
			features := a.featuresInExpression(s.Test, a.tainted, "generator.variable_init_unknown", nil)
			a.recordNestedStrategySinks(s.Consequent, features, out)
			a.recordNestedStrategySinks(s.Alternate, features, out)
			a.collectBacktestDiagnostics(s.Consequent, out)
			a.collectBacktestDiagnostics(s.Alternate, out)
		case *ast.ForStatement:
			features := a.featuresInExpression(s.From, a.tainted, "generator.variable_init_unknown", nil)
			mergeFeatureSets(features, a.featuresInExpression(s.To, a.tainted, "generator.variable_init_unknown", nil))
			mergeFeatureSets(features, a.featuresInExpression(s.Step, a.tainted, "generator.variable_init_unknown", nil))
			a.recordNestedStrategySinks(s.Body, features, out)
			a.collectBacktestDiagnostics(s.Body, out)
		case *ast.ForInStatement:
			features := a.featuresInExpression(s.Collection, a.tainted, "generator.variable_init_unknown", nil)
			a.recordNestedStrategySinks(s.Body, features, out)
			a.collectBacktestDiagnostics(s.Body, out)
		case *ast.WhileStatement:
			features := a.featuresInExpression(s.Condition, a.tainted, "generator.variable_init_unknown", nil)
			a.recordNestedStrategySinks(s.Body, features, out)
			a.collectBacktestDiagnostics(s.Body, out)
		}
	}
}

func (a *compatibilityAnalyzer) recordNestedStrategySinks(nodes []ast.Node, features compatibilityFeatureSet, out map[string]compatibilityDiagnostic) {
	if len(features) == 0 {
		return
	}
	walkCallExpressions(nodes, func(call *ast.CallExpression) {
		sink := extractCallFunctionName(call)
		if !isBacktestCriticalSink(sink) {
			return
		}
		for _, feature := range sortedFeatures(features) {
			addDiagnostic(out, feature, "backtest-critical", sink)
		}
	})
}

func (a *compatibilityAnalyzer) recordCallSinkDiagnostic(call *ast.CallExpression, controlFeatures compatibilityFeatureSet, out map[string]compatibilityDiagnostic) {
	sink := extractCallFunctionName(call)
	if !isBacktestCriticalSink(sink) {
		return
	}
	for _, feature := range sortedFeatures(controlFeatures) {
		addDiagnostic(out, feature, "backtest-critical", sink)
	}
	for _, arg := range backtestCriticalArguments(sink, call.Arguments) {
		for _, feature := range sortedFeatures(a.featuresInExpression(arg, a.tainted, "generator.variable_init_unknown", nil)) {
			addDiagnostic(out, feature, "backtest-critical", sink)
		}
	}
}

func backtestCriticalArguments(sink string, args []ast.Expression) []ast.Expression {
	critical := make([]ast.Expression, 0, len(args))
	for index, arg := range args {
		if object, ok := arg.(*ast.ObjectExpression); ok {
			for _, prop := range object.Properties {
				name := propertyName(prop.Key)
				if isPresentationOnlyStrategyArg(sink, name) {
					continue
				}
				critical = append(critical, prop.Value)
			}
			continue
		}
		if isPresentationOnlyStrategyPositionalArg(sink, index) {
			continue
		}
		critical = append(critical, arg)
	}
	return critical
}

func propertyName(expr ast.Expression) string {
	id, ok := expr.(*ast.Identifier)
	if !ok {
		return ""
	}
	return id.Name
}

func isPresentationOnlyStrategyArg(_ string, name string) bool {
	switch name {
	case "comment", "comment_profit", "comment_loss", "comment_trailing",
		"alert_message", "alert_profit", "alert_loss", "alert_trailing",
		"disable_alert":
		return true
	default:
		return false
	}
}

func isPresentationOnlyStrategyPositionalArg(sink string, index int) bool {
	switch sink {
	case "strategy.entry", "strategy.order":
		return index == 7 || index == 8 || index == 9
	case "strategy.exit":
		return index >= 12 && index <= 20
	case "strategy.close":
		return index == 1 || index == 4 || index == 6
	case "strategy.close_all":
		return index == 0 || index == 1 || index == 3
	default:
		return false
	}
}

func (a *compatibilityAnalyzer) featuresInExpression(
	expr ast.Expression,
	tainted map[string]compatibilityFeatureSet,
	defaultSource string,
	udfStack map[string]bool,
) compatibilityFeatureSet {
	features := compatibilityFeatureSet{}
	if expr == nil {
		return features
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		mergeFeatureSets(features, tainted[e.Name])
	case *ast.CallExpression:
		name := extractCallFunctionName(e)
		switch {
		case isCalculationBearingChartCall(name):
			addFeature(features, name, "chart_namespace_getter", e.Location)
		case a.gaps[name]:
			addFeature(features, name, defaultSource, e.Location)
		case a.udfs[name] != nil:
			mergeFeatureSets(features, a.featuresInUDFCall(name, e, tainted, udfStack))
		}
		for _, arg := range e.Arguments {
			mergeFeatureSets(features, a.featuresInExpression(arg, tainted, defaultSource, udfStack))
		}
	case *ast.BinaryExpression:
		mergeFeatureSets(features, a.featuresInExpression(e.Left, tainted, defaultSource, udfStack))
		mergeFeatureSets(features, a.featuresInExpression(e.Right, tainted, defaultSource, udfStack))
	case *ast.LogicalExpression:
		if value, ok := a.boolConstantInExpression(e.Left); ok {
			mergeFeatureSets(features, a.featuresInExpression(e.Left, tainted, defaultSource, udfStack))
			if (e.Operator == "&&" && !value) || (e.Operator == "||" && value) {
				break
			}
			mergeFeatureSets(features, a.featuresInExpression(e.Right, tainted, defaultSource, udfStack))
			break
		}
		if value, ok := a.boolConstantInExpression(e.Right); ok {
			if (e.Operator == "&&" && !value) || (e.Operator == "||" && value) {
				mergeFeatureSets(features, a.featuresInExpression(e.Right, tainted, defaultSource, udfStack))
				break
			}
		}
		mergeFeatureSets(features, a.featuresInExpression(e.Left, tainted, defaultSource, udfStack))
		mergeFeatureSets(features, a.featuresInExpression(e.Right, tainted, defaultSource, udfStack))
	case *ast.ConditionalExpression:
		mergeFeatureSets(features, a.featuresInExpression(e.Test, tainted, defaultSource, udfStack))
		if value, ok := a.boolConstantInExpression(e.Test); ok {
			if value {
				mergeFeatureSets(features, a.featuresInExpression(e.Consequent, tainted, defaultSource, udfStack))
			} else {
				mergeFeatureSets(features, a.featuresInExpression(e.Alternate, tainted, defaultSource, udfStack))
			}
			break
		}
		mergeFeatureSets(features, a.featuresInExpression(e.Consequent, tainted, defaultSource, udfStack))
		mergeFeatureSets(features, a.featuresInExpression(e.Alternate, tainted, defaultSource, udfStack))
	case *ast.UnaryExpression:
		mergeFeatureSets(features, a.featuresInExpression(e.Argument, tainted, defaultSource, udfStack))
	case *ast.MemberExpression:
		mergeFeatureSets(features, a.featuresInExpression(e.Object, tainted, defaultSource, udfStack))
		if e.Computed {
			mergeFeatureSets(features, a.featuresInExpression(e.Property, tainted, defaultSource, udfStack))
		}
	case *ast.ObjectExpression:
		for _, prop := range e.Properties {
			mergeFeatureSets(features, a.featuresInExpression(prop.Value, tainted, defaultSource, udfStack))
		}
	}
	return features
}

func (a *compatibilityAnalyzer) boolConstantInExpression(expr ast.Expression) (bool, bool) {
	switch e := expr.(type) {
	case *ast.Literal:
		value, ok := e.Value.(bool)
		return value, ok
	case *ast.Identifier:
		value, ok := a.boolConstants[e.Name]
		return value, ok
	case *ast.CallExpression:
		if !isInputCallExpr(e) {
			return false, false
		}
		lit, ok := extractDefvalExpression(e).(*ast.Literal)
		if !ok {
			return false, false
		}
		value, ok := lit.Value.(bool)
		return value, ok
	case *ast.UnaryExpression:
		if e.Operator != "!" {
			return false, false
		}
		value, ok := a.boolConstantInExpression(e.Argument)
		return !value, ok
	case *ast.ConditionalExpression:
		testValue, ok := a.boolConstantInExpression(e.Test)
		if !ok {
			return false, false
		}
		if testValue {
			return a.boolConstantInExpression(e.Consequent)
		}
		return a.boolConstantInExpression(e.Alternate)
	case *ast.LogicalExpression:
		leftValue, leftOK := a.boolConstantInExpression(e.Left)
		rightValue, rightOK := a.boolConstantInExpression(e.Right)
		switch e.Operator {
		case "&&":
			if leftOK && !leftValue {
				return false, true
			}
			if rightOK && !rightValue {
				return false, true
			}
			if leftOK && rightOK {
				return leftValue && rightValue, true
			}
		case "||":
			if leftOK && leftValue {
				return true, true
			}
			if rightOK && rightValue {
				return true, true
			}
			if leftOK && rightOK {
				return leftValue || rightValue, true
			}
		}
	}
	return false, false
}

func (a *compatibilityAnalyzer) featuresInUDFCall(
	name string,
	call *ast.CallExpression,
	outerTaint map[string]compatibilityFeatureSet,
	udfStack map[string]bool,
) compatibilityFeatureSet {
	if udfStack == nil {
		udfStack = map[string]bool{}
	}
	if udfStack[name] {
		return nil
	}
	arrow := a.udfs[name]
	if arrow == nil {
		return nil
	}

	localTaint := map[string]compatibilityFeatureSet{}
	for key, value := range outerTaint {
		localTaint[key] = cloneFeatureSet(value)
	}
	for i, param := range arrow.Params {
		if i >= len(call.Arguments) {
			continue
		}
		features := a.featuresInExpression(call.Arguments[i], outerTaint, "generator.variable_init_unknown", udfStack)
		if len(features) > 0 {
			localTaint[param.Name] = cloneFeatureSet(features)
		}
	}

	nextStack := map[string]bool{}
	for key, value := range udfStack {
		nextStack[key] = value
	}
	nextStack[name] = true
	return a.featuresInNodes(arrow.Body, localTaint, "arrow_expression", nextStack)
}

func (a *compatibilityAnalyzer) featuresInNodes(
	nodes []ast.Node,
	tainted map[string]compatibilityFeatureSet,
	defaultSource string,
	udfStack map[string]bool,
) compatibilityFeatureSet {
	features := compatibilityFeatureSet{}
	for _, node := range nodes {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			mergeFeatureSets(features, a.featuresInExpression(s.Expression, tainted, defaultSource, udfStack))
		case *ast.VariableDeclaration:
			for _, d := range s.Declarations {
				initFeatures := a.featuresInExpression(d.Init, tainted, defaultSource, udfStack)
				mergeFeatureSets(features, initFeatures)
				for _, name := range extractPatternNames(d.ID) {
					if tainted[name] == nil {
						tainted[name] = compatibilityFeatureSet{}
					}
					mergeFeatureSets(tainted[name], initFeatures)
				}
			}
		case *ast.IfStatement:
			mergeFeatureSets(features, a.featuresInExpression(s.Test, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInNodes(s.Consequent, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInNodes(s.Alternate, tainted, defaultSource, udfStack))
		case *ast.ForStatement:
			mergeFeatureSets(features, a.featuresInExpression(s.From, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInExpression(s.To, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInExpression(s.Step, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInNodes(s.Body, tainted, defaultSource, udfStack))
		case *ast.ForInStatement:
			mergeFeatureSets(features, a.featuresInExpression(s.Collection, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInNodes(s.Body, tainted, defaultSource, udfStack))
		case *ast.WhileStatement:
			mergeFeatureSets(features, a.featuresInExpression(s.Condition, tainted, defaultSource, udfStack))
			mergeFeatureSets(features, a.featuresInNodes(s.Body, tainted, defaultSource, udfStack))
		}
	}
	return features
}

func addFeature(features compatibilityFeatureSet, id, source string, location ast.SourceLocation) {
	if existing, ok := features[id]; ok {
		if existing.Location.IsZero() && !location.IsZero() {
			existing.Location = location
			features[id] = existing
		}
		return
	}
	features[id] = compatibilityFeature{ID: id, Source: source, Location: location}
}

func mergeFeatureSets(dst, src compatibilityFeatureSet) {
	for id, feature := range src {
		if existing, ok := dst[id]; ok {
			if existing.Location.IsZero() && !feature.Location.IsZero() {
				existing.Location = feature.Location
				dst[id] = existing
			}
			continue
		}
		dst[id] = feature
	}
}

func cloneFeatureSet(src compatibilityFeatureSet) compatibilityFeatureSet {
	if len(src) == 0 {
		return nil
	}
	dst := make(compatibilityFeatureSet, len(src))
	for id, feature := range src {
		dst[id] = feature
	}
	return dst
}

func sortedFeatures(features compatibilityFeatureSet) []compatibilityFeature {
	out := make([]compatibilityFeature, 0, len(features))
	for _, feature := range features {
		out = append(out, feature)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].Source < out[j].Source
	})
	return out
}

func addDiagnostic(out map[string]compatibilityDiagnostic, feature compatibilityFeature, impact, sink string) {
	key := feature.ID + "\x00" + feature.Source
	if sink != "" {
		key += "\x00" + sink
	}
	if existing, ok := out[key]; ok {
		existing.Sinks = appendUnique(existing.Sinks, sink)
		if existing.Location.IsZero() && !feature.Location.IsZero() {
			existing.Location = feature.Location
		}
		if compatibilityImpactRank(impact) > compatibilityImpactRank(existing.Impact) {
			existing.Impact = impact
		}
		out[key] = existing
		return
	}
	out[key] = compatibilityDiagnostic{
		FeatureID:    feature.ID,
		Source:       feature.Source,
		Location:     feature.Location,
		Phase:        "codegen",
		Impact:       impact,
		Substitution: "NaN",
		Sinks:        appendUnique(nil, sink),
	}
}

func compatibilityImpactRank(impact string) int {
	switch impact {
	case "silent-presentation":
		return 1
	case "observable-non-backtest":
		return 2
	case "backtest-critical":
		return 3
	case "structural":
		return 4
	default:
		return 0
	}
}

func diagnosticSource(featureID string) string {
	if isCalculationBearingChartCall(featureID) {
		return "chart_namespace_getter"
	}
	return "generator.variable_init_unknown"
}

func isBacktestCriticalSink(name string) bool {
	switch name {
	case "strategy.entry", "strategy.order", "strategy.exit", "strategy.close",
		"strategy.close_all", "strategy.cancel", "strategy.cancel_all",
		"strategy.risk.allow_entry_in", "strategy.risk.max_cons_loss_days",
		"strategy.risk.max_drawdown", "strategy.risk.max_intraday_filled_orders",
		"strategy.risk.max_intraday_loss", "strategy.risk.max_position_size":
		return true
	default:
		return false
	}
}

func compatibilitySetupCode(diagnostics []compatibilityDiagnostic) string {
	if len(diagnostics) == 0 {
		return ""
	}
	var b strings.Builder
	for _, diag := range diagnostics {
		fmt.Fprintf(&b, "\tfeaturegap.RecordStaticAt(%q, %q, %q, %d, %d, %q, %q, %q, []string{%s})\n",
			diag.FeatureID,
			diag.Source,
			diag.Location.File,
			diag.Location.Line,
			diag.Location.Column,
			diag.Phase,
			diag.Impact,
			diag.Substitution,
			quotedStrings(diag.Sinks),
		)
	}
	return b.String()
}

func quotedStrings(values []string) string {
	if len(values) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, ", ")
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

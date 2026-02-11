package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestPeriodExpression_InputInt(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: input.int()")
plot(ta.sma(close, input.int(14, "Period")))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-input-int", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_CustomFunction(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: getPeriod()")
getPeriod() => 14
plot(ta.sma(close, getPeriod()))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-custom-func", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_Variable(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: varPeriod")
varPeriod = 14
plot(ta.sma(close, varPeriod))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-variable", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_BinaryExpr(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: 7 * 2")
plot(ta.sma(close, 7 * 2))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-binary-expr", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_Highest(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: ta.highest with getPeriod()")
getPeriod() => 14
plot(ta.highest(close, getPeriod()))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-highest", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_Stdev(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Period: ta.stdev with getPeriod()")
getPeriod() => 14
plot(ta.stdev(close, getPeriod()))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-stdev", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

func TestPeriodExpression_ConstantFolding(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Constant folding: basePeriod + 6")
basePeriod = 8
plot(ta.sma(close, basePeriod + 6))`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "period-const-fold", pineScript)

	values := exec.ExtractPlotValues(t, output, "Plot 1")
	if len(values) == 0 {
		t.Fatal("Expected plot values")
	}
}

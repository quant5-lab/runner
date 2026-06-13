package testutil

import "testing"

func TestFinancialToleranceHeadroom(t *testing.T) {
	if MeasuredSizeResidualBound <= 0 {
		t.Fatalf("MeasuredSizeResidualBound = %.10f, want positive", MeasuredSizeResidualBound)
	}
	if FinancialToleranceSafetyMultiple < 5 {
		t.Fatalf("FinancialToleranceSafetyMultiple = %.2f, want at least 5", FinancialToleranceSafetyMultiple)
	}
	want := MeasuredSizeResidualBound * FinancialToleranceSafetyMultiple
	if FinancialRelEps != want {
		t.Fatalf("FinancialRelEps = %.10f, want MeasuredSizeResidualBound * FinancialToleranceSafetyMultiple = %.10f", FinancialRelEps, want)
	}
}

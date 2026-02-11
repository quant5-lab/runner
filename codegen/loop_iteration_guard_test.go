package codegen

import (
	"strings"
	"testing"
)

func TestLoopIterationGuard_InitCode(t *testing.T) {
	guard := NewLoopIterationGuard()
	got := guard.InitCode("\t\t")

	if got != "\t\t__whileGuard := 0\n" {
		t.Errorf("InitCode() = %q, want %q", got, "\t\t__whileGuard := 0\n")
	}
}

func TestLoopIterationGuard_CheckCode(t *testing.T) {
	guard := NewLoopIterationGuard()
	got := guard.CheckCode("\t\t")

	if !strings.Contains(got, "__whileGuard++") {
		t.Error("CheckCode() missing increment")
	}
	if !strings.Contains(got, "100000") {
		t.Error("CheckCode() missing max iterations constant")
	}
	if !strings.Contains(got, "break") {
		t.Error("CheckCode() missing break")
	}
}

func TestLoopIterationGuard_CheckCodeFormat(t *testing.T) {
	guard := NewLoopIterationGuard()
	got := guard.CheckCode("\t")

	expected := "\tif __whileGuard++; __whileGuard > 100000 {\n\t\tbreak\n\t}\n"
	if got != expected {
		t.Errorf("CheckCode() =\n%q\nwant:\n%q", got, expected)
	}
}

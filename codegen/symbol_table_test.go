package codegen

import "testing"

func TestVariableType(t *testing.T) {
	tests := []struct {
		name       string
		varType    VariableType
		isSeries   bool
		isScalar   bool
		stringRepr string
	}{
		{"scalar type", VariableTypeScalar, false, true, "scalar"},
		{"series type", VariableTypeSeries, true, false, "series"},
		{"function type", VariableTypeFunction, false, false, "function"},
		{"unknown type", VariableTypeUnknown, false, false, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.varType.IsSeries(); got != tt.isSeries {
				t.Errorf("IsSeries() = %v, want %v", got, tt.isSeries)
			}
			if got := tt.varType.IsScalar(); got != tt.isScalar {
				t.Errorf("IsScalar() = %v, want %v", got, tt.isScalar)
			}
			if got := tt.varType.String(); got != tt.stringRepr {
				t.Errorf("String() = %v, want %v", got, tt.stringRepr)
			}
		})
	}
}

func TestSymbolTable(t *testing.T) {
	t.Run("register and lookup", func(t *testing.T) {
		st := NewSymbolTable()

		st.Register("price", VariableTypeSeries)
		st.Register("count", VariableTypeScalar)

		if got := st.Lookup("price"); got != VariableTypeSeries {
			t.Errorf("Lookup(price) = %v, want series", got)
		}
		if got := st.Lookup("count"); got != VariableTypeScalar {
			t.Errorf("Lookup(count) = %v, want scalar", got)
		}
		if got := st.Lookup("unknown"); got != VariableTypeUnknown {
			t.Errorf("Lookup(unknown) = %v, want unknown", got)
		}
	})

	t.Run("series and scalar checks", func(t *testing.T) {
		st := NewSymbolTable()

		st.Register("close", VariableTypeSeries)
		st.Register("volume", VariableTypeScalar)

		if !st.IsSeries("close") {
			t.Error("IsSeries(close) should be true")
		}
		if st.IsScalar("close") {
			t.Error("IsScalar(close) should be false")
		}
		if st.IsSeries("volume") {
			t.Error("IsSeries(volume) should be false")
		}
		if !st.IsScalar("volume") {
			t.Error("IsScalar(volume) should be true")
		}
	})

	t.Run("clone creates independent copy", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("original", VariableTypeSeries)

		clone := st.Clone()
		clone.Register("cloned", VariableTypeScalar)

		if st.Lookup("cloned") != VariableTypeUnknown {
			t.Error("Original should not have cloned symbol")
		}
		if clone.Lookup("original") != VariableTypeSeries {
			t.Error("Clone should have original symbol")
		}
	})

	t.Run("merge combines symbols", func(t *testing.T) {
		st1 := NewSymbolTable()
		st1.Register("var1", VariableTypeSeries)

		st2 := NewSymbolTable()
		st2.Register("var2", VariableTypeScalar)

		st1.Merge(st2)

		if st1.Lookup("var2") != VariableTypeScalar {
			t.Error("Merged symbol should exist")
		}
		if st1.Lookup("var1") != VariableTypeSeries {
			t.Error("Original symbol should remain")
		}
	})

	t.Run("overwrite existing symbol", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("var", VariableTypeScalar)
		st.Register("var", VariableTypeSeries)

		if st.Lookup("var") != VariableTypeSeries {
			t.Error("Symbol should be overwritten")
		}
	})
}

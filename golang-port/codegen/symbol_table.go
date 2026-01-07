package codegen

// SymbolInfo holds type information for a single variable
type SymbolInfo struct {
	Name string
	Type VariableType
}

// SymbolTable tracks variable type information during code generation
// Responsibility: Maintain variable→type mappings for type-aware code generation
type SymbolTable interface {
	// Register declares a variable with its type
	Register(name string, varType VariableType)

	// Lookup retrieves type information for a variable
	// Returns VariableTypeUnknown if variable not registered
	Lookup(name string) VariableType

	// IsSeries checks if a variable is of series type
	IsSeries(name string) bool

	// IsScalar checks if a variable is of scalar type
	IsScalar(name string) bool

	// Clone creates an independent copy for nested scopes
	Clone() SymbolTable

	// Merge combines symbols from another table (for scope hierarchies)
	Merge(other SymbolTable)

	// AllSymbols returns all registered symbols
	AllSymbols() []SymbolInfo
}

// NewSymbolTable creates a new symbol table instance
func NewSymbolTable() SymbolTable {
	return &symbolTableImpl{
		symbols: make(map[string]VariableType),
	}
}

type symbolTableImpl struct {
	symbols map[string]VariableType
}

func (s *symbolTableImpl) Register(name string, varType VariableType) {
	s.symbols[name] = varType
}

func (s *symbolTableImpl) Lookup(name string) VariableType {
	if varType, exists := s.symbols[name]; exists {
		return varType
	}
	return VariableTypeUnknown
}

func (s *symbolTableImpl) IsSeries(name string) bool {
	return s.Lookup(name).IsSeries()
}

func (s *symbolTableImpl) IsScalar(name string) bool {
	return s.Lookup(name).IsScalar()
}

func (s *symbolTableImpl) Clone() SymbolTable {
	clone := &symbolTableImpl{
		symbols: make(map[string]VariableType, len(s.symbols)),
	}
	for name, varType := range s.symbols {
		clone.symbols[name] = varType
	}
	return clone
}

func (s *symbolTableImpl) Merge(other SymbolTable) {
	if otherImpl, ok := other.(*symbolTableImpl); ok {
		for name, varType := range otherImpl.symbols {
			s.symbols[name] = varType
		}
	}
}

func (s *symbolTableImpl) AllSymbols() []SymbolInfo {
	result := make([]SymbolInfo, 0, len(s.symbols))
	for name, varType := range s.symbols {
		result = append(result, SymbolInfo{Name: name, Type: varType})
	}
	return result
}

package validation

type ConstantRegistry struct {
	store map[string]float64
}

func NewConstantRegistry() *ConstantRegistry {
	return &ConstantRegistry{
		store: make(map[string]float64),
	}
}

func (r *ConstantRegistry) Set(name string, value float64) {
	r.store[name] = value
}

func (r *ConstantRegistry) Get(name string) (float64, bool) {
	value, exists := r.store[name]
	return value, exists
}

func (r *ConstantRegistry) Clear() {
	r.store = make(map[string]float64)
}

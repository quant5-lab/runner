package input

/* Manager handles input parameter overrides */
type Manager struct {
	overrides map[string]interface{}
}

/* NewManager creates input manager with override map */
func NewManager(overrides map[string]interface{}) *Manager {
	if overrides == nil {
		overrides = make(map[string]interface{})
	}
	return &Manager{
		overrides: overrides,
	}
}

/* Int returns int input with title-based override support */
func (m *Manager) Int(defval int, title string) int {
	if title != "" {
		if override, exists := m.overrides[title]; exists {
			if v, ok := override.(int); ok {
				return v
			}
			if v, ok := override.(float64); ok {
				return int(v)
			}
		}
	}
	return defval
}

/* Float returns float input with title-based override support */
func (m *Manager) Float(defval float64, title string) float64 {
	if title != "" {
		if override, exists := m.overrides[title]; exists {
			if v, ok := override.(float64); ok {
				return v
			}
			if v, ok := override.(int); ok {
				return float64(v)
			}
		}
	}
	return defval
}

/* String returns string input with title-based override support */
func (m *Manager) String(defval, title string) string {
	if title != "" {
		if override, exists := m.overrides[title]; exists {
			if v, ok := override.(string); ok {
				return v
			}
		}
	}
	return defval
}

/* Bool returns bool input with title-based override support */
func (m *Manager) Bool(defval bool, title string) bool {
	if title != "" {
		if override, exists := m.overrides[title]; exists {
			if v, ok := override.(bool); ok {
				return v
			}
		}
	}
	return defval
}

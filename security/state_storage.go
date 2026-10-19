package security

type StateStorage interface {
	Get(key string) (interface{}, bool)
	Set(key string, state interface{})
	Has(key string) bool
}

type MapStateStorage struct {
	storage map[string]interface{}
}

func NewMapStateStorage() *MapStateStorage {
	return &MapStateStorage{
		storage: make(map[string]interface{}),
	}
}

func (s *MapStateStorage) Get(key string) (interface{}, bool) {
	state, exists := s.storage[key]
	return state, exists
}

func (s *MapStateStorage) Set(key string, state interface{}) {
	s.storage[key] = state
}

func (s *MapStateStorage) Has(key string) bool {
	_, exists := s.storage[key]
	return exists
}

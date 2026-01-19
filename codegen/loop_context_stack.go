package codegen

type LoopContext struct {
	CounterVariable string
	IsActive        bool
}

type LoopContextStack struct {
	contexts []LoopContext
}

func NewLoopContextStack() *LoopContextStack {
	return &LoopContextStack{
		contexts: make([]LoopContext, 0, 4),
	}
}

func (s *LoopContextStack) Push(counterVar string) {
	s.contexts = append(s.contexts, LoopContext{
		CounterVariable: counterVar,
		IsActive:        true,
	})
}

func (s *LoopContextStack) Pop() {
	if len(s.contexts) > 0 {
		s.contexts = s.contexts[:len(s.contexts)-1]
	}
}

func (s *LoopContextStack) IsInLoop() bool {
	return len(s.contexts) > 0
}

func (s *LoopContextStack) CurrentCounter() string {
	if len(s.contexts) == 0 {
		return ""
	}
	return s.contexts[len(s.contexts)-1].CounterVariable
}

func (s *LoopContextStack) IsLoopCounter(name string) bool {
	for i := range s.contexts {
		if s.contexts[i].CounterVariable == name {
			return true
		}
	}
	return false
}

func (s *LoopContextStack) Depth() int {
	return len(s.contexts)
}

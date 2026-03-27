package security

import "testing"

func TestMatchIndexRing_EmptyRing(t *testing.T) {
	ring := NewMatchIndexRing(2)

	if ring.Size() != 0 {
		t.Errorf("expected empty ring, got size %d", ring.Size())
	}

	if _, ok := ring.GetNthMostRecent(0); ok {
		t.Error("expected GetNthMostRecent(0) to fail on empty ring")
	}
}

func TestMatchIndexRing_SingleElement(t *testing.T) {
	ring := NewMatchIndexRing(2)
	ring.Push(42)

	if ring.Size() != 1 {
		t.Errorf("expected size 1, got %d", ring.Size())
	}

	val, ok := ring.GetNthMostRecent(0)
	if !ok {
		t.Fatal("expected GetNthMostRecent(0) to succeed")
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	if _, ok := ring.GetNthMostRecent(1); ok {
		t.Error("expected GetNthMostRecent(1) to fail with only 1 element")
	}
}

func TestMatchIndexRing_FillsToCapacity(t *testing.T) {
	ring := NewMatchIndexRing(2)
	ring.Push(10)
	ring.Push(20)
	ring.Push(30)

	if ring.Size() != 3 {
		t.Errorf("expected size 3, got %d", ring.Size())
	}

	tests := []struct {
		n        int
		expected int
	}{
		{0, 30},
		{1, 20},
		{2, 10},
	}

	for _, tt := range tests {
		val, ok := ring.GetNthMostRecent(tt.n)
		if !ok {
			t.Fatalf("expected GetNthMostRecent(%d) to succeed", tt.n)
		}
		if val != tt.expected {
			t.Errorf("GetNthMostRecent(%d): expected %d, got %d", tt.n, tt.expected, val)
		}
	}
}

func TestMatchIndexRing_ExceedsCapacity(t *testing.T) {
	ring := NewMatchIndexRing(1)
	ring.Push(100)
	ring.Push(200)
	ring.Push(300)

	if ring.Size() != 2 {
		t.Errorf("expected size 2 (capped at capacity), got %d", ring.Size())
	}

	val0, ok := ring.GetNthMostRecent(0)
	if !ok || val0 != 300 {
		t.Errorf("expected most recent = 300, got %d (ok=%v)", val0, ok)
	}

	val1, ok := ring.GetNthMostRecent(1)
	if !ok || val1 != 200 {
		t.Errorf("expected second most recent = 200, got %d (ok=%v)", val1, ok)
	}

	if _, ok := ring.GetNthMostRecent(2); ok {
		t.Error("expected GetNthMostRecent(2) to fail (oldest element 100 evicted)")
	}
}

func TestMatchIndexRing_OccurrenceZero(t *testing.T) {
	ring := NewMatchIndexRing(0)
	ring.Push(5)
	ring.Push(10)
	ring.Push(15)

	if ring.Size() != 1 {
		t.Errorf("expected size 1 (occurrence=0 means capacity=1), got %d", ring.Size())
	}

	val, ok := ring.GetNthMostRecent(0)
	if !ok || val != 15 {
		t.Errorf("expected most recent = 15, got %d (ok=%v)", val, ok)
	}
}

func TestMatchIndexRing_BoundaryConditions(t *testing.T) {
	ring := NewMatchIndexRing(1)

	if _, ok := ring.GetNthMostRecent(-1); ok {
		t.Error("expected negative index to fail")
	}

	ring.Push(7)
	if _, ok := ring.GetNthMostRecent(10); ok {
		t.Error("expected out-of-bounds index to fail")
	}
}

func TestMatchIndexRing_SequentialPushes(t *testing.T) {
	ring := NewMatchIndexRing(3)
	indices := []int{0, 5, 12, 47, 103, 250}

	for _, idx := range indices {
		ring.Push(idx)
	}

	if ring.Size() != 4 {
		t.Errorf("expected size 4 (capacity), got %d", ring.Size())
	}

	expected := []int{250, 103, 47, 12}
	for n, exp := range expected {
		val, ok := ring.GetNthMostRecent(n)
		if !ok {
			t.Fatalf("expected GetNthMostRecent(%d) to succeed", n)
		}
		if val != exp {
			t.Errorf("GetNthMostRecent(%d): expected %d, got %d", n, exp, val)
		}
	}
}

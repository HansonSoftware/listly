package main

import (
	"testing"
)

func TestUndoStack_PushPop(t *testing.T) {
	u := NewUndoStack(10)

	called := false
	u.Push(func() { called = true })

	if u.Len() != 1 {
		t.Errorf("Len() = %d, want 1", u.Len())
	}

	fn := u.Pop()
	if fn == nil {
		t.Fatal("Pop() returned nil")
	}
	fn()

	if !called {
		t.Error("Popped function was not called")
	}
	if u.Len() != 0 {
		t.Errorf("Len() = %d after Pop, want 0", u.Len())
	}
}

func TestUndoStack_Empty(t *testing.T) {
	u := NewUndoStack(10)
	if fn := u.Pop(); fn != nil {
		t.Error("Pop() on empty stack should return nil")
	}
}

func TestUndoStack_MaxCap(t *testing.T) {
	u := NewUndoStack(3)

	for i := 0; i < 5; i++ {
		u.Push(func() {})
	}

	if u.Len() != 3 {
		t.Errorf("Len() = %d, want 3 (capped)", u.Len())
	}
}

func TestUndoStack_Clear(t *testing.T) {
	u := NewUndoStack(10)
	u.Push(func() {})
	u.Push(func() {})

	u.Clear()

	if u.Len() != 0 {
		t.Errorf("Len() = %d after Clear, want 0", u.Len())
	}
}

func TestUndoStack_Order(t *testing.T) {
	u := NewUndoStack(10)

	var order []int
	for i := 0; i < 3; i++ {
		i := i // capture
		u.Push(func() { order = append(order, i) })
	}

	// Pop executes in reverse order (LIFO)
	for i := 0; i < 3; i++ {
		if fn := u.Pop(); fn != nil {
			fn()
		}
	}

	if len(order) != 3 || order[0] != 2 || order[1] != 1 || order[2] != 0 {
		t.Errorf("Undo order = %v, want [2, 1, 0]", order)
	}
}

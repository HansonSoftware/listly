package main

import (
	"testing"
)

func TestUndoStack_PushPop(t *testing.T) {
	u := NewUndoStack(10)

	called := false
	u.Push(func() { called = true }, func() {})

	if u.Len() != 1 {
		t.Errorf("Len() = %d, want 1", u.Len())
	}

	o, ok := u.Pop()
	if !ok {
		t.Fatal("Pop() reported empty")
	}
	o.undo()

	if !called {
		t.Error("Popped undo function was not called")
	}
	if u.Len() != 0 {
		t.Errorf("Len() = %d after Pop, want 0", u.Len())
	}
}

func TestUndoStack_Empty(t *testing.T) {
	u := NewUndoStack(10)
	if _, ok := u.Pop(); ok {
		t.Error("Pop() on empty stack should report empty")
	}
}

func TestUndoStack_MaxCap(t *testing.T) {
	u := NewUndoStack(3)

	for i := 0; i < 5; i++ {
		u.Push(func() {}, func() {})
	}

	if u.Len() != 3 {
		t.Errorf("Len() = %d, want 3 (capped)", u.Len())
	}
}

func TestUndoStack_Clear(t *testing.T) {
	u := NewUndoStack(10)
	u.Push(func() {}, func() {})
	u.Push(func() {}, func() {})

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
		u.Push(func() { order = append(order, i) }, func() {})
	}

	// Pop executes in reverse order (LIFO)
	for i := 0; i < 3; i++ {
		if o, ok := u.Pop(); ok {
			o.undo()
		}
	}

	if len(order) != 3 || order[0] != 2 || order[1] != 1 || order[2] != 0 {
		t.Errorf("Undo order = %v, want [2, 1, 0]", order)
	}
}

func TestUndoStack_RedoPair(t *testing.T) {
	u := NewUndoStack(10)
	states := []string{"a"}
	u.Push(
		func() { states = append(states, "b") }, // undo -> forward op
		func() { states = append(states, "a") }, // redo -> reverse op
	)
	o, _ := u.Pop()
	o.undo()
	if states[len(states)-1] != "b" {
		t.Fatalf("undo did not run: %v", states)
	}
	o.redo()
	if states[len(states)-1] != "a" {
		t.Fatalf("redo did not run: %v", states)
	}
}
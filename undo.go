package main

// UndoStack stores reverse operations for undo functionality
type UndoStack struct {
	ops   []func()
	max   int
}

// NewUndoStack creates a new undo stack with max capacity
func NewUndoStack(maxOps int) *UndoStack {
	return &UndoStack{
		ops: make([]func(), 0),
		max: maxOps,
	}
}

// Push adds a reverse operation to the stack
func (u *UndoStack) Push(fn func()) {
	u.ops = append(u.ops, fn)
	if len(u.ops) > u.max {
		u.ops = u.ops[1:]
	}
}

// Pop removes and returns the last operation
func (u *UndoStack) Pop() func() {
	if len(u.ops) == 0 {
		return nil
	}
	fn := u.ops[len(u.ops)-1]
	u.ops = u.ops[:len(u.ops)-1]
	return fn
}

// Len returns the number of operations
func (u *UndoStack) Len() int {
	return len(u.ops)
}

// Clear removes all operations
func (u *UndoStack) Clear() {
	u.ops = u.ops[:0]
}

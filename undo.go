package main

// op is a paired undo/redo operation.
type op struct {
	undo func()
	redo func()
}

// UndoStack stores undo/redo pairs with a fixed capacity.
type UndoStack struct {
	ops []op
	max int
}

// NewUndoStack creates a new undo stack with max capacity
func NewUndoStack(maxOps int) *UndoStack {
	return &UndoStack{
		ops: make([]op, 0),
		max: maxOps,
	}
}

// Push adds an undo/redo pair to the stack
func (u *UndoStack) Push(undo, redo func()) {
	u.ops = append(u.ops, op{undo: undo, redo: redo})
	if len(u.ops) > u.max {
		u.ops = u.ops[1:]
	}
}

// Pop removes and returns the last operation pair
func (u *UndoStack) Pop() (op, bool) {
	if len(u.ops) == 0 {
		return op{}, false
	}
	o := u.ops[len(u.ops)-1]
	u.ops = u.ops[:len(u.ops)-1]
	return o, true
}

// Len returns the number of operations
func (u *UndoStack) Len() int {
	return len(u.ops)
}

// Clear removes all operations
func (u *UndoStack) Clear() {
	u.ops = u.ops[:0]
}

// History holds one session's undo/redo stacks.
type History struct {
	undo *UndoStack
	redo *UndoStack
}

// NewHistory creates an empty history for a session.
func NewHistory(maxOps int) *History {
	return &History{undo: NewUndoStack(maxOps), redo: NewUndoStack(maxOps)}
}

// histories keeps an undo/redo history per session ID for the lifetime of the
// program, so toggling between sessions preserves each one's undo/redo state.
type histories map[int64]*History

func (h histories) get(id int64) *History {
	if hist, ok := h[id]; ok {
		return hist
	}
	hist := NewHistory(historyMaxOps)
	h[id] = hist
	return hist
}

const historyMaxOps = 50
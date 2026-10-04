package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func newTestModel(tasks ...Task) *Model {
	m := &Model{
		focused:   todo,
		isDaily:   true, // skip store persistence in tests
		undoStack: NewUndoStack(50),
		lists:     make([]list.Model, 3),
	}
	delegate := TaskDelegate{}
	for i := range m.lists {
		var items []list.Item
		for _, t := range tasks {
			if t.status == status(i) {
				items = append(items, t)
			}
		}
		m.lists[i] = list.New(items, delegate, 40, 20)
	}
	return m
}

func TestMoveToNext_UndoRestoresOriginalColumn(t *testing.T) {
	task := NewTask(todo, "Task A", "desc")
	m := newTestModel(task)

	m.MoveToNext()

	if got := len(m.lists[todo].Items()); got != 0 {
		t.Fatalf("todo should be empty after move, got %d", got)
	}
	if got := len(m.lists[completing].Items()); got != 1 {
		t.Fatalf("completing should have 1 item after move, got %d", got)
	}

	m.undo()

	if got := len(m.lists[completing].Items()); got != 0 {
		t.Fatalf("completing should be empty after undo, got %d", got)
	}
	items := m.lists[todo].Items()
	if len(items) != 1 {
		t.Fatalf("todo should have 1 item after undo, got %d", len(items))
	}
	restored := items[0].(Task)
	if restored.Title() != "Task A" || restored.Status() != todo {
		t.Errorf("undo restored wrong task: title=%q status=%d", restored.Title(), restored.Status())
	}
}

func TestMoveToNext_UndoRestoresOriginalIndex(t *testing.T) {
	first := NewTask(todo, "First", "")
	second := NewTask(todo, "Second", "")
	m := newTestModel(first, second)

	// Move "First" (index 0) to completing.
	m.MoveToNext()
	// Move "Second" (now index 0 in todo) to completing.
	m.MoveToNext()

	// Undo last move: "Second" should return to todo at index 0.
	m.undo()
	items := m.lists[todo].Items()
	if len(items) != 1 || items[0].(Task).Title() != "Second" {
		t.Fatalf("expected todo[0]='Second' after undo, got %v", items)
	}

	// Undo first move: "First" should return to todo at index 0, above "Second".
	m.undo()
	items = m.lists[todo].Items()
	if len(items) != 2 || items[0].(Task).Title() != "First" || items[1].(Task).Title() != "Second" {
		t.Fatalf("expected todo=[First, Second] after undos, got %v", items)
	}
}

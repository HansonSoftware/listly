package main

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

type failStore struct {
	Store
	err error
}

func (f failStore) SaveSession(int64, []Task) error { return f.err }

func TestDailySession_AutoSaves(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	id, _ := store.GetDailySession()

	m := newTestModel(NewTask(todo, "Daily task", ""))
	m.isDaily = true
	m.sessionID = id
	m.store = store

	m.MoveToNext()

	loaded, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Status() != completing {
		t.Fatalf("daily session did not autosave, got %v", loaded)
	}
}

func TestAutoSave_SurfacesErrors(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	id, _ := store.CreateSession("Err Test")

	m := newTestModel(NewTask(todo, "A", ""))
	m.isDaily = false
	m.sessionID = id
	m.store = failStore{Store: store, err: fmt.Errorf("disk full")}

	m.autoSave()
	if m.err == nil || m.err.Error() != "disk full" {
		t.Fatalf("expected m.err to surface save failure, got %v", m.err)
	}

	// A successful save clears the error.
	m.store = store
	m.autoSave()
	if m.err != nil {
		t.Fatalf("expected m.err cleared after successful save, got %v", m.err)
	}
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

func TestDeleteTask_UndoRestoresTaskAtIndex(t *testing.T) {
	first := NewTask(todo, "First", "")
	second := NewTask(todo, "Second", "")
	m := newTestModel(first, second)

	// Select and delete the first task.
	m.lists[todo].Select(0)
	m.DeleteTask()

	items := m.lists[todo].Items()
	if len(items) != 1 || items[0].(Task).Title() != "Second" {
		t.Fatalf("expected todo=[Second], got %v", items)
	}

	m.undo()
	items = m.lists[todo].Items()
	if len(items) != 2 || items[0].(Task).Title() != "First" || items[1].(Task).Title() != "Second" {
		t.Fatalf("expected todo=[First, Second] after undo, got %v", items)
	}
}

func TestTaskMsg_UndoRemovesInsertedTask(t *testing.T) {
	existing := NewTask(todo, "Existing", "")
	m := newTestModel(existing)

	// Simulate the form emitting a new Task.
	newTask := NewTask(todo, "Fresh", "desc")
	msg := tea.Msg(newTask)
	updated, _ := m.Update(msg.(Task))
	*m = updated.(Model)

	items := m.lists[todo].Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(items))
	}

	m.undo()
	items = m.lists[todo].Items()
	if len(items) != 1 || items[0].(Task).Title() != "Existing" {
		t.Fatalf("expected only 'Existing' after undo, got %v", items)
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

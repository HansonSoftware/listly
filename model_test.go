package main

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func newTestModel(tasks ...Task) *Model {
	m := &Model{
		focused:      todo,
		isDaily:      true, // skip store persistence in tests
		undoStack:    NewUndoStack(50),
		lists:        make([]list.Model, 3),
		sessionsList: list.New(nil, list.NewDefaultDelegate(), 44, 8),
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

func TestWelcomeMode_EnterOpensSession(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	m := newTestModel()
	m.store = store
	m.sessions = []Session{{ID: 1, Name: "S", IsDaily: false}}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(*Model)

	if got.mode != normal {
		t.Errorf("mode = %v, want %v (normal)", got.mode, normal)
	}
	if got.sessionID != 1 {
		t.Errorf("sessionID = %d, want 1", got.sessionID)
	}
	if got.isDaily {
		t.Errorf("isDaily = true, want false")
	}
}

func TestUpdate_KeyBeforeWindowSizeDoesNotPanic(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	m := New(store)
	// No WindowSizeMsg yet: m.loaded is false. A stray keypress must not
	// panic (regression: v2 panics updating a zero-value list).
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Update panicked on key before window size: %v", r)
		}
	}()
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	// Reaching here without a panic is the assertion.
}

func TestMainView_HelpBarVisible(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	m := newTestModel(NewTask(todo, "A", ""))
	m.store = store
	m.mode = normal

	m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	content := m.View().Content
	if !strings.Contains(content, "?: keybinds") {
		t.Errorf("help bar missing from mainView")
	}
	lines := strings.Count(content, "\n") + 1
	if lines > 30 {
		t.Errorf("mainView produced %d lines, exceeds terminal height 30", lines)
	}
}

func TestNormalMode_EscReturnsToWelcome(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	id, _ := store.CreateSession("S")

	m := newTestModel()
	m.store = store
	m.sessionID = id
	m.loaded = true
	m.mode = normal

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	u := updated.(*Model)
	if u.mode != welcome {
		t.Fatalf("mode = %v after esc, want welcome", u.mode)
	}
}

func TestCreationMode_EscReturnsToNormal(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	id, _ := store.CreateSession("Esc Test")

	m := newTestModel(NewTask(todo, "A", ""))
	m.store = store
	m.sessionID = id
	m.isDaily = false
	m.loaded = true
	m.mode = normal

	// Open the new-task form.
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"}); cmd == nil {
		t.Fatal("expected blink cmd from form init")
	}
	if m.mode != creation {
		t.Fatalf("mode = %v, want creation", m.mode)
	}

	// Esc must cancel back to normal.
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.mode != normal {
		t.Fatalf("mode = %v after esc, want normal", m.mode)
	}
}

func TestWelcomeMode_DeleteSessionConfirm(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()
	idA, _ := store.CreateSession("A")
	idB, _ := store.CreateSession("B")

	m := New(store)
	m.Update(sessionsLoadedMsg{[]Session{{ID: idA, Name: "A"}, {ID: idB, Name: "B"}}})

	// First x opens the confirm dialog; nothing deleted yet.
	m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if m.confirm == nil {
		t.Fatal("expected confirm dialog after first x")
	}
	if n := sessionCount(t, store); n != 2 {
		t.Fatalf("expected 2 sessions, got %d", n)
	}

	// A different key cancels the dialog.
	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.confirm != nil {
		t.Fatal("expected confirm dialog cleared by other key")
	}

	// Arm again, then x inside the dialog confirms: session deleted.
	m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	u := updated.(*Model)
	if u.confirm != nil {
		t.Fatal("expected confirm dialog cleared after confirm")
	}
	if n := sessionCount(t, store); n != 1 {
		t.Fatalf("expected 1 session after delete, got %d", n)
	}
	if cmd == nil {
		t.Fatal("expected reload cmd after delete")
	}
	msg := cmd()
	m.Update(msg)
	if len(m.sessions) != 1 {
		t.Fatalf("expected sessions reloaded to 1, got %d", len(m.sessions))
	}
}

func sessionCount(t *testing.T, s Store) int {
	t.Helper()
	sessions, err := s.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	return len(sessions)
}

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
	_ = updated

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

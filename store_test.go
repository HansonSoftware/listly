package main

import (
	"testing"
	"time"
)

func TestSQLiteStore_CreateAndListSessions(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	// Create a session
	id, err := store.CreateSession("Test Session")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateSession returned invalid ID: %d", id)
	}

	// List sessions
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Name != "Test Session" {
		t.Errorf("Expected name 'Test Session', got '%s'", sessions[0].Name)
	}
}

func TestSQLiteStore_DailySession(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id1, err := store.GetOrCreateDailySession()
	if err != nil {
		t.Fatalf("GetOrCreateDailySession failed: %v", err)
	}

	id2, err := store.GetOrCreateDailySession()
	if err != nil {
		t.Fatalf("GetOrCreateDailySession (2nd) failed: %v", err)
	}

	if id1.ID != id2.ID {
		t.Errorf("GetOrCreateDailySession returned different IDs: %d != %d", id1.ID, id2.ID)
	}
	if id1.Name != DailySessionName(time.Now()) {
		t.Errorf("unexpected daily session name: %q", id1.Name)
	}
}

func TestSQLiteStore_DailySession_IsNormalSession(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	s, err := store.GetOrCreateDailySession()
	if err != nil {
		t.Fatalf("GetOrCreateDailySession failed: %v", err)
	}
	if err := store.SaveSession(s.ID, []Task{NewTask(todo, "Today", "")}); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Reopening today keeps tasks; it behaves like any other session.
	again, err := store.GetOrCreateDailySession()
	if err != nil {
		t.Fatalf("GetOrCreateDailySession failed: %v", err)
	}
	if again.ID != s.ID {
		t.Errorf("expected same session id, got %d != %d", again.ID, s.ID)
	}
	tasks, err := store.LoadSession(s.ID)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title() != "Today" {
		t.Errorf("daily session tasks should persist, got %v", tasks)
	}
}

func TestSQLiteStore_SaveAndLoadSession(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id, err := store.CreateSession("Save Test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	tasks := []Task{
		NewTask(todo, "Task 1", "Desc 1"),
		NewTask(completing, "Task 2", "Desc 2"),
		NewTask(done, "Task 3", "Desc 3"),
	}

	if err := store.SaveSession(id, tasks); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(loaded))
	}
	if loaded[0].Title() != "Task 1" {
		t.Errorf("Expected 'Task 1', got '%s'", loaded[0].Title())
	}
	if loaded[1].Status() != completing {
		t.Errorf("Expected completing status, got %d", loaded[1].Status())
	}
}

func TestSQLiteStore_SavePreservesTaskIDs(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id, _ := store.CreateSession("ID Test")
	tasks := []Task{NewTask(todo, "A", ""), NewTask(done, "B", "")}
	if err := store.SaveSession(id, tasks); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	first, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if first[0].id == 0 || first[1].id == 0 {
		t.Fatalf("loaded tasks should have ids, got %d, %d", first[0].id, first[1].id)
	}

	// Re-saving the loaded tasks must keep their ids.
	if err := store.SaveSession(id, first); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}
	second, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(second) != 2 || second[0].id != first[0].id || second[1].id != first[1].id {
		t.Errorf("ids changed across re-save: first=%v,%v second=%v,%v",
			first[0].id, first[1].id, second[0].id, second[1].id)
	}
}

func TestSQLiteStore_SaveAppendsNewWithFreshIDs(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id, _ := store.CreateSession("Append Test")
	store.SaveSession(id, []Task{NewTask(todo, "Existing", "")})
	loaded, _ := store.LoadSession(id)

	loaded = append(loaded, NewTask(todo, "New", ""))
	if err := store.SaveSession(id, loaded); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	after, _ := store.LoadSession(id)
	if len(after) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(after))
	}
	if after[0].id != loaded[0].id {
		t.Errorf("existing task id changed: %d != %d", after[0].id, loaded[0].id)
	}
	if after[1].id == 0 || after[1].id == after[0].id {
		t.Errorf("new task should get a fresh distinct id, got %d", after[1].id)
	}
}

func TestSQLiteStore_UpdateSessionName(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id, err := store.CreateSession("Old Name")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if err := store.UpdateSessionName(id, "New Name"); err != nil {
		t.Fatalf("UpdateSessionName failed: %v", err)
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 || sessions[0].Name != "New Name" {
		t.Errorf("Expected 'New Name', got '%s'", sessions[0].Name)
	}
}

func TestSQLiteStore_MemoryDBStableAcrossManyOps(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	// Exercise the pool heavily: every query path must hit the same
	// underlying in-memory database.
	for i := 0; i < 50; i++ {
		id, err := store.CreateSession("Session")
		if err != nil {
			t.Fatalf("CreateSession %d failed: %v", i, err)
		}
		if err := store.SaveSession(id, []Task{NewTask(todo, "T", "")}); err != nil {
			t.Fatalf("SaveSession %d failed: %v", i, err)
		}
		if _, err := store.LoadSession(id); err != nil {
			t.Fatalf("LoadSession %d failed: %v", i, err)
		}
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 50 {
		t.Errorf("Expected 50 sessions, got %d", len(sessions))
	}
}

func TestSQLiteStore_DeleteSession(t *testing.T) {
	store, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("NewStoreWithPath failed: %v", err)
	}
	defer store.Close()

	id, err := store.CreateSession("To Delete")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	tasks := []Task{NewTask(todo, "T1", ""), NewTask(done, "T2", "")}
	if err := store.SaveSession(id, tasks); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if err := store.DeleteSession(id); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after delete, got %d", len(sessions))
	}

	loaded, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession after delete failed: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Expected 0 tasks after session delete, got %d (orphaned rows)", len(loaded))
	}
}

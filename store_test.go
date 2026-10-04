package main

import (
	"testing"
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

	id1, err := store.GetDailySession()
	if err != nil {
		t.Fatalf("GetDailySession failed: %v", err)
	}

	id2, err := store.GetDailySession()
	if err != nil {
		t.Fatalf("GetDailySession (2nd) failed: %v", err)
	}

	if id1 != id2 {
		t.Errorf("GetDailySession returned different IDs: %d != %d", id1, id2)
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
}

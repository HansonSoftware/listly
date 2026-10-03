package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func getDBPath() string {
	var home string

	if runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	} else {
		home = os.Getenv("HOME")
	}

	// Store local DB in ~/.local/share/listly/listly.db
	return filepath.Join(home, ".local/share/listly", "listly.db")
}

func initializeDB() (*sql.DB, error) {
	dbPath := getDBPath()
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	/*
		lists:
			id: A unique identifier for each list.
			name: The name of the list (e.g., "December", "Daily").
			created_at: The timestamp when the list was created.
			is_daily: Flag indicating if this is a daily session (not auto-saved)
	*/
	createListsTable := `
			CREATE TABLE IF NOT EXISTS lists (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					is_daily INTEGER DEFAULT 0
			);
    `

	_, err = db.Exec(createListsTable)
	if err != nil {
		return nil, err
	}

	// Migration: add is_daily column if it doesn't exist
	_, err = db.Exec(`ALTER TABLE lists ADD COLUMN is_daily INTEGER DEFAULT 0`)
	// Ignore error if column already exists

	/*
		tasks:
			id: A unique identifier for each task.
			list_id: Foreign key linking the task to a specific list (relates to the lists table).
			status: The task's status (0: todo, 1: completing, 2: done).
			title: The title of the task.
			description: A description of the task.
	*/
	createTasksTable := `
		CREATE TABLE IF NOT EXISTS tasks (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				list_id INTEGER NOT NULL,               -- Foreign key to the lists table
				status INTEGER NOT NULL,                -- 0: todo, 1: completing, 2: done
				title TEXT NOT NULL,
				description TEXT,
				FOREIGN KEY (list_id) REFERENCES lists(id)
		);
	`

	_, err = db.Exec(createTasksTable)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Session represents a saved task list
type Session struct {
	ID        int64
	Name      string
	CreatedAt string
	IsDaily   bool
}

// Task represents a task in the database
type DBTask struct {
	ID          int64
	ListID      int64
	Status      int
	Title       string
	Description string
}

// ListSessions returns all saved sessions
func ListSessions() ([]Session, error) {
	db, err := initializeDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, name, created_at, is_daily FROM lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var isDaily int
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &isDaily); err != nil {
			return nil, err
		}
		s.IsDaily = isDaily == 1
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// CreateSession creates a new session and returns its ID
func CreateSession(name string) (int64, error) {
	db, err := initializeDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	result, err := db.Exec(`INSERT INTO lists (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetDailySession returns the daily session, creating it if it doesn't exist
func GetDailySession() (int64, error) {
	db, err := initializeDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Try to find existing daily session
	var id int64
	err = db.QueryRow(`SELECT id FROM lists WHERE is_daily = 1 LIMIT 1`).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create daily session
	result, err := db.Exec(`INSERT INTO lists (name, is_daily) VALUES (?, 1)`, "Daily")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SaveSession saves all tasks for a session (replaces existing tasks)
func SaveSession(listID int64, tasks []Task) error {
	db, err := initializeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing tasks for this list
	_, err = tx.Exec(`DELETE FROM tasks WHERE list_id = ?`, listID)
	if err != nil {
		return err
	}

	// Insert current tasks
	stmt, err := tx.Prepare(`INSERT INTO tasks (list_id, status, title, description) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, task := range tasks {
		_, err = stmt.Exec(listID, task.Status(), task.Title(), task.Description())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadSession loads all tasks for a session
func LoadSession(listID int64) ([]Task, error) {
	db, err := initializeDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, list_id, status, title, description FROM tasks WHERE list_id = ?`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t DBTask
		if err := rows.Scan(&t.ID, &t.ListID, &t.Status, &t.Title, &t.Description); err != nil {
			return nil, err
		}
		tasks = append(tasks, Task{
			status:      status(t.Status),
			title:       t.Title,
			description: t.Description,
		})
	}
	return tasks, rows.Err()
}

// UpdateSessionName updates the name of a session
func UpdateSessionName(listID int64, name string) error {
	db, err := initializeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE lists SET name = ? WHERE id = ?`, name, listID)
	return err
}

// DeleteSession deletes a session and its tasks
func DeleteSession(listID int64) error {
	db, err := initializeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM lists WHERE id = ?`, listID)
	return err
}

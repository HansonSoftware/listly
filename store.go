package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

type Store interface {
	ListSessions() ([]Session, error)
	CreateSession(name string) (int64, error)
	GetDailySession() (int64, error)
	SaveSession(listID int64, tasks []Task) error
	LoadSession(listID int64) ([]Task, error)
	UpdateSessionName(listID int64, name string) error
	DeleteSession(listID int64) error
	Close() error
}

type sqliteStore struct {
	db *sql.DB
}

func NewStore() (Store, error) {
	dbPath := getDBPath()
	return NewStoreWithPath(dbPath)
}

func NewStoreWithPath(dbPath string) (Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// TODO: Test behavior when multiple listly instances are running.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// SQLite disables FK enforcement per connection by default.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &sqliteStore{db: db}, nil
}

func createTables(db *sql.DB) error {
	listsTable := `
		CREATE TABLE IF NOT EXISTS lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_daily INTEGER DEFAULT 0
		);
	`
	if _, err := db.Exec(listsTable); err != nil {
		return err
	}

	// Migration: add is_daily column if missing
	// TODO(1.0.0): remove this migration — 1.0.0 users start with a fresh DB
	// where is_daily already exists in the CREATE TABLE above.
	db.Exec(`ALTER TABLE lists ADD COLUMN is_daily INTEGER DEFAULT 0`)

	tasksTable := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id INTEGER NOT NULL,
			status INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(tasksTable); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_list_id ON tasks(list_id)`); err != nil {
		return err
	}

	// One-time sweep of rows orphaned before DeleteSession cleaned up after
	// itself. TODO(1.0.0): remove this — fresh 1.0.0 DBs never have orphans.
	db.Exec(`DELETE FROM tasks WHERE list_id NOT IN (SELECT id FROM lists)`)

	return nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

func (s *sqliteStore) ListSessions() ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at, is_daily FROM lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		var isDaily int
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.CreatedAt, &isDaily); err != nil {
			return nil, err
		}
		sess.IsDaily = isDaily == 1
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func (s *sqliteStore) CreateSession(name string) (int64, error) {
	result, err := s.db.Exec(`INSERT INTO lists (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *sqliteStore) GetDailySession() (int64, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM lists WHERE is_daily = 1 LIMIT 1`).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := s.db.Exec(`INSERT INTO lists (name, is_daily) VALUES (?, 1)`, "Daily")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *sqliteStore) SaveSession(listID int64, tasks []Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM tasks WHERE list_id = ?`, listID)
	if err != nil {
		return err
	}

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

func (s *sqliteStore) LoadSession(listID int64) ([]Task, error) {
	rows, err := s.db.Query(`SELECT id, list_id, status, title, description FROM tasks WHERE list_id = ?`, listID)
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

func (s *sqliteStore) UpdateSessionName(listID int64, name string) error {
	_, err := s.db.Exec(`UPDATE lists SET name = ? WHERE id = ?`, name, listID)
	return err
}

func (s *sqliteStore) DeleteSession(listID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete tasks explicitly so cleanup does not depend on FK pragma state.
	if _, err := tx.Exec(`DELETE FROM tasks WHERE list_id = ?`, listID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM lists WHERE id = ?`, listID); err != nil {
		return err
	}
	return tx.Commit()
}

func getDBPath() string {
	var home string
	if runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	} else {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".local/share/listly", "listly.db")
}

type Session struct {
	ID        int64
	Name      string
	CreatedAt string
	IsDaily   bool
}

type DBTask struct {
	ID          int64
	ListID      int64
	Status      int
	Title       string
	Description string
}

var _ Store = (*sqliteStore)(nil) // Compile-time interface check

package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

type Store interface {
	ListSessions() ([]Session, error)
	CreateSession(name string) (int64, error)
	GetOrCreateDailySession() (Session, error)
	SaveSession(listID int64, tasks []Task) error
	LoadSession(listID int64) ([]Task, error)
	UpdateSessionName(listID int64, name string) error
	DeleteSession(listID int64) error
	Close() error
}

type Session struct {
	ID        int64
	Name      string
	CreatedAt string
}

func (s Session) FilterValue() string { return s.Name }

type DBTask struct {
	ID          int64
	ListID      int64
	Status      int
	Title       string
	Description string
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

	db, err := sql.Open("sqlite", dbPath)
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(listsTable); err != nil {
		return err
	}

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

	return nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

func (s *sqliteStore) ListSessions() ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.CreatedAt); err != nil {
			return nil, err
		}
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

// GetOrCreateDailySession returns today's daily session, named like
// "Oct 05 2026 TODO". Daily notes are ordinary sessions in the DB — no
// special columns — so they autosave and behave like everything else.
// Pressing "d" again on the same day reopens the same note; a new day
// produces a new one (created lazily, only when "d" is pressed).
func (s *sqliteStore) GetOrCreateDailySession() (Session, error) {
	name := DailySessionName(time.Now())
	var sess Session
	err := s.db.QueryRow(`SELECT id, name, created_at FROM lists WHERE name = ? LIMIT 1`, name).
		Scan(&sess.ID, &sess.Name, &sess.CreatedAt)
	if err == nil {
		return sess, nil
	}
	if err != sql.ErrNoRows {
		return Session{}, err
	}
	id, err := s.CreateSession(name)
	if err != nil {
		return Session{}, err
	}
	return Session{ID: id, Name: name}, nil
}

// DailySessionName is the canonical name of the daily note for t.
func DailySessionName(t time.Time) string {
	return t.Format("Jan 02 2006") + " TODO"
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

	insertWithID, err := tx.Prepare(`INSERT INTO tasks (id, list_id, status, title, description) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insertWithID.Close()

	insertNew, err := tx.Prepare(`INSERT INTO tasks (list_id, status, title, description) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insertNew.Close()

	for _, task := range tasks {
		if task.id > 0 {
			_, err = insertWithID.Exec(task.id, listID, task.Status(), task.Title(), task.Description())
		} else {
			_, err = insertNew.Exec(listID, task.Status(), task.Title(), task.Description())
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *sqliteStore) LoadSession(listID int64) ([]Task, error) {
	rows, err := s.db.Query(`SELECT id, list_id, status, title, description FROM tasks WHERE list_id = ? ORDER BY id`, listID)
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
			id:          t.ID,
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
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "listly", "listly.db")
		}
		if dir := os.Getenv("APPDATA"); dir != "" {
			return filepath.Join(dir, "listly", "listly.db")
		}
		return filepath.Join(home, "AppData", "Local", "listly", "listly.db")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "listly", "listly.db")
	default:
		return filepath.Join(home, ".local", "share", "listly", "listly.db")
	}
}

var _ Store = (*sqliteStore)(nil) // Compile-time interface check

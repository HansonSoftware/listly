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

	return filepath.Join(home, ".listly", "listly.db")
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
	*/
	createListsTable := `
		-- Table to store tasks for each list
		CREATE TABLE IF NOT EXISTS tasks (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				list_id INTEGER NOT NULL,               -- Foreign key to the lists table
				status INTEGER NOT NULL,                -- 0: todo, 1: completing, 2: done
				title TEXT NOT NULL,
				description TEXT,
				FOREIGN KEY (list_id) REFERENCES lists(id)
		);
	`

	_, err = db.Exec(createListsTable)
	if err != nil {
		return nil, err
	}

	/*
		tasks:
			id: A unique identifier for each task.
			list_id: Foreign key linking the task to a specific list (relates to the lists table).
			status: The task's status (0: todo, 1: completing, 2: done).
			title: The title of the task.
			description: A description of the task.
	*/
	createTasksTable := `
			CREATE TABLE IF NOT EXISTS lists (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
    `

	_, err = db.Exec(createTasksTable)
	if err != nil {
		return nil, err
	}

	return db, nil
}

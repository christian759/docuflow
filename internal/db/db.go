package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	dbPath := "./docuflow.db"

	// Ensure file exists or can be created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := InitSchema(db); err != nil {
		return nil, err
	}

	log.Println("Connected to SQLite database at", dbPath)
	return db, nil
}

func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'editor',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		owner_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(owner_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS revisions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		document_id INTEGER,
		content TEXT,
		editor_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		change_summary TEXT,
		FOREIGN KEY(document_id) REFERENCES documents(id),
		FOREIGN KEY(editor_id) REFERENCES users(id)
	);
	`
	_, err := db.Exec(schema)
	return err
}

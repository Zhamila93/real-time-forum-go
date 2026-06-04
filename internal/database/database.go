package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	pathToFile := filepath.Join("migrations", "schema.sql")
	file, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(string(file))
	if err != nil {
		return nil, err
	}
	return db, nil
}

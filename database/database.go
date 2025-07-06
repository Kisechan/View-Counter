package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	mu  sync.Mutex
)

func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func GetDB() *sql.DB {
	return db
}

func GetMutex() *sync.Mutex {
	return &mu
}
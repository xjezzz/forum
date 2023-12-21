package internal

import (
	"database/sql"
	"forum-project/config"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// New function returns new database
func New(config *config.Config) (*sql.DB, error) {
	_, err := os.Create(config.StoragePath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(config.DriverName, config.StoragePath)
	if err != nil {
		return nil, err
	}
	// database migrations
	init, err := os.ReadFile(config.InitSQL)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(init))
	if err != nil {
		return nil, err
	}

	mock, err := os.ReadFile(config.MockDataSQL)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(mock))
	if err != nil {
		return nil, err
	}

	return db, nil
}

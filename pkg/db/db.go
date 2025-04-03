package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// New returns a new database storage layer.
func New() (*DB, error) {
	return new(DB), nil
}

// Connect sets up the database connection
func (db *DB) Connect(file string) error {
	d, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return err
	}
	db.d = d
	return nil
}

// Migrate reconciles the database schema with the
func (db *DB) Migrate() error {
	tables := []interface{}{
		&types.Account{},
	}

	if err := db.d.AutoMigrate(tables...); err != nil {
		return err
	}

	return nil
}

// Raw provides a handle to the underlying gorm instance for when the
// wrapped queries are insufficient.
func (db *DB) Raw() *gorm.DB {
	return db.d
}

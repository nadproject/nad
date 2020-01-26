package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	// DialectPostgres represenets a SQL dialect of Postgres
	DialectPostgres = "postgres"
)

// First executes the given gorm.DB query and saves the first result
// in the given destination.
func First(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

// Find executes the given gorm.DB query and saves the slice of result
// in the given destination.
func Find(db *gorm.DB, dst interface{}) error {
	err := db.Find(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

// Model is the base for the database model definitions.
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

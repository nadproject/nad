package models

import (
	"github.com/jinzhu/gorm"
)

// First executes the given gorm.DB query and saves the rersult
// in the given destination.
func First(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

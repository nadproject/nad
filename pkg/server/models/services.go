package models

import (
	"github.com/jinzhu/gorm"

	"github.com/nadproject/nad/pkg/server/migrations"
	"github.com/pkg/errors"
	// use postgres
	_ "github.com/lib/pq"
)

// Services encapsulates the services that are used to interact with the
// database.
type Services struct {
	User    UserService
	Session SessionService
	Note    NoteService
	Book    BookService
	DB      *gorm.DB
}

// NewServices instantiates a new Services.
func NewServices(dialect, connStr string) (*Services, error) {
	var s Services

	DB, err := gorm.Open(dialect, connStr)
	if err != nil {
		return nil, err
	}

	s.DB = DB
	s.Book = NewBookService(s.DB)
	s.Note = NewNoteService(s.DB)
	s.User = NewUserService(s.DB)
	s.Session = NewSessionService(s.DB)

	return &s, nil
}

// Close closes the database connection of the service.
func (s *Services) Close() error {
	return s.DB.Close()
}

// InitDB initializes the database for use by applying the model definitions and running
// any pending migrations.
func (s *Services) InitDB() error {
	if err := s.SetupDB(); err != nil {
		return err
	}

	if err := s.MigrateDB(); err != nil {
		return err
	}

	return nil
}

// SetupDB rutomatically applies a set of model definitions to the tables. It also
// creates any extensions if missing.
func (s *Services) SetupDB() error {
	if err := s.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return errors.Wrap(err, "creating uuid extension")
	}

	err := s.DB.AutoMigrate(&User{}, &Note{}, &Book{}, &Session{}).Error
	if err != nil {
		return errors.Wrap(err, "updating schema")
	}

	return nil
}

// MigrateDB runs migrations.
func (s *Services) MigrateDB() error {
	err := migrations.Run(s.DB)
	if err != nil {
		return errors.Wrap(err, "running migrations")
	}

	return nil
}

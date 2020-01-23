package models

import (
	"github.com/jinzhu/gorm"

	"github.com/nadproject/nad/pkg/server/migrations"
	"github.com/pkg/errors"
	// use postgres
	_ "github.com/lib/pq"
)

// ServicesConfig configures the given service by mutating it.
type ServicesConfig func(*Services) error

// WithGorm returns a service configuration procedure that initializes
// a database connection and saves it into the given service.
func WithGorm(dialect, connStr string) ServicesConfig {
	return func(s *Services) error {
		DB, err := gorm.Open(dialect, connStr)
		if err != nil {
			return err
		}

		s.DB = DB
		return nil
	}
}

// WithUser returns a service configuration procedure that configures
// a user service.
func WithUser() ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.DB)
		return nil
	}
}

// WithSession returns a service configuration procedure that configures
// a session service.
func WithSession() ServicesConfig {
	return func(s *Services) error {
		s.Session = NewSessionService(s.DB)
		return nil
	}
}

// WithNote returns a service configuration procedure that configures
// a note service.
func WithNote() ServicesConfig {
	return func(s *Services) error {
		s.Note = NewNoteService(s.DB)
		return nil
	}
}

// WithBook returns a service configuration procedure that configures
// a book service.
func WithBook() ServicesConfig {
	return func(s *Services) error {
		s.Book = NewBookService(s.DB)
		return nil
	}
}

// NewServices instantiates a new Services by using the given slice of
// service configuration procedures.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services

	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

// Services encapsulates the services that are used to interact with the
// database.
type Services struct {
	User    UserService
	Session SessionService
	Note    NoteService
	Book    BookService
	DB      *gorm.DB
}

// Close closes the database connection of the service.
func (s *Services) Close() error {
	return s.DB.Close()
}

// InitDB automatically migrates all tables using a set of model
// definitions.
func (s *Services) InitDB() error {
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

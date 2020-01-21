package models

import (
	"github.com/jinzhu/gorm"

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

// AutoMigrate automatically migrates all tables using a set of model
// definitions.
func (s *Services) AutoMigrate() error {
	return s.DB.AutoMigrate(&User{}, &Session{}).Error
}

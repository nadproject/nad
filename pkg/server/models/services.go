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
func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}

		s.db = db
		return nil
	}
}

// WithUser returns a service configuration procedure that configures
// a user service.
func WithUser() ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db)
		return nil
	}
}

// WithSession returns a service configuration procedure that configures
// a user service.
func WithSession() ServicesConfig {
	return func(s *Services) error {
		s.Session = NewSessionService(s.db)
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
	db      *gorm.DB
}

// Close closes the database connection of the service.
func (s *Services) Close() error {
	return s.db.Close()
}

// AutoMigrate automatically migrates all tables using a set of model
// definitions.
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Session{}).Error
}

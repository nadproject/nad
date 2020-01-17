package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Session represents a user session
type Session struct {
	gorm.Model
	UserID     uint   `gorm:"index"`
	Key        string `gorm:"index"`
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

// SessionDB is an interface for database operations
// related to sessions.
type SessionDB interface {
	ByKey(key string) (*Session, error)
}

// sessionGorm encapsulates the actual implementations of
// the database operations involving sessions.
type sessionGorm struct {
	db *gorm.DB
}

// SessionService is a set of methods for interacting with the session model
type SessionService interface {
	SessionDB
}

type sessionService struct {
	SessionDB
}

// NewSessionService returns a new sessionService
func NewSessionService(db *gorm.DB) SessionService {
	sg := &sessionGorm{db}
	sv := newSessionValidator(sg)

	return &sessionService{
		SessionDB: sv,
	}
}

type sessionValidator struct {
	SessionDB
}

func newSessionValidator(sdb SessionDB) *sessionValidator {
	return &sessionValidator{
		SessionDB: sdb,
	}
}

// ByKey looks up a session with the given key.
func (sg *sessionGorm) ByKey(key string) (*Session, error) {
	var ret Session
	err := First(sg.db.Where("key = ?", key), &ret)

	return &ret, err
}

type sessionValFunc func(*Session) error

func runSessionValFuncs(session *Session, fns ...sessionValFunc) error {
	for _, fn := range fns {
		if err := fn(session); err != nil {
			return err
		}
	}
	return nil
}

// ByKey validates the parameters for retreiving a sesison by key.
func (sv *sessionValidator) ByKey(key string) (*Session, error) {
	s := Session{
		Key: key,
	}
	if err := runSessionValFuncs(&s, sv.requireKey); err != nil {
		return nil, err
	}

	return sv.SessionDB.ByKey(key)
}

func (sv *sessionValidator) requireKey(s *Session) error {
	if s.Key == "" {
		return ErrSessionKeyRequired
	}

	return nil
}

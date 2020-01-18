package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/server/crypt"
	"github.com/pkg/errors"
)

// Session represents a user session
type Session struct {
	Model
	UserID     uint   `gorm:"index"`
	Key        string `gorm:"index"`
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

// SessionDB is an interface for database operations
// related to sessions.
type SessionDB interface {
	ByKey(key string) (*Session, error)

	Create(*Session) error
	Delete(key string) error
}

// sessionGorm encapsulates the actual implementations of
// the database operations involving sessions.
type sessionGorm struct {
	db *gorm.DB
}

// SessionService is a set of methods for interacting with the session model
type SessionService interface {
	SessionDB
	Login(userID uint) (*Session, error)
}

type sessionService struct {
	SessionDB
}

func (ss sessionService) Login(userID uint) (*Session, error) {
	key, err := crypt.GetRandomStr(32)
	if err != nil {
		return nil, errors.Wrap(err, "generating key")
	}

	session := Session{
		UserID:     userID,
		Key:        key,
		LastUsedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * 100 * time.Hour),
	}

	if err := ss.SessionDB.Create(&session); err != nil {
		return nil, errors.Wrap(err, "saving session")
	}

	return &session, nil
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

func (sg *sessionGorm) Delete(key string) error {
	if err := sg.db.Where("key = ?", key).Delete(&Session{}).Error; err != nil {
		return err
	}

	return nil
}

func (sg *sessionGorm) Create(s *Session) error {
	if err := sg.db.Debug().Save(s).Error; err != nil {
		return errors.Wrap(err, "saving session")
	}

	return nil
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

// Create validates the parameters for retreiving a sesison by key.
func (sv *sessionValidator) Create(s *Session) error {
	if err := runSessionValFuncs(s, sv.requireKey, sv.requireUserID); err != nil {
		return err
	}

	return sv.SessionDB.Create(s)
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

func (sv *sessionValidator) requireUserID(s *Session) error {
	if s.UserID == 0 {
		return ErrSessionKeyRequired
	}

	return nil
}

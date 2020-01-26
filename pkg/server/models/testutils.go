package models

import (
	"encoding/base64"
	"math/rand"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// SetupUser creates and returns a new user for testing purposes
func SetupUser(t *testing.T, us UserService, ss SessionService, email, password string) (User, Session) {
	user := User{
		Email:    email,
		Password: password,
		Pro:      true,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(errors.Wrap(err, "hashing password"))
	}
	user.Password = string(hashedPassword)

	if err := us.Create(&user); err != nil {
		t.Fatal(errors.Wrap(err, "preparing user"))
	}

	session := SetupSession(t, ss, user.ID)

	return user, session
}

// SetupSession creates and returns a new user session
func SetupSession(t *testing.T, ss SessionService, userID uint) Session {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(errors.Wrap(err, "reading random bits"))
	}

	session := Session{
		Key:       base64.StdEncoding.EncodeToString(b),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	if err := ss.Create(&session); err != nil {
		t.Fatal(errors.Wrap(err, "Failed to prepare session"))
	}

	return session
}

// ClearTestData deletes all records from the database
func ClearTestData(t *testing.T, db *gorm.DB) {
	if err := db.Delete(&Book{}).Error; err != nil {
		t.Fatal(errors.Wrap(err, "Failed to clear books"))
	}
	if err := db.Delete(&Note{}).Error; err != nil {
		t.Fatal(errors.Wrap(err, "Failed to clear notes"))
	}
	if err := db.Delete(&User{}).Error; err != nil {
		t.Fatal(errors.Wrap(err, "Failed to clear users"))
	}
	if err := db.Delete(&Session{}).Error; err != nil {
		t.Fatal(errors.Wrap(err, "Failed to clear sessions"))
	}
}

// MustExec fails the test if the given database query has error
func MustExec(t *testing.T, db *gorm.DB, message string) {
	if err := db.Error; err != nil {
		t.Fatalf("%s: %s", message, err.Error())
	}
}

// TestServices is the service for test
var TestServices *Services

// InitTestService initializes test service
func InitTestService(cfg config.Config) error {
	services, err := NewServices(DialectPostgres, cfg.DB.GetConnectionStr())
	if err != nil {
		return err
	}

	err = services.SetupDB()
	if err != nil {
		return err
	}

	TestServices = services

	return nil
}

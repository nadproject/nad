package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User is a user model
type User struct {
	gorm.Model
	UUID             string     `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	StripeCustomerID string     `json:"-"`
	BillingCountry   string     `json:"-"`
	LastLoginAt      *time.Time `json:"-"`
	MaxUSN           int        `json:"-" gorm:"default:0"`
	Pro              bool       `json:"-" gorm:"default:false"`
	Email            string     `json:"-" gorm:"index"`
	Password         string     `gorm:"-" json:"-"`
	PasswordHash     string     `json:"-"`
	EmailVerified    bool       `gorm:"default:false"`
}

// UserDB is an interface for database operations
// related to users.
type UserDB interface {
	ByID(id uint) (*User, error)
	ByUUID(uuid string) (*User, error)
	ByEmail(email string) (*User, error)
	BySession(token string) (*User, error)

	Create(user *User) error
	Update(user *User) error
}

// UserService is a set of methods for interacting with the user model
type UserService interface {
	// Authenticate verifies that the provided email and password combination
	// matches a user.
	Authenticate(email, password string) (*User, error)
	UserDB
}

// NewUserService returns a new userService
func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{db}
	uv := newUserValidator(ug)

	return &userService{
		UserDB: uv,
	}
}

type userService struct {
	UserDB
}

// Authenticate authenticates a user with the given email and password.
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrLoginInvalid
		default:
			return nil, err
		}
	}

	return foundUser, nil
}

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func newUserValidator(udb UserDB) *userValidator {
	return &userValidator{
		UserDB:     udb,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

type userValidator struct {
	UserDB
	emailRegex *regexp.Regexp
}

// ByID validates the params for ByID
func (uv *userValidator) ByID(id uint) (*User, error) {
	var user User
	user.ID = id
	err := runUserValFuncs(&user, uv.idValid())
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByID(id)
}

// ByEmail normalizes the given email address.
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}

// Update validates the given user for creation.
func (uv *userValidator) Create(user *User) error {
	err := runUserValFuncs(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

// Update validates the given user for update.
func (uv *userValidator) Update(user *User) error {
	err := runUserValFuncs(user,
		uv.passwordMinLength,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	// If same ID, it is an update
	if user.ID != existing.ID {
		return ErrEmailDuplicate
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}

func (uv *userValidator) idValid() userValFunc {
	return userValFunc(func(user *User) error {
		if user.ID <= 0 {
			return ErrIDInvalid
		}
		return nil
	})
}

// bcryptPassword hashes the given password for the user
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// userGorm encapsulates the actual implementations of
// the database operations involving users.
type userGorm struct {
	db *gorm.DB
}

// ByID looks up the user with the given uuid
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	err := First(ug.db.Where("id = ?", id), &user)

	return &user, err
}

// ByUUID looks up the user with the given uuid
func (ug *userGorm) ByUUID(uuid string) (*User, error) {
	var user User
	err := First(ug.db.Where("uuid = ?", uuid), &user)

	return &user, err
}

// ByEmail looks up a user with the given email address and returns that user.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	err := First(ug.db.Where("email = ?", email), &user)

	return &user, err
}

// BySessionKey looks up a user with the given session key.
func (ug *userGorm) BySession(sessionKey string) (*User, error) {
	var session Session
	err := First(ug.db.Where("key = ?", sessionKey), session)
	if err != nil {
		return nil, err
	}

	var user User
	err = First(ug.db.Where("id = ?", session.UserID), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Create creates the provided user.
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update will update the provided user with the provided data
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(&user).Error
}

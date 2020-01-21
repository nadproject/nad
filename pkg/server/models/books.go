package models

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Book is a model for a book
type Book struct {
	Model
	UUID      string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID    uint   `json:"user_id" gorm:"index"`
	Label     string `json:"label" gorm:"index"`
	Notes     []Note `json:"books" gorm:"foreignkey:book_uuid"`
	AddedOn   int64  `json:"added_on"`
	EditedOn  int64  `json:"edited_on"`
	USN       int    `json:"-" gorm:"index"`
	Deleted   bool   `json:"-" gorm:"default:false"`
	Encrypted bool   `json:"-" gorm:"default:false"`
}

// BookDB is an interface for database operations related to books.
type BookDB interface {
	Search(userID uint) ([]Book, error)
	ByUUID(uuid string) (*Book, error)

	Create(*Book, *gorm.DB) error
	// Update(*Book) error
	Delete(key string) error
}

// bookGorm encapsulates the actual implementations of
// the database operations involving books.
type bookGorm struct {
	db *gorm.DB
}

// BookService is a set of methods for interacting with the book model
type BookService interface {
	BookDB
}

type bookService struct {
	BookDB
}

// NewBookService returns a new bookService
func NewBookService(db *gorm.DB) BookService {
	ng := &bookGorm{db}
	nv := newBookValidator(ng)

	return &bookService{
		BookDB: nv,
	}
}

type bookValidator struct {
	BookDB
}

func newBookValidator(ndb BookDB) *bookValidator {
	return &bookValidator{
		BookDB: ndb,
	}
}

// Search looks up a book with the given key.
func (ng *bookGorm) Search(userID uint) ([]Book, error) {
	var ret []Book
	err := Find(ng.db.Debug().Where("user_id = ?", userID), &ret)

	return ret, err
}

// ByUUID looks up a book with the given key.
func (ng *bookGorm) ByUUID(uuid string) (*Book, error) {
	var ret Book
	err := First(ng.db.Where("uuid = ?", uuid), &ret)

	return &ret, err
}

// TODO
func (ng *bookGorm) Delete(key string) error {
	if err := ng.db.Where("key = ?", key).Delete(&Book{}).Error; err != nil {
		return err
	}

	return nil
}

func (ng *bookGorm) Create(n *Book, tx *gorm.DB) error {
	var conn *gorm.DB
	if tx != nil {
		conn = tx
	} else {
		conn = ng.db
	}

	if err := conn.Save(n).Error; err != nil {
		return errors.Wrap(err, "saving book")
	}

	return nil
}

type bookValFunc func(*Book) error

func runBookValFuncs(book *Book, fns ...bookValFunc) error {
	for _, fn := range fns {
		if err := fn(book); err != nil {
			return err
		}
	}
	return nil
}

// Create validates the parameters for retreiving a sesison by key.
func (nv *bookValidator) Create(s *Book, tx *gorm.DB) error {
	if err := runBookValFuncs(s,
		nv.requireUserID,
		nv.requireAddedOn,
		nv.requireUSN,
	); err != nil {
		return err
	}

	return nv.BookDB.Create(s, tx)
}

// Search validates the parameters for retreiving a sesison by key.
func (nv *bookValidator) Search(userID uint) ([]Book, error) {
	s := Book{
		UserID: userID,
	}
	if err := runBookValFuncs(&s, nv.requireUserID); err != nil {
		return nil, err
	}

	return nv.BookDB.Search(userID)
}

// ByUUID validates the parameters for retreiving a sesison by key.
func (nv *bookValidator) ByUUID(uuid string) (*Book, error) {
	s := Book{
		UUID: uuid,
	}
	if err := runBookValFuncs(&s, nv.requireUUID); err != nil {
		return nil, err
	}

	return nv.BookDB.ByUUID(uuid)
}

func (nv *bookValidator) requireUUID(s *Book) error {
	if s.UUID == "" {
		return ErrBookUUIDRequired
	}

	return nil
}

func (nv *bookValidator) requireUserID(s *Book) error {
	if s.UserID == 0 {
		return ErrBookUserIDRequired
	}

	return nil
}

func (nv *bookValidator) requireAddedOn(s *Book) error {
	if s.AddedOn == 0 {
		return ErrBookAddedOnRequired
	}

	return nil
}

func (nv *bookValidator) requireEditedOn(s *Book) error {
	if s.EditedOn == 0 {
		return ErrBookEditedOnRequired
	}

	return nil
}

func (nv *bookValidator) requireUSN(s *Book) error {
	if s.AddedOn == 0 {
		return ErrBookUSNRequired
	}

	return nil
}

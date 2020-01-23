package models

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Book is a model for a book
type Book struct {
	Model
	UUID      string `gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID    uint   `gorm:"index"`
	Name      string `gorm:"index"`
	Notes     []Note `gorm:"foreignkey:book_uuid"`
	USN       int    `gorm:"index"`
	Deleted   bool   `gorm:"default:false"`
	Encrypted bool   `gorm:"default:false"`
	AddedOn   int64
	EditedOn  int64
}

// BookDB is an interface for database operations related to bookb.
type BookDB interface {
	Search(p BookSearchParams) ([]Book, error)
	ByUUID(uuid string) (*Book, error)
	ByName(userID uint, name string) (*Book, error)
	ByUSNRange(userID uint, lb, ub, limit int) ([]Book, error)

	Create(*Book, *gorm.DB) error
	Update(*Book, *gorm.DB) error
}

// bookGorm encapsulates the actual implementations of
// the database operations involving bookb.
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
	bg := &bookGorm{db}
	bv := newBookValidator(bg)

	return &bookService{
		BookDB: bv,
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

// BookSearchParams is a group of paramters for searching books
type BookSearchParams struct {
	Name   string
	UserID uint
	Offset int
	Limit  int
}

// Search looks up books with the given params
func (bg *bookGorm) Search(p BookSearchParams) ([]Book, error) {
	var ret []Book

	conn := bg.db.Where("user_id = ? AND NOT deleted", p.UserID)

	if p.Name != "" {
		part := fmt.Sprintf("%%%s%%", p.Name)
		conn = conn.Where("LOWER(name) LIKE ?", part)
	}

	err := Find(conn.Order("name ASC"), &ret)

	return ret, err
}

// ByName looks up a book with the given name.
func (bg *bookGorm) ByName(userID uint, name string) (*Book, error) {
	var ret Book
	err := First(bg.db.Where("user_id = ? AND name = ?", userID, name), &ret)

	return &ret, err
}

func (bg *bookGorm) ByUSNRange(userID uint, lb, ub, limit int) ([]Book, error) {
	var ret []Book
	err := Find(bg.db.Where("user_id = ? AND usn > ? AND usn <= ?", userID, lb, ub).Order("usn ASC").Limit(limit), &ret)

	return ret, err
}

// ByUUID looks up a book with the given key.
func (bg *bookGorm) ByUUID(uuid string) (*Book, error) {
	var ret Book
	err := First(bg.db.Where("uuid = ?", uuid), &ret)

	return &ret, err
}

func (bg *bookGorm) Update(b *Book, tx *gorm.DB) error {
	var conn *gorm.DB
	if tx != nil {
		conn = tx
	} else {
		conn = bg.db
	}

	if err := conn.Save(b).Error; err != nil {
		return errors.Wrap(err, "saving note")
	}

	return nil
}

func (bg *bookGorm) Create(n *Book, tx *gorm.DB) error {
	var conn *gorm.DB
	if tx != nil {
		conn = tx
	} else {
		conn = bg.db
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

// Create validates the parameters for create.
func (bv *bookValidator) Create(b *Book, tx *gorm.DB) error {
	if err := runBookValFuncs(b,
		bv.ensureNameUnique,
		bv.requireUserID,
		bv.requireUSN,
	); err != nil {
		log.Printf("err %+v", err)
		return err
	}

	return bv.BookDB.Create(b, tx)
}

// Update validates the parameters for update.
func (bv *bookValidator) Update(b *Book, tx *gorm.DB) error {
	if err := runBookValFuncs(b,
		bv.ensureNameUnique,
		bv.requireUserID,
		bv.requireUSN,
	); err != nil {
		return err
	}

	return bv.BookDB.Update(b, tx)
}

// ByUUID validates the parameters for retreiving a sesison by key.
func (bv *bookValidator) ByUUID(uuid string) (*Book, error) {
	b := Book{
		UUID: uuid,
	}
	if err := runBookValFuncs(&b, bv.requireUUID); err != nil {
		return nil, err
	}

	return bv.BookDB.ByUUID(uuid)
}

func (bv *bookValidator) requireUUID(b *Book) error {
	if b.UUID == "" {
		return ErrBookUUIDRequired
	}

	return nil
}

func (bv *bookValidator) requireUserID(b *Book) error {
	if b.UserID == 0 {
		return ErrBookUserIDRequired
	}

	return nil
}

func (bv *bookValidator) requireEditedOn(b *Book) error {
	if b.EditedOn == 0 {
		return ErrBookEditedOnRequired
	}

	return nil
}

func (bv *bookValidator) ensureNameUnique(b *Book) error {
	// An empty book name should be considered valid. Deleted books
	// have an empty string as a name.
	if b.Name == "" {
		return nil
	}

	_, err := bv.BookDB.ByName(b.UserID, b.Name)
	if err == ErrNotFound {
		return nil
	}

	return ErrBookNameTaken
}

func (bv *bookValidator) requireUSN(b *Book) error {
	if b.USN == 0 {
		return ErrBookUSNRequired
	}

	return nil
}

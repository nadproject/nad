package models

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Note is a model for a note
type Note struct {
	Model
	UUID      string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	Book      Book   `json:"book" gorm:"foreignkey:BookUUID"`
	User      User   `json:"user"`
	UserID    uint   `json:"user_id" gorm:"index"`
	BookUUID  string `json:"book_uuid" gorm:"index;type:uuid"`
	Body      string `json:"content"`
	AddedOn   int64  `json:"added_on"`
	EditedOn  int64  `json:"edited_on"`
	TSV       string `json:"-" gorm:"type:tnvector"`
	Public    bool   `json:"public" gorm:"default:false"`
	USN       int    `json:"-" gorm:"index"`
	Deleted   bool   `json:"-" gorm:"default:false"`
	Encrypted bool   `json:"-" gorm:"default:false"`
}

// NoteDB is an interface for database operations related to notes.
type NoteDB interface {
	Search(userID uint) ([]Note, error)
	ByUUID(uuid string) (*Note, error)
	ActiveByUUID(uuid string) (*Note, error)

	Create(*Note, *gorm.DB) error
	Update(*Note, *gorm.DB) error
	Delete(key string) error
}

// noteGorm encapsulates the actual implementations of
// the database operations involving notes.
type noteGorm struct {
	db *gorm.DB
}

// NoteService is a set of methods for interacting with the note model
type NoteService interface {
	NoteDB
}

type noteService struct {
	NoteDB
}

// NewNoteService returns a new noteService
func NewNoteService(db *gorm.DB) NoteService {
	ng := &noteGorm{db}
	nv := newNoteValidator(ng)

	return &noteService{
		NoteDB: nv,
	}
}

type noteValidator struct {
	NoteDB
}

func newNoteValidator(ndb NoteDB) *noteValidator {
	return &noteValidator{
		NoteDB: ndb,
	}
}

// Search looks up a note with the given key.
func (ng *noteGorm) Search(userID uint) ([]Note, error) {
	var ret []Note
	err := Find(ng.db.Debug().Where("user_id = ?", userID), &ret)

	return ret, err
}

// ByUUID looks up a note with the given uuid.
func (ng *noteGorm) ByUUID(uuid string) (*Note, error) {
	var ret Note
	err := First(ng.db.Where("uuid = ?", uuid), &ret)

	return &ret, err
}

// ActiveByUUID looks up a note that has the given uuid and has not been deleted.
func (ng *noteGorm) ActiveByUUID(uuid string) (*Note, error) {
	var ret Note
	err := First(ng.db.Where("uuid = ? AND deleted = ?", uuid, false), &ret)

	return &ret, err
}

// TODO
func (ng *noteGorm) Delete(key string) error {
	if err := ng.db.Where("key = ?", key).Delete(&Note{}).Error; err != nil {
		return err
	}

	return nil
}

func (ng *noteGorm) Update(n *Note, tx *gorm.DB) error {
	var conn *gorm.DB
	if tx != nil {
		conn = tx
	} else {
		conn = ng.db
	}

	if err := conn.Save(n).Error; err != nil {
		return errors.Wrap(err, "saving note")
	}

	return nil
}

func (ng *noteGorm) Create(n *Note, tx *gorm.DB) error {
	var conn *gorm.DB
	if tx != nil {
		conn = tx
	} else {
		conn = ng.db
	}

	if err := conn.Save(n).Error; err != nil {
		return errors.Wrap(err, "saving note")
	}

	return nil
}

type noteValFunc func(*Note) error

func runNoteValFuncs(note *Note, fns ...noteValFunc) error {
	for _, fn := range fns {
		if err := fn(note); err != nil {
			return err
		}
	}
	return nil
}

// Create validates the parameters for retreiving a sesison by key.
func (nv *noteValidator) Create(s *Note, tx *gorm.DB) error {
	if err := runNoteValFuncs(s,
		nv.requireUserID,
		nv.requireBookUUID,
		nv.requireAddedOn,
		nv.requireEditedOn,
		nv.requireUSN,
	); err != nil {
		return err
	}

	return nv.NoteDB.Create(s, tx)
}

// Update validates the parameters for retreiving a sesison by key.
func (nv *noteValidator) Update(s *Note, tx *gorm.DB) error {
	if err := runNoteValFuncs(s,
		nv.requireUserID,
		nv.requireBookUUID,
		nv.requireAddedOn,
		nv.requireEditedOn,
		nv.requireUSN,
	); err != nil {
		return err
	}

	return nv.NoteDB.Update(s, tx)
}

// Search validates the parameters for retreiving a sesison by key.
func (nv *noteValidator) Search(userID uint) ([]Note, error) {
	s := Note{
		UserID: userID,
	}
	if err := runNoteValFuncs(&s, nv.requireUserID); err != nil {
		return nil, err
	}

	return nv.NoteDB.Search(userID)
}

// ByUUID validates the parameters for retreiving a sesison by key.
func (nv *noteValidator) ByUUID(uuid string) (*Note, error) {
	s := Note{
		UUID: uuid,
	}
	if err := runNoteValFuncs(&s, nv.requireUUID); err != nil {
		return nil, err
	}

	return nv.NoteDB.ByUUID(uuid)
}

func (nv *noteValidator) requireUUID(s *Note) error {
	if s.UUID == "" {
		return ErrNoteUUIDRequired
	}

	return nil
}

func (nv *noteValidator) requireUserID(s *Note) error {
	if s.UserID == 0 {
		return ErrNoteUserIDRequired
	}

	return nil
}

func (nv *noteValidator) requireBookUUID(s *Note) error {
	if s.BookUUID == "" {
		return ErrNoteBookUUIDRequired
	}

	return nil
}

func (nv *noteValidator) requireAddedOn(s *Note) error {
	if s.AddedOn == 0 {
		return ErrNoteAddedOnRequired
	}

	return nil
}

func (nv *noteValidator) requireEditedOn(s *Note) error {
	if s.EditedOn == 0 {
		return ErrNoteEditedOnRequired
	}

	return nil
}

func (nv *noteValidator) requireUSN(s *Note) error {
	if s.AddedOn == 0 {
		return ErrNoteUSNRequired
	}

	return nil
}

package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/permissions"
	"github.com/nadproject/nad/pkg/server/presenters"
	"github.com/nadproject/nad/pkg/server/views"
	"github.com/pkg/errors"
)

// NewNotes creates a new Notes controller.
func NewNotes(cfg config.Config, ns models.NoteService, us models.UserService, c clock.Clock, db *gorm.DB) *Notes {
	return &Notes{
		IndexView: views.NewView(cfg.PageTemplateDir, views.Config{Title: "", Layout: "base", HeaderTemplate: "navbar"}, "notes/index"),
		c:         c,
		ns:        ns,
		us:        us,
		db:        db,
	}
}

// Notes is a static controller
type Notes struct {
	IndexView *views.View
	c         clock.Clock
	ns        models.NoteService
	us        models.UserService
	db        *gorm.DB
}

// Index handles GET /
func (n *Notes) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	notes, err := n.ns.Search(user.ID)

	var vd views.Data
	vd.Yield = struct {
		Notes []models.Note
	}{
		Notes: notes,
	}

	if err != nil {
		handleHTMLError(w, err, "getting notes", &vd)
		n.IndexView.Render(w, r, vd)
		return
	}

	n.IndexView.Render(w, r, vd)
}

// NoteForm is the form data for a note
type NoteForm struct {
	BookUUID *string `schema:"book_uuid" json:"book_uuid"`
	Content  *string `schema:"content" json:"content"`
	AddedOn  *int64  `schema:"added_on" json:"added_on"`
	EditedOn *int64  `schema:"edited_on" json:"edited_on"`
}

// GetBookUUID gets the bookUUID from the NoteForm
func (r NoteForm) GetBookUUID() string {
	if r.BookUUID == nil {
		return ""
	}

	return *r.BookUUID
}

// GetContent gets the content from the NoteForm
func (r NoteForm) GetContent() string {
	if r.Content == nil {
		return ""
	}

	return *r.Content
}

// GetAddedOn gets the public field from the NoteForm
func (r NoteForm) GetAddedOn() int64 {
	if r.AddedOn == nil {
		return 0
	}

	return *r.AddedOn
}

// GetEditedOn gets the public field from the NoteForm
func (r NoteForm) GetEditedOn() int64 {
	if r.EditedOn == nil {
		return 0
	}

	return *r.EditedOn
}

func (n *Notes) create(r *http.Request) (models.Note, error) {
	var form NoteForm
	if err := parseRequestData(r, &form); err != nil {
		return models.Note{}, err
	}

	user := context.User(r.Context())
	tx := n.db.Begin()

	nextUSN, err := n.us.IncrementUSN(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.Note{}, errors.Wrap(err, "incrementing user max_usn")
	}

	now := n.c.Now().UnixNano()
	if form.AddedOn == nil {
		form.AddedOn = &now
	}
	if form.EditedOn == nil {
		form.EditedOn = &now
	}

	note := models.Note{
		UserID:   user.ID,
		USN:      nextUSN,
		BookUUID: form.GetBookUUID(),
		AddedOn:  form.GetAddedOn(),
		EditedOn: form.GetEditedOn(),
		Body:     form.GetContent(),
	}
	if err := n.ns.Create(&note, tx); err != nil {
		tx.Rollback()
		return note, errors.Wrap(err, "inserting note")
	}

	tx.Commit()

	return note, nil
}

// V1Create handles POST /api/v1/notes
func (n *Notes) V1Create(w http.ResponseWriter, r *http.Request) {
	note, err := n.create(r)
	if err != nil {
		handleJSONError(w, err, "creating note")
		return
	}

	resp := presenters.PresentNote(note)
	respondJSON(w, http.StatusCreated, resp)
}

func (n *Notes) update(r *http.Request) (models.Note, error) {
	vars := mux.Vars(r)
	noteUUID := vars["noteUUID"]

	var form NoteForm
	if err := parseRequestData(r, &form); err != nil {
		return models.Note{}, err
	}

	user := context.User(r.Context())
	tx := n.db.Begin()

	note, err := n.ns.ByUUID(noteUUID)
	if err != nil {
		return models.Note{}, errors.Wrap(err, "getting note")
	}

	// Check for permission. If not allowed, respond with not found.
	if ok := permissions.UpdateNote(user.ID, *note); !ok {
		return models.Note{}, models.ErrNotFound
	}

	nextUSN, err := n.us.IncrementUSN(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.Note{}, errors.Wrap(err, "incrementing user max_usn")
	}

	if form.BookUUID != nil {
		note.BookUUID = form.GetBookUUID()
	}
	if form.Content != nil {
		note.Body = form.GetContent()
	}
	note.USN = nextUSN
	note.EditedOn = n.c.Now().UnixNano()
	note.Deleted = false

	err = n.ns.Update(note, tx)
	if err != nil {
		tx.Rollback()
		return models.Note{}, errors.Wrap(err, "updating")
	}

	tx.Commit()

	return models.Note{}, nil
}

// V1Update handles PATCH /api/v1/notes/:uuid
func (n *Notes) V1Update(w http.ResponseWriter, r *http.Request) {
	note, err := n.update(r)
	if err != nil {
		handleJSONError(w, err, "creating note")
		return
	}

	resp := presenters.PresentNote(note)
	respondJSON(w, http.StatusOK, resp)
}

func removeNote(tx *gorm.DB, userID uint, noteUUID string, ns models.NoteService, us models.UserService) error {
	note, err := ns.ByUUID(noteUUID)
	if err != nil {
		return errors.Wrap(err, "getting note")
	}

	if ok := permissions.DeleteNote(userID, *note); !ok {
		return models.ErrNotFound
	}

	nextUSN, err := us.IncrementUSN(tx, userID)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "incrementing user max_usn")
	}

	note.USN = nextUSN
	note.Deleted = true
	note.Body = ""

	err = ns.Update(note, tx)
	if err != nil {
		return errors.Wrap(err, "updating")
	}

	return nil
}

func (n *Notes) remove(r *http.Request) (models.Note, error) {
	vars := mux.Vars(r)
	noteUUID := vars["noteUUID"]

	user := context.User(r.Context())
	tx := n.db.Begin()

	if err := removeNote(tx, user.ID, noteUUID, n.ns, n.us); err != nil {
		tx.Rollback()
		return models.Note{}, errors.Wrap(err, "removing note")
	}

	tx.Commit()

	return models.Note{}, nil
}

// V1Delete handles DELETE /api/v1/notes/:uuid
func (n *Notes) V1Delete(w http.ResponseWriter, r *http.Request) {
	note, err := n.remove(r)
	if err != nil {
		handleJSONError(w, err, "creating note")
		return
	}

	resp := presenters.PresentNote(note)
	respondJSON(w, http.StatusOK, resp)
}

func (n *Notes) get(r *http.Request) (models.Note, error) {
	user := context.User(r.Context())

	vars := mux.Vars(r)
	noteUUID := vars["noteUUID"]

	note, err := n.ns.ActiveByUUID(noteUUID)
	if err != nil {
		return models.Note{}, errors.Wrap(err, "getting note")
	}

	if ok := permissions.ViewNote(user.ID, *note); !ok {
		return models.Note{}, models.ErrNotFound
	}

	return *note, nil
}

// V1Get handles GET /api/v1/notes/:uuid
func (n *Notes) V1Get(w http.ResponseWriter, r *http.Request) {
	note, err := n.get(r)
	if err != nil {
		handleJSONError(w, err, "creating note")
		return
	}

	resp := presenters.PresentNote(note)
	respondJSON(w, http.StatusOK, resp)
}

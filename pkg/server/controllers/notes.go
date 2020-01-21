package controllers

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/presenters"
	"github.com/nadproject/nad/pkg/server/views"
	"github.com/pkg/errors"
)

// NewNotes creates a new Notes controller.
func NewNotes(ns models.NoteService, us models.UserService, db *gorm.DB) *Notes {
	return &Notes{
		IndexView: views.NewView(views.Config{Title: "", Layout: "base", HeaderTemplate: "navbar"}, "notes/index"),
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
		BookUUID: *form.BookUUID,
		AddedOn:  *form.AddedOn,
		EditedOn: *form.EditedOn,
		Body:     *form.Content,
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

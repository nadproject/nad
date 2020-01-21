package controllers

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/views"
)

// NewNotes creates a new Notes controller.
func NewNotes(ns models.NoteService) *Notes {
	return &Notes{
		IndexView: views.NewView(views.Config{Title: "", Layout: "base", HeaderTemplate: "navbar"}, "notes/index"),
		ns:        ns,
	}
}

// Notes is a static controller
type Notes struct {
	IndexView *views.View
	ns        models.NoteService
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

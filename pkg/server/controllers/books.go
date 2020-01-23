package controllers

import (
	"log"
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

// NewBooks creates a new Books controller.
func NewBooks(cfg config.Config, bs models.BookService, us models.UserService, ns models.NoteService, c clock.Clock, db *gorm.DB) *Books {
	return &Books{
		IndexView: views.NewView(cfg.PageTemplateDir, views.Config{Title: "", Layout: "base", HeaderTemplate: "navbar"}, "books/index"),
		c:         c,
		bs:        bs,
		ns:        ns,
		us:        us,
		db:        db,
	}
}

// Books is a static controller
type Books struct {
	IndexView *views.View
	c         clock.Clock
	bs        models.BookService
	ns        models.NoteService
	us        models.UserService
	db        *gorm.DB
}

// Index handles GET /
func (b *Books) Index(w http.ResponseWriter, r *http.Request) {
	// user := context.User(r.Context())
	books, err := b.bs.Search(models.BookSearchParams{})

	var vd views.Data
	vd.Yield = struct {
		Books []models.Book
	}{
		Books: books,
	}

	if err != nil {
		handleHTMLError(w, err, "getting books", &vd)
		b.IndexView.Render(w, r, vd)
		return
	}

	b.IndexView.Render(w, r, vd)
}

// BookForm is the form data for a book
type BookForm struct {
	Name *string `schema:"name" json:"name"`
}

// GetName gets the name from the BookForm
func (r BookForm) GetName() string {
	if r.Name == nil {
		return ""
	}

	return *r.Name
}

func (b *Books) create(r *http.Request) (models.Book, error) {
	var form BookForm
	if err := parseRequestData(r, &form); err != nil {
		return models.Book{}, err
	}

	user := context.User(r.Context())
	tx := b.db.Begin()

	log.Println(b.us, user.ID)

	nextUSN, err := b.us.IncrementUSN(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "incrementing user max_usn")
	}

	book := models.Book{
		UserID:    user.ID,
		Name:      form.GetName(),
		AddedOn:   b.c.Now().UnixNano(),
		USN:       nextUSN,
		Encrypted: false,
	}
	if err := b.bs.Create(&book, tx); err != nil {
		tx.Rollback()
		return book, errors.Wrapf(err, "inserting book %s", book.Name)
	}

	tx.Commit()

	return book, nil
}

// V1Create handles POST /api/v1/books
func (b *Books) V1Create(w http.ResponseWriter, r *http.Request) {
	book, err := b.create(r)
	if err != nil {
		handleJSONError(w, err, "creating book")
		return
	}

	resp := presenters.PresentBook(book)
	respondJSON(w, http.StatusCreated, resp)
}

func (b *Books) update(r *http.Request) (models.Book, error) {
	vars := mux.Vars(r)
	bookUUID := vars["bookUUID"]

	var form BookForm
	if err := parseRequestData(r, &form); err != nil {
		return models.Book{}, err
	}

	user := context.User(r.Context())
	tx := b.db.Begin()

	book, err := b.bs.ByUUID(bookUUID)
	if err != nil {
		return models.Book{}, errors.Wrap(err, "getting book")
	}

	// Check for permission. If not allowed, respond with not found.
	if ok := permissions.UpdateBook(user.ID, *book); !ok {
		return models.Book{}, models.ErrNotFound
	}

	nextUSN, err := b.us.IncrementUSN(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "incrementing user max_usn")
	}

	if form.Name != nil {
		book.Name = form.GetName()
	}

	book.USN = nextUSN
	book.EditedOn = b.c.Now().UnixNano()
	book.Deleted = false

	if err := b.bs.Update(book, tx); err != nil {
		return *book, errors.Wrap(err, "updating the book")
	}

	tx.Commit()

	return models.Book{}, nil
}

// V1Update handles PATCH /api/v1/books/:uuid
func (b *Books) V1Update(w http.ResponseWriter, r *http.Request) {
	book, err := b.update(r)
	if err != nil {
		handleJSONError(w, err, "creating book")
		return
	}

	resp := presenters.PresentBook(book)
	respondJSON(w, http.StatusOK, resp)
}

func (b *Books) remove(r *http.Request) (models.Book, error) {
	vars := mux.Vars(r)
	bookUUID := vars["bookUUID"]

	user := context.User(r.Context())
	tx := b.db.Begin()

	book, err := b.bs.ByUUID(bookUUID)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "getting book")
	}

	if ok := permissions.DeleteBook(user.ID, *book); !ok {
		tx.Rollback()
		return models.Book{}, models.ErrNotFound
	}

	notes, err := b.ns.ActiveByBookUUID(book.UUID)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "getting notes for the book")
	}

	for _, note := range notes {
		if err := removeNote(tx, user.ID, note.UUID, b.ns, b.us); err != nil {
			tx.Rollback()
			return models.Book{}, errors.Wrapf(err, "deleting note %s", note.UUID)
		}
	}

	nextUSN, err := b.us.IncrementUSN(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "incrementing user max_usn")
	}

	book.USN = nextUSN
	book.Deleted = true
	book.Name = ""

	err = b.bs.Update(book, tx)
	if err != nil {
		tx.Rollback()
		return models.Book{}, errors.Wrap(err, "updating")
	}

	tx.Commit()

	return models.Book{}, nil
}

// V1Delete handles DELETE /api/v1/books/:uuid
func (b *Books) V1Delete(w http.ResponseWriter, r *http.Request) {
	book, err := b.remove(r)
	if err != nil {
		handleJSONError(w, err, "deleting book")
		return
	}

	resp := presenters.PresentBook(book)
	respondJSON(w, http.StatusOK, resp)
}

func (b *Books) show(r *http.Request) (models.Book, error) {
	user := context.User(r.Context())

	vars := mux.Vars(r)
	bookUUID := vars["bookUUID"]

	book, err := b.bs.ByUUID(bookUUID)
	if err != nil {
		return models.Book{}, errors.Wrap(err, "getting book")
	}

	if ok := permissions.ViewBook(user.ID, *book); !ok {
		return models.Book{}, models.ErrNotFound
	}

	return *book, nil
}

// V1Show handles GET /api/v1/books/:uuid
func (b *Books) V1Show(w http.ResponseWriter, r *http.Request) {
	book, err := b.show(r)
	if err != nil {
		handleJSONError(w, err, "getting book")
		return
	}

	resp := presenters.PresentBook(book)
	respondJSON(w, http.StatusOK, resp)
}

func (b *Books) index(r *http.Request) ([]models.Book, error) {
	user := context.User(r.Context())

	books, err := b.bs.Search(models.BookSearchParams{UserID: user.ID, Offset: 0, Limit: 10})
	if err != nil {
		return nil, errors.Wrap(err, "getting books")
	}

	for _, b := range books {
		if ok := permissions.ViewBook(user.ID, b); !ok {
			return nil, models.ErrNotFound
		}
	}

	return books, nil
}

// V1Index handles GET /api/v1/books
func (b *Books) V1Index(w http.ResponseWriter, r *http.Request) {
	books, err := b.index(r)
	if err != nil {
		log.Println(err)
		handleJSONError(w, err, "getting book")
		return
	}

	resp := presenters.PresentBooks(books)
	respondJSON(w, http.StatusOK, resp)
}

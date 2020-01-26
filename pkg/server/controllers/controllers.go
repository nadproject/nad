package controllers

import (
	"github.com/nadproject/nad/pkg/clock"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

// Controllers is a group of controllers
type Controllers struct {
	Users  *Users
	Notes  *Notes
	Books  *Books
	Sync   *Sync
	Static *Static
}

// New returns a new group of controllers
func New(cfg config.Config, s *models.Services, cl clock.Clock) *Controllers {
	c := Controllers{}

	c.Users = NewUsers(cfg, s.User, s.Session)
	c.Notes = NewNotes(cfg, s.Note, s.User, cl, s.DB)
	c.Books = NewBooks(cfg, s.Book, s.User, s.Note, cl, s.DB)
	c.Sync = NewSync(s.Note, s.Book, cl)
	c.Static = NewStatic(cfg)

	return &c
}

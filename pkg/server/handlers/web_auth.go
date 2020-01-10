package handlers

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/log"
)

type joinData struct {
	Title string
}

func (c *Context) renderSignup(w http.ResponseWriter, r *http.Request) {
	dat := homeData{
		Title: getTitle("Join"),
	}

	if err := c.render(w, "user/join.html", dat); err != nil {
		log.ErrorWrap(err, "rendering")
	}
}

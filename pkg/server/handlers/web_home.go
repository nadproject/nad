package handlers

import (
	"net/http"
)

func (c *Context) renderHome(w http.ResponseWriter, r *http.Request) {
	c.App.Templates.ExecuteTemplate(w, "index.html", nil)
}

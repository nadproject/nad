package handlers

//
//import (
//	"fmt"
//	"net/http"
//
//	"github.com/nadproject/nad/pkg/server/database"
//	"github.com/nadproject/nad/pkg/server/helpers"
//	"github.com/nadproject/nad/pkg/server/log"
//)
//
//const baseTitle = "NAD"
//
//func getTitle(lead string) string {
//	if lead == "" {
//		return baseTitle
//	}
//
//	return fmt.Sprintf("%s | %s", lead, baseTitle)
//}
//
//type homeData struct {
//	Title string
//}
//
//func (c *Context) renderHome(w http.ResponseWriter, r *http.Request) {
//	_, ok := r.Context().Value(helpers.KeyUser).(database.User)
//	if !ok {
//		http.Redirect(w, r, "/register", 301)
//		return
//	}
//
//	dat := homeData{
//		Title: getTitle(""),
//	}
//
//	if err := c.render(w, "index.html", dat); err != nil {
//		log.ErrorWrap(err, "rendering home")
//	}
//}

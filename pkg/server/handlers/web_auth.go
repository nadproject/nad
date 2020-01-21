package handlers

//
// import (
// 	"net/http"
//
// 	"github.com/nadproject/nad/pkg/server/database"
// 	"github.com/nadproject/nad/pkg/server/log"
// 	"github.com/pkg/errors"
// )
//
// type joinData struct {
// 	Title string
// 	Error string
// }
//
// func (c *Context) renderRegister(w http.ResponseWriter, r *http.Request) {
// 	dat := joinData{
// 		Title: getTitle("Join"),
// 	}
//
// 	if err := c.render(w, "user/join.html", dat); err != nil {
// 		log.ErrorWrap(err, "rendering")
// 	}
// }
//
// func parseRegisterPaylaod(r *http.Request) (registerPayload, error) {
// 	if err := r.ParseForm(); err != nil {
// 		return registerPayload{}, errors.Wrap(err, "parsing form")
// 	}
//
// 	ret := registerPayload{
// 		Email:    r.PostFormValue("email"),
// 		Password: r.PostFormValue("password"),
// 	}
// 	return ret, nil
// }
//
// func (c *Context) register(w http.ResponseWriter, r *http.Request) {
// 	if c.App.Config.DisableRegistration {
// 		respondForbidden(w)
// 		return
// 	}
//
// 	params, err := parseRegisterPaylaod(r)
// 	if err != nil {
// 		HandleError(w, "parsing payload", err, http.StatusInternalServerError)
// 		return
// 	}
// 	if err := validateRegisterPayload(params); err != nil {
// 		dat := joinData{
// 			Title: getTitle("Join"),
// 			Error: err.Error(),
// 		}
//
// 		if err := c.render(w, "user/join.html", dat); err != nil {
// 			log.ErrorWrap(err, "rendering")
// 		}
// 		return
// 	}
//
// 	var count int
// 	if err := c.App.DB.Model(database.User{}).Where("email = ?", params.Email).Count(&count).Error; err != nil {
// 		HandleError(w, "checking duplicate user", err, http.StatusInternalServerError)
// 		return
// 	}
// 	if count > 0 {
// 		http.Error(w, "Duplicate email", http.StatusBadRequest)
// 		return
// 	}
//
// 	user, err := c.App.CreateUser(params.Email, params.Password)
// 	if err != nil {
// 		HandleError(w, "creating user", err, http.StatusInternalServerError)
// 		return
// 	}
//
// 	session, err := c.App.CreateSession(user.ID)
// 	if err != nil {
// 		HandleError(w, "creating session", nil, http.StatusBadRequest)
// 		return
// 	}
// 	setSessionCookie(w, session.Key, session.ExpiresAt)
//
// 	http.Redirect(w, r, "/", http.StatusCreated)
//
// 	if err := c.App.SendWelcomeEmail(params.Email); err != nil {
// 		log.ErrorWrap(err, "sending welcome email")
// 	}
// }

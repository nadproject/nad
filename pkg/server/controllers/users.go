package controllers

import (
	"net/http"
	"time"

	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/views"
	"github.com/pkg/errors"
)

// NewUsers creates a new Users controller.
// It panics if the necessary templates are not parsed.
func NewUsers(us models.UserService, ss models.SessionService) *Users {
	return &Users{
		NewView:   views.NewView(views.Config{Title: "Join", Layout: "base"}, "users/new"),
		LoginView: views.NewView(views.Config{Title: "Sign in", Layout: "base"}, "users/login"),
		us:        us,
		ss:        ss,
	}
}

// Users is a user controller.
type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPwView *views.View
	ResetPwView  *views.View
	us           models.UserService
	ss           models.SessionService
}

// New handles GET /register
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form RegistrationForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}

// RegistrationForm is the form data for registering
type RegistrationForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create handles POST /register
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	vd := views.Data{}
	var form RegistrationForm
	vd.Yield = &form
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		handleHTMLError(w, err, "creating user", &vd)
		u.NewView.Render(w, r, vd)
		return
	}

	s, err := u.signIn(&user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	setSessionCookie(w, s.Key, s.ExpiresAt)

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome",
	}
	views.RedirectAlert(w, r, "/", http.StatusFound, alert)
}

// LoginForm is the form data for log in
type LoginForm struct {
	Email    string `schema:"email" json:"email"`
	Password string `schema:"password" json:"password"`
}

func (u *Users) login(r *http.Request) (*models.Session, error) {
	form := LoginForm{}

	if err := parseRequestContent(r, &form); err != nil {
		return nil, err
	}

	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		// If the user is not found, treat it as invalid login
		if err == models.ErrNotFound {
			return nil, models.ErrLoginInvalid
		}

		return nil, err
	}

	s, err := u.signIn(user)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Login handles POST: /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	vd := views.Data{}

	s, err := u.login(r)
	if err != nil {
		handleHTMLError(w, err, "logging in user", &vd)
		u.LoginView.Render(w, r, vd)
		return
	}

	setSessionCookie(w, s.Key, s.ExpiresAt)
	http.Redirect(w, r, "/", http.StatusFound)
}

// V1Login handles POST /v1/login
func (u *Users) V1Login(w http.ResponseWriter, r *http.Request) {
	s, err := u.login(r)
	if err != nil {
		handleJSONError(w, err, "logging in user")
		return
	}

	respondWithSession(w, http.StatusOK, *s)
}

// logout deletes a users session.
func (u *Users) logout(r *http.Request) error {
	key, err := GetCredential(r)
	if err != nil {
		return errors.Wrap(err, "getting credential")
	}

	if err = u.ss.Delete(key); err != nil {
		return errors.Wrap(err, "deleting user")
	}

	return nil
}

// Logout handles POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	if err := u.logout(r); err != nil {
		logError(err, "")

		var vd views.Data
		vd.SetAlert(err)
		views.RedirectAlert(w, r, "/", http.StatusFound, *vd.Alert)
		return
	}

	unsetSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// V1Logout handles POST /v1/logout
func (u *Users) V1Logout(w http.ResponseWriter, r *http.Request) {
	if err := u.logout(r); err != nil {
		handleJSONError(w, err, "logging out")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResetPwForm is used to process the forgot password form
// and the reset password form.
type ResetPwForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

// ResetPw displays the reset password form and has a method
// so that we can prefill the form data with a token provided
// via the URL query params.
//
// GET /reset
func (u *Users) ResetPw(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form
	if err := parseURLParams(r, &form); err != nil {
		vd.SetAlert(err)
	}
	u.ResetPwView.Render(w, r, vd)
}

// signIn is used to sign the given user by creating a session
func (u *Users) signIn(user *models.User) (*models.Session, error) {
	t := time.Now()

	user.LastLoginAt = &t
	if err := u.us.Update(user); err != nil {
		return nil, errors.Wrap(err, "updating last_login_at")
	}

	s, err := u.ss.Login(user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "logging in")
	}

	return s, nil
}

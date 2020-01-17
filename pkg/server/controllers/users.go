package controllers

import (
	"net/http"
	"time"

	// "github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/views"
	"github.com/pkg/errors"
)

// NewUsers is used to create a new Users controller.
// This function will panic if the templates are not
// parsed correctly, and should only be used during
// initial setup.
func NewUsers(us models.UserService, ss models.SessionService) *Users {
	return &Users{
		NewView:      views.NewView("bootstrap", "users/new"),
		LoginView:    views.NewView("bootstrap", "users/login"),
		ForgotPwView: views.NewView("bootstrap", "users/forgot_pw"),
		ResetPwView:  views.NewView("bootstrap", "users/reset_pw"),
		us:           us,
		ss:           ss,
	}
}

// Users is a user controller
type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPwView *views.View
	ResetPwView  *views.View
	us           models.UserService
	ss           models.SessionService
}

// New is used to render the form where a user can
// create a new user account.
//
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}

// SignupForm is the form data for sign up
type SignupForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create creates a new user.
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm
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
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	if err := u.signIn(w, &user); err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome to LensLocked.com!",
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
}

// LoginForm is the form data for log in
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to verify the provided email address and
// password and then log the user in if they are correct.
//
// POST /login
// func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
// 	vd := views.Data{}
// 	form := LoginForm{}
// 	if err := parseForm(r, &form); err != nil {
// 		vd.SetAlert(err)
// 		u.LoginView.Render(w, r, vd)
// 		return
// 	}
//
// 	user, err := u.us.Authenticate(form.Email, form.Password)
// 	if err != nil {
// 		switch err {
// 		case models.ErrNotFound:
// 			vd.AlertError("Invalid email address")
// 		default:
// 			vd.SetAlert(err)
// 		}
// 		u.LoginView.Render(w, r, vd)
// 		return
// 	}
//
// 	err = u.signIn(w, user)
// 	if err != nil {
// 		vd.SetAlert(err)
// 		u.LoginView.Render(w, r, vd)
// 		return
// 	}
//
// 	http.Redirect(w, r, "/galleries", http.StatusFound)
// }

// Logout deletes a users session.
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	key, err := GetCredential(r)

	var vd views.Data
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	if err = u.ss.Delete(key); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	unsetSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)

	http.Redirect(w, r, "/", http.StatusFound)
}

// ResetPwForm is used to process the forgot password form
// and the reset password form.
type ResetPwForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

// POST /forgot
// func (u *Users) InitiateReset(w http.ResponseWriter, r *http.Request) {
// 	// TODO: Process the forgot password form and iniiate that process
// 	var vd views.Data
// 	var form ResetPwForm
// 	vd.Yield = &form
// 	if err := parseForm(r, &form); err != nil {
// 		vd.SetAlert(err)
// 		u.ForgotPwView.Render(w, r, vd)
// 		return
// 	}
//
// 	token, err := u.us.InitiateReset(form.Email)
// 	if err != nil {
// 		vd.SetAlert(err)
// 		u.ForgotPwView.Render(w, r, vd)
// 		return
// 	}
//
// 	err = u.emailer.ResetPw(form.Email, token)
// 	if err != nil {
// 		vd.SetAlert(err)
// 		u.ForgotPwView.Render(w, r, vd)
// 		return
// 	}
//
// 	views.RedirectAlert(w, r, "/reset", http.StatusFound, views.Alert{
// 		Level:   views.AlertLvlSuccess,
// 		Message: "Instructions for resetting your password have been emailed to you.",
// 	})
// }

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

// CompleteReset processed the reset password form
//
// POST /reset
// func (u *Users) CompleteReset(w http.ResponseWriter, r *http.Request) {
// 	var vd views.Data
// 	var form ResetPwForm
// 	vd.Yield = &form
// 	if err := parseForm(r, &form); err != nil {
// 		vd.SetAlert(err)
// 		u.ResetPwView.Render(w, r, vd)
// 		return
// 	}
//
// 	user, err := u.us.CompleteReset(form.Token, form.Password)
// 	if err != nil {
// 		vd.SetAlert(err)
// 		u.ResetPwView.Render(w, r, vd)
// 		return
// 	}
//
// 	u.signIn(w, user)
// 	views.RedirectAlert(w, r, "/galleries", http.StatusFound, views.Alert{
// 		Level:   views.AlertLvlSuccess,
// 		Message: "Your password has been reset and you have been logged in!",
// 	})
// }
//
// signIn is used to sign the given user in via cookies

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	t := time.Now()

	user.LastLoginAt = &t
	if err := u.us.Update(user); err != nil {
		return errors.Wrap(err, "updating last_login_at")
	}

	s, err := u.ss.Login(user.ID)
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	setSessionCookie(w, s.Key, s.ExpiresAt)

	return nil
}

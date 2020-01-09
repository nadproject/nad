/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of nad.
 *
 * nad is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * nad is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with nad.  If not, see <https://www.gnu.org/licenses/>.
 */

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// ErrLoginFailure is an error for failed login
var ErrLoginFailure = errors.New("Wrong email and password combination")

// SessionResponse is a response containing a session information
type SessionResponse struct {
	Key       string `json:"key"`
	ExpiresAt int64  `json:"expires_at"`
}

func setSessionCookie(w http.ResponseWriter, key string, expires time.Time) {
	cookie := http.Cookie{
		Name:     "id",
		Value:    key,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func unsetSessionCookie(w http.ResponseWriter) {
	expire := time.Now().Add(time.Hour * -24 * 30)
	cookie := http.Cookie{
		Name:     "id",
		Value:    "",
		Expires:  expire,
		Path:     "/",
		HttpOnly: true,
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.SetCookie(w, &cookie)
}

func touchLastLoginAt(db *gorm.DB, user database.User) error {
	t := time.Now()
	if err := db.Model(&user).Update(database.User{LastLoginAt: &t}).Error; err != nil {
		return errors.Wrap(err, "updating last_login_at")
	}

	return nil
}

type signinPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) signin(w http.ResponseWriter, r *http.Request) {
	var params signinPayload
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		HandleError(w, "decoding payload", err, http.StatusInternalServerError)
		return
	}
	if params.Email == "" || params.Password == "" {
		http.Error(w, ErrLoginFailure.Error(), http.StatusUnauthorized)
		return
	}

	var user database.User
	conn := a.App.DB.Where("email = ?", params.Email).First(&user)
	if conn.RecordNotFound() {
		http.Error(w, ErrLoginFailure.Error(), http.StatusUnauthorized)
		return
	} else if conn.Error != nil {
		HandleError(w, "getting user", err, http.StatusInternalServerError)
		return
	}

	password := []byte(params.Password)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), password)
	if err != nil {
		http.Error(w, ErrLoginFailure.Error(), http.StatusUnauthorized)
		return
	}

	err = a.App.TouchLastLoginAt(user, a.App.DB)
	if err != nil {
		http.Error(w, errors.Wrap(err, "touching login timestamp").Error(), http.StatusInternalServerError)
		return
	}

	a.respondWithSession(a.App.DB, w, user.ID, http.StatusOK)
}

func (a *API) signoutOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Version")
}

func (a *API) signout(w http.ResponseWriter, r *http.Request) {
	key, err := getCredential(r)
	if err != nil {
		HandleError(w, "getting credential", nil, http.StatusInternalServerError)
		return
	}

	if key == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = a.App.DeleteSession(key)
	if err != nil {
		HandleError(w, "deleting session", nil, http.StatusInternalServerError)
		return
	}

	unsetSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type registerPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validateRegisterPayload(p registerPayload) error {
	if p.Email == "" {
		return errors.New("email is required")
	}
	if len(p.Password) < 8 {
		return errors.New("Password should be longer than 8 characters")
	}

	return nil
}

func parseRegisterPaylaod(r *http.Request) (registerPayload, error) {
	var ret registerPayload
	if err := json.NewDecoder(r.Body).Decode(&ret); err != nil {
		return ret, errors.Wrap(err, "decoding json")
	}

	return ret, nil
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	if a.App.Config.DisableRegistration {
		respondForbidden(w)
		return
	}

	params, err := parseRegisterPaylaod(r)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if err := validateRegisterPayload(params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var count int
	if err := a.App.DB.Model(database.User{}).Where("email = ?", params.Email).Count(&count).Error; err != nil {
		HandleError(w, "checking duplicate user", err, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Duplicate email", http.StatusBadRequest)
		return
	}

	user, err := a.App.CreateUser(params.Email, params.Password)
	if err != nil {
		HandleError(w, "creating user", err, http.StatusInternalServerError)
		return
	}

	a.respondWithSession(a.App.DB, w, user.ID, http.StatusCreated)

	if err := a.App.SendWelcomeEmail(params.Email); err != nil {
		log.ErrorWrap(err, "sending welcome email")
	}
}

// respondWithSession makes a HTTP response with the session from the user with the given userID.
// It sets the HTTP-Only cookie for browser clients and also sends a JSON response for non-browser clients.
func (a *API) respondWithSession(db *gorm.DB, w http.ResponseWriter, userID int, statusCode int) {
	session, err := a.App.CreateSession(userID)
	if err != nil {
		HandleError(w, "creating session", nil, http.StatusBadRequest)
		return
	}

	setSessionCookie(w, session.Key, session.ExpiresAt)

	response := SessionResponse{
		Key:       session.Key,
		ExpiresAt: session.ExpiresAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		HandleError(w, "encoding response", err, http.StatusInternalServerError)
		return
	}
}

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

	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/helpers"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/nadproject/nad/pkg/server/mailer"
	"github.com/nadproject/nad/pkg/server/presenters"
	"github.com/nadproject/nad/pkg/server/token"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type updateProfilePayload struct {
	Email string `json:"email"`
}

// updateProfile updates user
func (c *Context) updateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(helpers.KeyUser).(database.User)
	if !ok {
		HandleError(w, "No authenticated user found", nil, http.StatusInternalServerError)
		return
	}

	var params updateProfilePayload
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, errors.Wrap(err, "invalid params").Error(), http.StatusBadRequest)
		return
	}

	// Validate
	if len(params.Email) > 60 {
		http.Error(w, "Email is too long", http.StatusBadRequest)
		return
	}

	tx := c.App.DB.Begin()
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		HandleError(w, "saving user", err, http.StatusInternalServerError)
		return
	}

	// check if email was changed
	if params.Email != user.Email {
		user.EmailVerified = false
	}
	user.Email = params.Email

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		HandleError(w, "saving account", err, http.StatusInternalServerError)
		return
	}

	tx.Commit()

	c.respondWithSession(c.App.DB, w, user.ID, http.StatusOK)
}

type updateEmailPayload struct {
	NewEmail        string `json:"new_email"`
	NewCipherKeyEnc string `json:"new_cipher_key_enc"`
	OldAuthKey      string `json:"old_auth_key"`
	NewAuthKey      string `json:"new_auth_key"`
}

func (c *Context) createVerificationToken(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(helpers.KeyUser).(database.User)
	if !ok {
		HandleError(w, "No authenticated user found", nil, http.StatusInternalServerError)
		return
	}

	if user.EmailVerified {
		http.Error(w, "Email already verified", http.StatusGone)
		return
	}
	if user.Email == "" {
		http.Error(w, "Email not set", http.StatusUnprocessableEntity)
		return
	}

	tok, err := token.Create(c.App.DB, user.ID, database.TokenTypeEmailVerification)
	if err != nil {
		HandleError(w, "saving token", err, http.StatusInternalServerError)
		return
	}

	if err := c.App.SendVerificationEmail(user.Email, tok.Value); err != nil {
		if errors.Cause(err) == mailer.ErrSMTPNotConfigured {
			respondInvalidSMTPConfig(w)
		} else {
			HandleError(w, errors.Wrap(err, "sending verification email").Error(), nil, http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
}

type verifyEmailPayload struct {
	Token string `json:"token"`
}

func (c *Context) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var params verifyEmailPayload
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		HandleError(w, "decoding payload", err, http.StatusInternalServerError)
		return
	}

	var token database.Token
	if err := c.App.DB.
		Where("value = ? AND type = ?", params.Token, database.TokenTypeEmailVerification).
		First(&token).Error; err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	if token.UsedAt != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	// Expire after ttl
	if time.Since(token.CreatedAt).Minutes() > 30 {
		http.Error(w, "This link has been expired. Please request a new link.", http.StatusGone)
		return
	}

	var user database.User
	if err := c.App.DB.Where("id = ?", token.UserID).First(&user).Error; err != nil {
		HandleError(w, "finding account", err, http.StatusInternalServerError)
		return
	}
	if user.EmailVerified {
		http.Error(w, "Already verified", http.StatusConflict)
		return
	}

	tx := c.App.DB.Begin()
	user.EmailVerified = true
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		HandleError(w, "updating email_verified", err, http.StatusInternalServerError)
		return
	}
	if err := tx.Model(&token).Update("used_at", time.Now()).Error; err != nil {
		tx.Rollback()
		HandleError(w, "updating reset token", err, http.StatusInternalServerError)
		return
	}
	tx.Commit()

	session := makeSession(user)
	respondJSON(w, http.StatusOK, session)
}

type emailPreferernceParams struct {
	InactiveReminder *bool `json:"inactive_reminder"`
	ProductUpdate    *bool `json:"product_update"`
}

func (p emailPreferernceParams) getInactiveReminder() bool {
	if p.InactiveReminder == nil {
		return false
	}

	return *p.InactiveReminder
}

func (p emailPreferernceParams) getProductUpdate() bool {
	if p.ProductUpdate == nil {
		return false
	}

	return *p.ProductUpdate
}

func (c *Context) updateEmailPreference(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(helpers.KeyUser).(database.User)
	if !ok {
		HandleError(w, "No authenticated user found", nil, http.StatusInternalServerError)
		return
	}

	var params emailPreferernceParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		HandleError(w, "decoding payload", err, http.StatusInternalServerError)
		return
	}

	var pref database.EmailPreference
	if err := c.App.DB.Where(database.EmailPreference{UserID: user.ID}).FirstOrCreate(&pref).Error; err != nil {
		HandleError(w, "finding pref", err, http.StatusInternalServerError)
		return
	}

	tx := c.App.DB.Begin()

	if params.InactiveReminder != nil {
		pref.InactiveReminder = params.getInactiveReminder()
	}
	if params.ProductUpdate != nil {
		pref.ProductUpdate = params.getProductUpdate()
	}

	if err := tx.Save(&pref).Error; err != nil {
		tx.Rollback()
		HandleError(w, "saving pref", err, http.StatusInternalServerError)
		return
	}

	token, ok := r.Context().Value(helpers.KeyToken).(database.Token)
	if ok {
		// Mark token as used if the user was authenticated by token
		if err := tx.Model(&token).Update("used_at", time.Now()).Error; err != nil {
			tx.Rollback()
			HandleError(w, "updating reset token", err, http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	respondJSON(w, http.StatusOK, pref)
}

func (c *Context) getEmailPreference(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(helpers.KeyUser).(database.User)
	if !ok {
		HandleError(w, "No authenticated user found", nil, http.StatusInternalServerError)
		return
	}

	var pref database.EmailPreference
	if err := c.App.DB.Where(database.EmailPreference{UserID: user.ID}).First(&pref).Error; err != nil {
		HandleError(w, "finding pref", err, http.StatusInternalServerError)
		return
	}

	presented := presenters.PresentEmailPreference(pref)
	respondJSON(w, http.StatusOK, presented)
}

type updatePasswordPayload struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (c *Context) updatePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(helpers.KeyUser).(database.User)
	if !ok {
		HandleError(w, "No authenticated user found", nil, http.StatusInternalServerError)
		return
	}

	var params updatePasswordPayload
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if params.OldPassword == "" || params.NewPassword == "" {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	password := []byte(params.OldPassword)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), password); err != nil {
		log.WithFields(log.Fields{
			"user_id": user.ID,
		}).Warn("invalid password update attempt")
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	if err := validatePassword(params.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(params.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, errors.Wrap(err, "hashing password").Error(), http.StatusInternalServerError)
		return
	}

	if err := c.App.DB.Model(&user).Update("password", string(hashedNewPassword)).Error; err != nil {
		http.Error(w, errors.Wrap(err, "updating password").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

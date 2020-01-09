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
	"strings"

	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func paginate(conn *gorm.DB, page int) *gorm.DB {
	limit := 30

	// Paginate
	if page > 0 {
		offset := limit * (page - 1)
		conn = conn.Offset(offset)
	}

	conn = conn.Limit(limit)

	return conn
}

func getBookIDs(books []database.Book) []int {
	ret := []int{}

	for _, book := range books {
		ret = append(ret, book.ID)
	}

	return ret
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password should be longer than 8 characters")
	}

	return nil
}

func getClientType(origin string) string {
	if strings.HasPrefix(origin, "moz-extension://") {
		return "firefox-extension"
	}

	if strings.HasPrefix(origin, "chrome-extension://") {
		return "chrome-extension"
	}

	return "web"
}

// HandleError logs the error and responds with the given status code with a generic status text
func HandleError(w http.ResponseWriter, msg string, err error, statusCode int) {
	var message string
	if err == nil {
		message = msg
	} else {
		message = errors.Wrap(err, msg).Error()
	}

	log.WithFields(log.Fields{
		"statusCode": statusCode,
	}).Error(message)

	statusText := http.StatusText(statusCode)
	http.Error(w, statusText, statusCode)
}

// respondJSON encodes the given payload into a JSON format and writes it to the given response writer
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		HandleError(w, "encoding response", err, http.StatusInternalServerError)
	}
}

// notSupported is the handler for the route that is no longer supported
func (c *Context) notSupported(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "API version is not supported. Please upgrade your client.", http.StatusGone)
	return
}

func respondForbidden(w http.ResponseWriter) {
	http.Error(w, "forbidden", http.StatusForbidden)
}

func respondUnauthorized(w http.ResponseWriter) {
	unsetSessionCookie(w)
	w.Header().Add("WWW-Authenticate", `Bearer realm="nad Pro", charset="UTF-8"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

// RespondNotFound responds with not found
func RespondNotFound(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

func respondInvalidSMTPConfig(w http.ResponseWriter) {
	http.Error(w, "SMTP is not configured", http.StatusInternalServerError)
}

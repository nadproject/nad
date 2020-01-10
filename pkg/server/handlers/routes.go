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
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/nadproject/nad/pkg/server/app"
	"github.com/nadproject/nad/pkg/server/database"
	"github.com/nadproject/nad/pkg/server/helpers"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
)

// Route represents a single route
type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	RateLimit   bool
}

type authHeader struct {
	scheme     string
	credential string
}

func parseAuthHeader(h string) (authHeader, error) {
	parts := strings.Split(h, " ")

	if len(parts) != 2 {
		return authHeader{}, errors.New("Invalid authorization header")
	}

	parsed := authHeader{
		scheme:     parts[0],
		credential: parts[1],
	}

	return parsed, nil
}

// getSessionKeyFromCookie reads and returns a session key from the cookie sent by the
// request. If no session key is found, it returns an empty string
func getSessionKeyFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("id")

	if err == http.ErrNoCookie {
		return "", nil
	} else if err != nil {
		return "", errors.Wrap(err, "reading cookie")
	}

	return c.Value, nil
}

// getSessionKeyFromAuth reads and returns a session key from the Authorization header
func getSessionKeyFromAuth(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", nil
	}

	payload, err := parseAuthHeader(h)
	if err != nil {
		return "", errors.Wrap(err, "parsing the authorization header")
	}
	if payload.scheme != "Bearer" {
		return "", errors.New("unsupported scheme")
	}

	return payload.credential, nil
}

// getCredential extracts a session key from the request from the request header. Concretely,
// it first looks at the 'Cookie' and then the 'Authorization' header. If no credential is found,
// it returns an empty string.
func getCredential(r *http.Request) (string, error) {
	ret, err := getSessionKeyFromCookie(r)
	if err != nil {
		return "", errors.Wrap(err, "getting session key from cookie")
	}
	if ret != "" {
		return ret, nil
	}

	ret, err = getSessionKeyFromAuth(r)
	if err != nil {
		return "", errors.Wrap(err, "getting session key from Authorization header")
	}

	return ret, nil
}

// AuthWithSession performs user authentication with session
func AuthWithSession(db *gorm.DB, r *http.Request) (database.User, bool, error) {
	var user database.User

	sessionKey, err := getCredential(r)
	if err != nil {
		return user, false, errors.Wrap(err, "getting credential")
	}
	if sessionKey == "" {
		return user, false, nil
	}

	var session database.Session
	conn := db.Where("key = ?", sessionKey).First(&session)

	if conn.RecordNotFound() {
		return user, false, nil
	} else if err := conn.Error; err != nil {
		return user, false, errors.Wrap(err, "finding session")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return user, false, nil
	}

	conn = db.Where("id = ?", session.UserID).First(&user)

	if conn.RecordNotFound() {
		return user, false, nil
	} else if err := conn.Error; err != nil {
		return user, false, errors.Wrap(err, "finding user from token")
	}

	return user, true, nil
}

func authWithToken(db *gorm.DB, r *http.Request, tokenType string, p *AuthMiddlewareParams) (database.User, database.Token, bool, error) {
	var user database.User
	var token database.Token

	query := r.URL.Query()
	tokenValue := query.Get("token")
	if tokenValue == "" {
		return user, token, false, nil
	}

	conn := db.Where("value = ? AND type = ?", tokenValue, tokenType).First(&token)
	if conn.RecordNotFound() {
		return user, token, false, nil
	} else if err := conn.Error; err != nil {
		return user, token, false, errors.Wrap(err, "finding token")
	}

	if token.UsedAt != nil && time.Since(*token.UsedAt).Minutes() > 10 {
		return user, token, false, nil
	}

	if err := db.Where("id = ?", token.UserID).First(&user).Error; err != nil {
		return user, token, false, errors.Wrap(err, "finding user")
	}

	return user, token, true, nil
}

// AuthMiddlewareParams is the params for the authentication middleware
type AuthMiddlewareParams struct {
	ProOnly bool
}

func auth(next http.HandlerFunc, p *AuthMiddlewareParams) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(helpers.KeyUser).(database.User)
		if !ok {
			respondUnauthorized(w)
			return
		}

		if p != nil && p.ProOnly {
			if !user.Pro {
				respondForbidden(w)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (c *Context) tokenAuth(next http.HandlerFunc, tokenType string, p *AuthMiddlewareParams) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, token, ok, err := authWithToken(c.App.DB, r, tokenType, p)
		if err != nil {
			// log the error and continue
			log.ErrorWrap(err, "authenticating with token")
		}

		ctx := r.Context()

		if ok {
			ctx = context.WithValue(ctx, helpers.KeyToken, token)
		} else {
			// If token-based auth fails, fall back to session-based auth
			user, ok, err = AuthWithSession(c.App.DB, r)
			if err != nil {
				HandleError(w, "authenticating with session", err, http.StatusInternalServerError)
				return
			}

			if !ok {
				respondUnauthorized(w)
				return
			}
		}

		if p != nil && p.ProOnly {
			if !user.Pro {
				respondForbidden(w)
				return
			}
		}

		ctx = context.WithValue(ctx, helpers.KeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow browser extensions
		if strings.HasPrefix(origin, "moz-extension://") || strings.HasPrefix(origin, "chrome-extension://") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		next.ServeHTTP(w, r)
	})
}

// logResponseWriter wraps http.ResponseWriter to expose HTTP status code for logging.
// The optional interfaces of http.ResponseWriter are lost because of the wrapping, and
// such interfaces should be implemented if needed. (i.e. http.Pusher, http.Flusher, etc.)
type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func logging(inner http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := logResponseWriter{w, http.StatusOK}
		inner.ServeHTTP(&lw, r)

		log.WithFields(log.Fields{
			"remoteAddr": lookupIP(r),
			"uri":        r.RequestURI,
			"statusCode": lw.statusCode,
			"method":     r.Method,
			"duration":   fmt.Sprintf("%dms", time.Since(start)/1000000),
			"userAgent":  r.Header.Get("User-Agent"),
		}).Info("incoming request")
	}
}

func setUser(inner http.Handler, db *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok, err := AuthWithSession(db, r)
		if err != nil {
			HandleError(w, "authenticating with session", err, http.StatusInternalServerError)
			return
		}

		var ctx context.Context
		if ok {
			ctx = context.WithValue(r.Context(), helpers.KeyUser, user)
		} else {
			ctx = r.Context()
		}
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *Context) applyMiddleware(h http.HandlerFunc, rateLimit bool) http.Handler {
	ret := h
	ret = logging(ret)
	ret = setUser(ret, c.App.DB)

	if rateLimit && os.Getenv("GO_ENV") != "TEST" {
		ret = limit(ret)
	}

	return ret
}

// Context is a web Context configuration
type Context struct {
	App *app.App
}

func (c Context) render(w http.ResponseWriter, tmpl string, data interface{}) error {
	t := c.App.Templates

	if err := t[tmpl].ExecuteTemplate(w, "base", data); err != nil {
		return errors.Wrapf(err, "executing template %s", tmpl)
	}

	return nil
}

// init sets up the application based on the configuration
func (c *Context) init() error {
	if err := c.App.Validate(); err != nil {
		return errors.Wrap(err, "validating the app parameters")
	}

	stripe.Key = os.Getenv("StripeSecretKey")

	if c.App.StripeAPIBackend != nil {
		stripe.SetBackend(stripe.APIBackend, c.App.StripeAPIBackend)
	}

	return nil
}

// NewAPI creates and returns a new router
func (c *Context) NewAPI() (*mux.Router, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "initializing app")
	}

	proOnly := AuthMiddlewareParams{ProOnly: true}

	var routes = []Route{
		// internal
		{"GET", "/health", c.checkHealth, false},
		{"GET", "/me", auth(c.getMe, nil), true},
		// {"POST", "/verification-token", auth(c.createVerificationToken, nil), true},
		// {"PATCH", "/verify-email", c.verifyEmail, true},
		// {"POST", "/reset-token", c.createResetToken, true},
		// {"PATCH", "/reset-password", c.resetPassword, true},
		// {"PATCH", "/account/profile", auth(c.updateProfile, nil), true},
		// {"PATCH", "/account/password", auth(c.updatePassword, nil), true},
		// {"GET", "/account/email-preference", c.tokenAuth(c.getEmailPreference, database.TokenTypeEmailPreference, nil), true},
		// {"PATCH", "/account/email-preference", c.tokenAuth(c.updateEmailPreference, database.TokenTypeEmailPreference, nil), true},
		// {"POST", "/subscriptions", auth(c.createSub, nil), true},
		// {"PATCH", "/subscriptions", auth(c.updateSub, nil), true},
		//		{"GET", "/subscriptions", auth(c.getSub, nil), true},
		//		{"GET", "/stripe_source", auth(c.getStripeSource, nil), true},
		//		{"PATCH", "/stripe_source", auth(c.updateStripeSource, nil), true},
		{"GET", "/notes", auth(c.getNotes, nil), false},
		{"GET", "/notes/{noteUUID}", c.getNote, true},

		{"POST", "/webhooks/stripe", c.stripeWebhook, true},

		// v1
		{"GET", "/v1/sync/fragment", cors(auth(c.GetSyncFragment, &proOnly)), false},
		{"GET", "/v1/sync/state", cors(auth(c.GetSyncState, &proOnly)), false},
		{"OPTIONS", "/v1/books", cors(c.BooksOptions), true},
		{"GET", "/v1/books", cors(auth(c.GetBooks, nil)), true},
		{"GET", "/v1/books/{bookUUID}", cors(auth(c.GetBook, nil)), true},
		{"POST", "/v1/books", cors(auth(c.CreateBook, &proOnly)), false},
		{"PATCH", "/v1/books/{bookUUID}", cors(auth(c.UpdateBook, &proOnly)), false},
		{"DELETE", "/v1/books/{bookUUID}", cors(auth(c.DeleteBook, &proOnly)), false},
		{"OPTIONS", "/v1/notes", cors(c.NotesOptions), true},
		{"POST", "/v1/notes", cors(auth(c.CreateNote, &proOnly)), false},
		{"PATCH", "/v1/notes/{noteUUID}", auth(c.UpdateNote, &proOnly), false},
		{"DELETE", "/v1/notes/{noteUUID}", auth(c.DeleteNote, &proOnly), false},
		{"POST", "/v1/signin", cors(c.signin), true},
		{"OPTIONS", "/v1/signout", cors(c.signoutOptions), true},
		{"POST", "/v1/signout", cors(c.signout), true},
		{"POST", "/v1/register", c.register, true},
	}

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		handler := route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Handler(c.applyMiddleware(handler, route.RateLimit))
	}

	return router, nil
}

// NewWeb creates and returns a new router
func (c *Context) NewWeb() (*mux.Router, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "initializing app")
	}

	var routes = []Route{
		{"GET", "/test", auth(c.renderSignup, nil), true},
		{"GET", "/join", c.renderSignup, true},
		{"GET", "/", c.renderHome, true},
	}

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		handler := route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Handler(c.applyMiddleware(handler, route.RateLimit))
	}

	return router, nil
}

package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/nadproject/nad/pkg/server/views"
	"github.com/pkg/errors"
)

const (
	contentTypeForm = "application/x-www-form-urlencoded"
	contentTypeJSON = "application/json"
)

func parseRequestData(r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")

	if ct == contentTypeForm {
		if err := parseForm(r, dst); err != nil {
			return errors.Wrap(err, "parsing form")
		}

		return nil
	}

	// default to JSON
	if err := parseJSON(r, dst); err != nil {
		return errors.Wrap(err, "parsing JSON")
	}

	return nil
}

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.PostForm, dst)
}

func parseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.Form, dst)
}

func parseValues(values url.Values, dst interface{}) error {
	dec := schema.NewDecoder()

	// Ignore CSRF token field
	dec.IgnoreUnknownKeys(true)

	if err := dec.Decode(dst, values); err != nil {
		return err
	}

	return nil
}

func parseJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(dst); err != nil {
		return err
	}

	return nil
}

// GetCredential extracts a session key from the request from the request header. Concretely,
// it first looks at the 'Cookie' and then the 'Authorization' header. If no credential is found,
// it returns an empty string.
func GetCredential(r *http.Request) (string, error) {
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

type authHeader struct {
	scheme     string
	credential string
}

const (
	sessionCookieName = "id"
	sessionCookiePath = "/"
)

func setSessionCookie(w http.ResponseWriter, key string, expires time.Time) {
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    key,
		Expires:  expires,
		Path:     sessionCookiePath,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func unsetSessionCookie(w http.ResponseWriter) {
	expires := time.Now().Add(time.Hour * -24 * 30)
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  expires,
		Path:     sessionCookiePath,
		HttpOnly: true,
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.SetCookie(w, &cookie)
}

func logError(err error, msg string) {
	// log if internal error
	if _, ok := err.(views.PublicError); !ok {
		log.ErrorWrap(err, msg)
	}
}

func getErrStatusCode(err error) int {
	rootErr := errors.Cause(err)

	switch rootErr.(type) {
	case views.BadRequestError:
		return http.StatusBadRequest
	case views.ConflictError:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// handleHTMLError writes the error to the log and sets the error message in the data.
func handleHTMLError(w http.ResponseWriter, err error, msg string, d *views.Data) {
	statusCode := getErrStatusCode(err)
	w.WriteHeader(statusCode)

	logError(err, msg)
	d.SetAlert(err)
}

// handleJSONError logs the error and responds with the given status code with a generic status text
func handleJSONError(w http.ResponseWriter, err error, msg string) {
	logError(err, msg)
	statusCode := getErrStatusCode(err)

	rootErr := errors.Cause(err)

	var respText string
	if pErr, ok := rootErr.(views.PublicError); ok {
		respText = pErr.Public()
	} else {
		respText = http.StatusText(statusCode)
	}

	http.Error(w, respText, statusCode)
}

// SessionResponse is a response containing a session information
type SessionResponse struct {
	Key       string `json:"key"`
	ExpiresAt int64  `json:"expires_at"`
}

func respondWithSession(w http.ResponseWriter, statusCode int, s models.Session) {
	response := SessionResponse{
		Key:       s.Key,
		ExpiresAt: s.ExpiresAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleJSONError(w, err, "encoding response")
		return
	}
}

// respondJSON encodes the given payload into a JSON format and writes it to the given response writer
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		handleJSONError(w, err, "encoding response")
	}
}

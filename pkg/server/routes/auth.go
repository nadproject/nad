package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/nadproject/nad/pkg/server/models"
	"github.com/pkg/errors"
)

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
func AuthWithSession(r *http.Request, ss models.SessionService, us models.UserService) (*models.User, error) {
	sessionKey, err := getCredential(r)
	if err != nil {
		return nil, errors.Wrap(err, "getting credential")
	}
	if sessionKey == "" {
		return nil, nil
	}

	session, err := ss.ByKey(sessionKey)
	if err != nil {
		return nil, err
	}

	// check if the session has been expired.
	if session.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}

	user, err := us.ByID(session.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

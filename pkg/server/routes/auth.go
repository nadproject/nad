package routes

import (
	"net/http"
	"time"

	"github.com/nadproject/nad/pkg/server/controllers"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/pkg/errors"
)

// AuthWithSession performs user authentication with session
func AuthWithSession(r *http.Request, ss models.SessionService, us models.UserService) (*models.User, error) {
	sessionKey, err := controllers.GetCredential(r)
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

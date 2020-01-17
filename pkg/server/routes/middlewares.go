package routes

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/crypt"
	"github.com/nadproject/nad/pkg/server/models"
	"github.com/pkg/errors"
)

func newCSRFMw(c config.Config) func(http.Handler) http.Handler {
	b, err := crypt.RandomBytes(32)
	if err != nil {
		panic(errors.Wrap(err, "generating CSRF token"))
	}

	csrfMw := csrf.Protect(b, csrf.Secure(c.IsProd()))

	return csrfMw
}

func applyMiddlewares(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler {
	csrfMw := newCSRFMw(c)

	ret := h
	ret = userMw(ret, s.Session, s.User)
	ret = csrfMw(ret)
	ret = loggingMw(ret)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limitMw(ret)
	}

	return ret
}

package routes

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

func applyMiddlewares(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler {
	csrfMw := csrf.Protect([]byte(c.CSRFAuthKey), csrf.Secure(c.IsProd()))

	ret := h
	ret = userMw(ret, s.Session, s.User)
	ret = csrfMw(ret)
	ret = loggingMw(ret)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limitMw(ret)
	}

	return ret
}

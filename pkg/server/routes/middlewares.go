package routes

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

type middleware func(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler

func webMw(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler {
	csrfMw := csrf.Protect([]byte(c.CSRFAuthKey), csrf.Secure(c.IsProd()))

	ret := h
	ret = userMw(ret, s.Session, s.User)
	ret = csrfMw(ret)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limitMw(ret)
	}

	return ret
}

func apiMw(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler {
	ret := h
	ret = userMw(ret, s.Session, s.User)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limitMw(ret)
	}

	return ret
}

package routes

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/models"
)

func applyMiddlewares(h http.Handler, c config.Config, s *models.Services, rateLimit bool) http.Handler {
	ret := h
	ret = logging(ret)
	ret = setUser(ret, s.Session, s.User)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limit(ret)
	}

	return ret
}

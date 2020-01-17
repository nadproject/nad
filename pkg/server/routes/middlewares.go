package routes

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/config"
)

func applyMiddlewares(h http.Handler, c config.Config, rateLimit bool) http.Handler {
	ret := h
	ret = logging(ret)
	ret = setUser(ret, c.DB)

	if rateLimit && c.AppEnv != "TEST" {
		ret = limit(ret)
	}

	return ret
}

package routes

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/nadproject/nad/pkg/server/models"
)

func setUser(inner http.Handler, ss models.SessionService, us models.UserService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := AuthWithSession(r, ss, us)
		if err != nil {
			log.ErrorWrap(err, "authenticating with session")
			inner.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}

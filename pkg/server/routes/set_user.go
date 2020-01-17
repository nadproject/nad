package routes

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/models"
)

func setUser(inner http.Handler, us models.UserService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		us.Authenticate()
		user, ok, err := AuthWithSession(db, r)
		if err != nil {
			inner.ServeHTTP(w)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}

package routes

import (
	"net/http"

	"github.com/nadproject/nad/pkg/server/context"
	"github.com/nadproject/nad/pkg/server/log"
	"github.com/nadproject/nad/pkg/server/models"
)

func userMw(inner http.Handler, ss models.SessionService, us models.UserService) http.HandlerFunc {
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

// webRequireUserMw redirects the request to the login page if user is not set
func webRequireUserMw(inner http.Handler, us models.UserService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		inner.ServeHTTP(w, r)
	})
}

// apiRequireUserMw redirects the request to the login page if user is not set
func apiRequireUserMw(inner http.Handler, us models.UserService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		inner.ServeHTTP(w, r)
	})
}

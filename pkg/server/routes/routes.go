package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nadproject/nad/pkg/server/config"
	"github.com/nadproject/nad/pkg/server/controllers"
	"github.com/nadproject/nad/pkg/server/models"
)

// Route represents a single route
type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.Handler
	RateLimit   bool
}

// New creates and returns a new router
func New(c config.Config, s *models.Services) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	staticC := controllers.NewStatic()
	// usersC := controllers.NewUsers(s.User, s.Session)

	var routes = []Route{
		{"GET", "/", staticC.Home, true},
		// {"GET", "/register", usersC.New, true},
	}

	for _, route := range routes {
		handler := route.HandlerFunc

		router.
			Handle(route.Pattern, applyMiddlewares(handler, c, s, route.RateLimit)).
			Methods(route.Method)
	}

	return router
}

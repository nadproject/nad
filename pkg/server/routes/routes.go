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
	Method    string
	Pattern   string
	Handler   http.Handler
	RateLimit bool
}

// New creates and returns a new router
func New(c config.Config, s *models.Services) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(s.User, s.Session)

	var routes = []Route{
		{"GET", "/", staticC.Home, true},
		{"GET", "/register", http.HandlerFunc(usersC.New), true},
		{"POST", "/register", http.HandlerFunc(usersC.Create), true},
	}

	for _, route := range routes {
		wrapperHandler := applyMiddlewares(route.Handler, c, s, route.RateLimit)

		router.
			Handle(route.Pattern, wrapperHandler).
			Methods(route.Method)
	}

	return router
}

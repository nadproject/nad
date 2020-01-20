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

	usersC := controllers.NewUsers(s.User, s.Session)
	notesC := controllers.NewNotes(s.Note)
	staticC := controllers.NewStatic()

	var routes = []Route{
		{"GET", "/", requireUserMw(http.HandlerFunc(notesC.Index), s.User), true},
		{"GET", "/register", http.HandlerFunc(usersC.New), true},
		{"POST", "/register", http.HandlerFunc(usersC.Create), true},
		{"POST", "/logout", http.HandlerFunc(usersC.Logout), true},
		{"GET", "/login", usersC.LoginView, true},
		{"POST", "/login", http.HandlerFunc(usersC.Login), true},
	}

	for _, route := range routes {
		wrapperHandler := applyMiddlewares(route.Handler, c, s, route.RateLimit)

		router.
			Handle(route.Pattern, wrapperHandler).
			Methods(route.Method)
	}

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static/").Handler(staticHandler)

	router.PathPrefix("/").HandlerFunc(staticC.NotFound)

	return loggingMw(router)
}

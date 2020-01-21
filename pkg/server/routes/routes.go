package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nadproject/nad/pkg/clock"
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

func registerRoutes(router *mux.Router, mw middleware, c config.Config, s *models.Services, routes []Route) {
	for _, route := range routes {
		wrappedHandler := mw(route.Handler, c, s, route.RateLimit)

		router.
			Handle(route.Pattern, wrappedHandler).
			Methods(route.Method)
	}
}

// New creates and returns a new router
func New(cfg config.Config, s *models.Services, cl clock.Clock) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	usersC := controllers.NewUsers(s.User, s.Session)
	notesC := controllers.NewNotes(s.Note, s.User, s.DB)
	staticC := controllers.NewStatic()

	var webRoutes = []Route{
		{"GET", "/", requireUserMw(http.HandlerFunc(notesC.Index), s.User), true},
		{"GET", "/register", http.HandlerFunc(usersC.New), true},
		{"POST", "/register", http.HandlerFunc(usersC.Create), true},
		{"POST", "/logout", http.HandlerFunc(usersC.Logout), true},
		{"GET", "/login", usersC.LoginView, true},
		{"POST", "/login", http.HandlerFunc(usersC.Login), true},
	}
	webRouter := router.PathPrefix("/").Subrouter()
	registerRoutes(webRouter, webMw, cfg, s, webRoutes)

	var apiRoutes = []Route{
		{"POST", "/v1/login", http.HandlerFunc(usersC.V1Login), true},
		{"POST", "/v1/logout", http.HandlerFunc(usersC.V1Logout), true},
		{"POST", "/v1/notes", requireUserMw(http.HandlerFunc(notesC.V1Create), s.User), true},
	}
	apiRouter := router.PathPrefix("/api").Subrouter()
	registerRoutes(apiRouter, apiMw, cfg, s, apiRoutes)

	// static
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static/").Handler(staticHandler)

	// catch-all
	router.PathPrefix("/").HandlerFunc(staticC.NotFound)

	return loggingMw(router)
}

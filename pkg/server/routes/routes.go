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

	usersC := controllers.NewUsers(cfg, s.User, s.Session)
	notesC := controllers.NewNotes(cfg, s.Note, s.User, cl, s.DB)
	booksC := controllers.NewBooks(cfg, s.Book, s.User, s.Note, cl, s.DB)
	syncC := controllers.NewSync(s.Note, s.Book, cl)
	staticC := controllers.NewStatic(cfg)

	var webRoutes = []Route{
		{"GET", "/", webRequireUserMw(http.HandlerFunc(notesC.Index), s.User), true},
		{"GET", "/register", http.HandlerFunc(usersC.New), true},
		{"POST", "/register", http.HandlerFunc(usersC.Create), true},
		{"POST", "/logout", http.HandlerFunc(usersC.Logout), true},
		{"GET", "/login", usersC.LoginView, true},
		{"POST", "/login", http.HandlerFunc(usersC.Login), true},
	}
	var apiRoutes = []Route{
		{"POST", "/v1/login", http.HandlerFunc(usersC.V1Login), true},
		{"POST", "/v1/logout", http.HandlerFunc(usersC.V1Logout), true},

		{"GET", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(notesC.V1Get), s.User), true},
		{"POST", "/v1/notes", apiRequireUserMw(http.HandlerFunc(notesC.V1Create), s.User), true},
		{"PATCH", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(notesC.V1Update), s.User), true},
		{"DELETE", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(notesC.V1Delete), s.User), false},

		{"GET", "/v1/books", apiRequireUserMw(http.HandlerFunc(booksC.V1Index), s.User), true},
		{"GET", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(booksC.V1Show), s.User), true},
		{"POST", "/v1/books", apiRequireUserMw(http.HandlerFunc(booksC.V1Create), s.User), true},
		{"PATCH", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(booksC.V1Update), s.User), true},
		{"DELETE", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(booksC.V1Delete), s.User), false},

		{"GET", "/v1/sync/state", apiRequireUserMw(http.HandlerFunc(syncC.GetState), s.User), false},
		{"GET", "/v1/sync/fragment", apiRequireUserMw(http.HandlerFunc(syncC.GetFragment), s.User), false},
	}

	webRouter := router.PathPrefix("/").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	registerRoutes(webRouter, webMw, cfg, s, webRoutes)
	registerRoutes(apiRouter, apiMw, cfg, s, apiRoutes)

	// static
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static/").Handler(staticHandler)

	// catch-all
	router.PathPrefix("/").HandlerFunc(staticC.NotFound)

	return loggingMw(router)
}

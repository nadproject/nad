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

// NewWebRoutes returns a new web routes
func NewWebRoutes(cfg config.Config, c *controllers.Controllers, s *models.Services, cl clock.Clock) []Route {
	return []Route{
		{"GET", "/", webRequireUserMw(http.HandlerFunc(c.Notes.Index), s.User), true},
		{"GET", "/register", http.HandlerFunc(c.Users.New), true},
		{"POST", "/register", http.HandlerFunc(c.Users.Create), true},
		{"POST", "/logout", http.HandlerFunc(c.Users.Logout), true},
		{"GET", "/login", c.Users.LoginView, true},
		{"POST", "/login", http.HandlerFunc(c.Users.Login), true},

		{"GET", "/new", c.Notes.NewView, true},
		{"POST", "/notes", http.HandlerFunc(c.Notes.Create), true},
	}
}

// NewAPIRoutes returns a new api routes
func NewAPIRoutes(cfg config.Config, c *controllers.Controllers, s *models.Services, cl clock.Clock) []Route {
	return []Route{
		{"POST", "/v1/login", http.HandlerFunc(c.Users.V1Login), true},
		{"POST", "/v1/logout", http.HandlerFunc(c.Users.V1Logout), true},

		{"GET", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(c.Notes.V1Get), s.User), true},
		{"POST", "/v1/notes", apiRequireUserMw(http.HandlerFunc(c.Notes.V1Create), s.User), true},
		{"PATCH", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(c.Notes.V1Update), s.User), true},
		{"DELETE", "/v1/notes/{noteUUID}", apiRequireUserMw(http.HandlerFunc(c.Notes.V1Delete), s.User), false},

		{"GET", "/v1/books", apiRequireUserMw(http.HandlerFunc(c.Books.V1Index), s.User), true},
		{"GET", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(c.Books.V1Show), s.User), true},
		{"POST", "/v1/books", apiRequireUserMw(http.HandlerFunc(c.Books.V1Create), s.User), true},
		{"PATCH", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(c.Books.V1Update), s.User), true},
		{"DELETE", "/v1/books/{bookUUID}", apiRequireUserMw(http.HandlerFunc(c.Books.V1Delete), s.User), false},

		{"GET", "/v1/sync/state", apiRequireUserMw(http.HandlerFunc(c.Sync.GetState), s.User), false},
		{"GET", "/v1/sync/fragment", apiRequireUserMw(http.HandlerFunc(c.Sync.GetFragment), s.User), false},
	}
}

// RouteConfig is the configuration for routes
type RouteConfig struct {
	Controllers *controllers.Controllers
	WebRoutes   []Route
	APIRoutes   []Route
}

// New creates and returns a new router
func New(cfg config.Config, s *models.Services, rc RouteConfig) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	webRouter := router.PathPrefix("/").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	registerRoutes(webRouter, WebMw, cfg, s, rc.WebRoutes)
	registerRoutes(apiRouter, APIMw, cfg, s, rc.APIRoutes)

	// static
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir)))
	router.PathPrefix("/static/").Handler(staticHandler)

	// catch-all
	router.PathPrefix("/").HandlerFunc(rc.Controllers.Static.NotFound)

	return router
}

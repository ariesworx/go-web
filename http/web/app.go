package web

import (
	"context"
	"net/http"
	"os"
)

type App struct {
	*http.ServeMux
	middleware map[MwKey]MwHandler
	routes     []*route
	shutdown   chan os.Signal
	sorted     bool
}

// NewApp creates a new App instance with the given configuration.
func NewApp(shutdown chan os.Signal) *App {
	return &App{
		ServeMux: http.NewServeMux(),
		shutdown: shutdown,
		routes:   make([]*route, 0),
	}
}

// Handle registers a new handler function for the given method, pattern, handler, and
// optional middleware keys.
func (a *App) Handle(method, path string, handler Handler, middlewareKeys ...MwKey) {
	// Create the middleware chain based on the provided keys.
	mw := make([]MwHandler, 0, len(middlewareKeys))
	for _, key := range middlewareKeys {
		if m, ok := a.middleware[key]; ok && m != nil {
			mw = append(mw, m)
		}
	}

	h := chainMiddleware(mw, handler)
	r := newRoute(method, path, h)

	a.routes = append(a.routes, r)
	a.sorted = false // Mark routes as unsorted.
	pattern := method + " " + path

	a.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Check if routes are sorted; if not, sort them.
		if !a.sorted {
			a.sortRoutes()
			a.sorted = true
		}

		// Find the matching route.
		for _, rte := range a.routes {
			if rte.method != "" && rte.method != r.Method {
				continue
			}

			vars, matches := matchRoute(rte.segments, r.URL.Path)
			if matches {
				// Store the path variables in the request context.
				ctx := context.WithValue(r.Context(), pathVarsKey, vars)
				ctx = context.WithValue(ctx, patternKey, pattern)
				r = r.WithContext(ctx)

				// Call the handler.
				if err := rte.handler(ctx, w, r); err != nil {
					http.Error(w, "Internal server error.", http.StatusInternalServerError)
				}

				return
			}
		}

		http.NotFound(w, r)
	})
}

// buildMiddleware constructs the middleware chain for a given handler and keys.
func (a *App) buildMiddleware(handler Handler, keys []MwKey) Handler {
	mw := make([]MwHandler, 0, len(keys))
	for _, key := range keys {
		if m, ok := a.middleware[key]; ok && m != nil {
			mw = append(mw, m)
		}
	}

	return chainMiddleware(mw, handler)
}

func (a *App) sortRoutes() {
	// Sort using a stable sort to preserve registration order for equal specificity
	for i := 0; i < len(a.routes); i++ {
		for j := i + 1; j < len(a.routes); j++ {
			if routeScore(a.routes[j]) > routeScore(a.routes[i]) {
				a.routes[i], a.routes[j] = a.routes[j], a.routes[i]
			}
		}
	}
}

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	success = "✓"
	failure = "✗"
)

func TestNewApp(t *testing.T) {
	shutdown := make(chan os.Signal, 1)
	app := NewApp(shutdown)

	t.Run("creates non-nil instance", func(t *testing.T) {
		if app == nil {
			t.Fatalf("\t%s\tExpected non-nil App instance", failure)
		}
		t.Logf("\t%s\tApp instance created successfully", success)
	})

	t.Run("sets shutdown channel", func(t *testing.T) {
		if app.shutdown != shutdown {
			t.Errorf("\t%s\tExpected shutdown channel to be set correctly", failure)
		} else {
			t.Logf("\t%s\tShutdown channel set correctly", success)
		}
	})

	t.Run("initializes middleware map", func(t *testing.T) {
		if app.middleware == nil {
			t.Errorf("\t%s\tExpected middleware map to be initialized", failure)
		} else {
			t.Logf("\t%s\tMiddleware map initialized", success)
		}
	})

	t.Run("initializes empty routes slice", func(t *testing.T) {
		if len(app.routes) != 0 {
			t.Errorf("\t%s\tExpected routes slice to be initialized empty, got %d routes", failure, len(app.routes))
		} else {
			t.Logf("\t%s\tRoutes slice initialized empty", success)
		}
	})
}

func TestApp_Handle(t *testing.T) {
	t.Run("registers route correctly", func(t *testing.T) {
		shutdown := make(chan os.Signal, 1)
		app := NewApp(shutdown)

		handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return nil
		}

		path := "/test"
		app.Handle("GET", path, handler)

		if len(app.routes) != 1 {
			t.Fatalf("\t%s\tExpected 1 route to be registered, got %d", failure, len(app.routes))
		}
		t.Logf("\t%s\tRoute registered successfully", success)

		route := app.routes[0]
		if route.method != http.MethodGet || route.pattern != path || route.handler == nil {
			t.Errorf("\t%s\tRegistered route does not match expected values", failure)
		} else {
			t.Logf("\t%s\tRegistered route matches expected values", success)
		}
	})

	t.Run("handles successful request", func(t *testing.T) {
		shutdown := make(chan os.Signal, 1)
		app := NewApp(shutdown)

		expectedStatus := struct {
			Status string `json:"status"`
		}{
			Status: "OK",
		}

		handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(w).Encode(expectedStatus)
		}

		path := "/test"
		app.Handle("GET", path, handler)

		server := httptest.NewServer(app)
		t.Cleanup(server.Close)

		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("\t%s\tFailed to make GET request: %v", failure, err)
		}
		t.Cleanup(func() {
			if err = resp.Body.Close(); err != nil {
				t.Fatalf("\t%s\tFailed to close response body: %v", failure, err)
			}
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("\t%s\tExpected status code 200, got %d", failure, resp.StatusCode)
		} else {
			t.Logf("\t%s\tReceived expected status code 200", success)
		}

		var respStatus struct {
			Status string `json:"status"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&respStatus); err != nil {
			t.Fatalf("\t%s\tFailed to decode response body: %v", failure, err)
		}

		if respStatus.Status != expectedStatus.Status {
			t.Errorf("\t%s\tExpected response status %q, got %q", failure, expectedStatus.Status, respStatus.Status)
		} else {
			t.Logf("\t%s\tReceived expected response status %q", success, respStatus.Status)
		}
	})

	t.Run("returns 404 for nonexistent route", func(t *testing.T) {
		shutdown := make(chan os.Signal, 1)
		app := NewApp(shutdown)

		server := httptest.NewServer(app)
		t.Cleanup(server.Close)

		resp, err := http.Get(server.URL + "/nonexistent")
		if err != nil {
			t.Fatalf("\t%s\tFailed to make GET request to nonexistent route: %v", failure, err)
		}
		t.Cleanup(func() {
			if err = resp.Body.Close(); err != nil {
				t.Fatalf("\t%s\tFailed to close response body: %v", failure, err)
			}
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("\t%s\tExpected status code 404 for nonexistent route, got %d", failure, resp.StatusCode)
		} else {
			t.Logf("\t%s\tReceived expected status code 404 for nonexistent route", success)
		}
	})
}

// Optional: Table-driven test example with success/failure format
func TestApp_Handle_MultipleRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET route", http.MethodGet, "/users", http.StatusOK},
		{"POST route", http.MethodPost, "/users", http.StatusCreated},
		{"PUT route", http.MethodPut, "/users/1", http.StatusOK},
		{"DELETE route", http.MethodDelete, "/users/1", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown := make(chan os.Signal, 1)
			app := NewApp(shutdown)

			handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(tt.expectedStatus)
				return nil
			}

			app.Handle(tt.method, tt.path, handler)

			server := httptest.NewServer(app)
			t.Cleanup(server.Close)

			req, err := http.NewRequest(tt.method, server.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("\t%s\tFailed to create request: %v", failure, err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\tRequest failed: %v", failure, err)
			}

			t.Cleanup(func() {
				if err = resp.Body.Close(); err != nil {
					t.Fatalf("\t%s\tFailed to close response body: %v", failure, err)
				}
			})

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("\t%s\tExpected status %d, got %d", failure, tt.expectedStatus, resp.StatusCode)
			} else {
				t.Logf("\t%s\tReceived expected status %d", success, resp.StatusCode)
			}
		})
	}
}

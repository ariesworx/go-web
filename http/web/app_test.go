package web

import (
	"os"
	"testing"
)

const (
	success = "✓"
	failure = "✗"
)

func TestNewApp(t *testing.T) {
	t.Log("Test NewApp function")
	shutdown := make(chan os.Signal, 1)
	app := NewApp(shutdown)

	if app == nil {
		t.Fatalf("%s\tExpected non-nil App instance", failure)
	}
	t.Logf("%s\tApp instance created successfully", success)

	if app.shutdown != shutdown {
		t.Errorf("%s\tExpected shutdown channel to be set correctly", failure)
	} else {
		t.Logf("%s\tShutdown channel set correctly", success)
	}

	if len(app.routes) != 0 {
		t.Errorf("%s\tExpected routes slice to be initialized empty", failure)
	} else {
		t.Logf("%s\tRoutes slice initialized empty", success)
	}
}

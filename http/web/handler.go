package web

import (
	"context"
	"net/http"
)

// Handler defines the signature for HTTP request handlers within the mux package.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

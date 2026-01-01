package web

// MwKey is used to select middleware.
type MwKey string

// MwHandler defines the signature for middleware handlers.
type MwHandler func(next Handler) Handler

// chainMiddleware applies a list of middleware to a handler in the order they are provided.
func chainMiddleware(mw []MwHandler, handler Handler) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}

	return handler
}

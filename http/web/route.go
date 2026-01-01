package web

import "strings"

type route struct {
	method   string
	pattern  string
	handler  Handler
	segments []segment
}

type segment struct {
	isParam bool
	name    string
	value   string
}

// Context key for path variables
type contextKey string

const (
	pathVarsKey contextKey = "path_vars"
	patternKey  contextKey = "pattern"
)

// newRoute creates a new route instance by parsing the given pattern.
func newRoute(method, pattern string, handler Handler) *route {
	return &route{
		method:   method,
		pattern:  pattern,
		handler:  handler,
		segments: parsePattern(pattern),
	}
}

// matchRoute checks if the given path matches the route segments. It returns a
// map of path variables if there's a match.
func matchRoute(segments []segment, path string) (map[string]string, bool) {
	parts := strings.Split(path, "/")
	params := make(map[string]string)

	if len(segments) != len(parts)-1 {
		return nil, false
	}

	for i, seg := range segments {
		if seg.isParam {
			params[seg.name] = parts[i]
		} else {
			if seg.value != parts[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// parsePattern splits the path pattern into segments, identifying static and
// variable parts.
func parsePattern(path string) []segment {
	parts := strings.Split(path, "/")
	segments := make([]segment, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// Variable segment.
			name := strings.Trim(part, "{}")
			segments = append(segments, segment{
				isParam: true,
				name:    name,
			})
		} else {
			// Static segment.
			segments = append(segments, segment{
				isParam: false,
				value:   part,
			})
		}
	}

	return segments
}

// pattern returns the full pattern string for the route, including method if
// specified.
func pattern(route *route) string {
	if route.method == "" {
		return route.pattern
	}

	return route.method + " " + route.pattern
}

// routeScore calculates a specificity score for the route based on its segments
// and method.
func routeScore(route *route) int {
	score := 0

	// Each static segment adds 1000 points Each variable segment adds 1 point Longer
	// paths are more specific.
	for _, seg := range route.segments {
		if seg.isParam {
			score += 1
		} else {
			score += 1000
		}
	}

	// Routes with methods are more specific than those without.
	if route.method != "" {
		score += 10000
	}

	return score
}

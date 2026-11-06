package http

import "net/http"

const (
	corsAllowMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Content-Type"
	corsMaxAge       = "600"
)

// CORS builds a middleware that applies CORS headers for the allowed origins.
// It answers preflight OPTIONS requests and echoes the origin when it is allowed.
// State-changing requests from a disallowed browser origin are rejected.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	isAllowed := func(origin string) bool {
		_, ok := allowed[origin]
		return ok
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin != "" && !isAllowed(origin) && isUnsafeMethod(r.Method) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if origin != "" && isAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
					w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
					w.Header().Set("Access-Control-Max-Age", corsMaxAge)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
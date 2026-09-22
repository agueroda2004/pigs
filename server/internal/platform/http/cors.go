package http

import (
	"net/http"
	"strings"
)

const (
	corsAllowMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Content-Type"
	corsMaxAge       = "600"
)

// CORS builds a middleware that applies CORS headers for the allowed origins.
// It answers preflight OPTIONS requests and echoes the origin when it is allowed.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if _, ok := allowed[origin]; origin != "" && ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)

					headers := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))
					if headers == "" {
						headers = corsAllowHeaders
					}
					w.Header().Set("Access-Control-Allow-Headers", headers)
					w.Header().Set("Access-Control-Max-Age", corsMaxAge)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

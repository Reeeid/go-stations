package middleware

import (
	"net/http"
	"os"
)

func Authorization(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != os.Getenv("BASIC_AUTH_USER_ID") || pass != os.Getenv("BASIC_AUTH_PASSWORD") {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

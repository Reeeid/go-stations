package middleware

import (
	"context"
	"net/http"

	"github.com/mileusna/useragent"
)

type ctxKey string

const osKey ctxKey = "OS"

func UsarAgent(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		s := r.UserAgent()
		// ua のパース
		ua := useragent.Parse(s)
		// context に User-Agent 情報をセット
		println(ua.OS)
		OSdata := context.WithValue(r.Context(), osKey, ua.OS)
		r = r.WithContext(OSdata)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

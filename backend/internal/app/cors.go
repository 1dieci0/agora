package app

import "net/http"

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Access-Control-Allow-Origin",
			"http://localhost:5173",
		)

		w.Header().Set(
			"Access-Control-Allow-Credentials",
			"true",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

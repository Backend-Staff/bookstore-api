package middleware

import (
	"log"
	"net/http"
	"time"
)

// requestCount tracks how many requests the API has served since startup.
var requestCount int

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestCount++
		count := requestCount

		next.ServeHTTP(w, r)

		log.Printf("[%d] %s %s took %s", count, r.Method, r.URL.Path, time.Since(start))
	})
}

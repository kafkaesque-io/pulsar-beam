package middleware

//middleware includes auth, rate limit, and etc.
import (
	"log"
	"net/http"
)

// Rate is the default global rate limit
var Rate = NewSema(500)

// Authenticate middleware function
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("topicKey")

		// UUID as the topic key length validation
		if len(key) != 36 {
			log.Printf("Authenticated \n")
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

// LimitRate rate limites against http handler
// use semaphore as a simple rate limiter
// TODO: should be added to inidividual endpoint
func LimitRate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := Rate.Acquire()
		if err != nil {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		} else {
			next.ServeHTTP(w, r)
		}
		Rate.Release()
	})
}

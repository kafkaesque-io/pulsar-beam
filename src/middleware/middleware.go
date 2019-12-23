package middleware

//middleware includes auth, rate limit, and etc.
import (
	"log"
	"net/http"
	"strings"

	"github.com/pulsar-beam/src/util"
)

// Rate is the default global rate limit
var Rate = NewSema(200)

// Authenticate middleware function
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
		_, err := util.JWTAuth.DecodeToken(tokenStr)

		// key := r.Header.Get("apikey")
		if err == nil {
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

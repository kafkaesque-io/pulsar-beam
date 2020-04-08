package middleware

//middleware includes auth, rate limit, and etc.
import (
	"net/http"
	"strings"

	"github.com/kafkaesque-io/pulsar-beam/src/util"

	log "github.com/sirupsen/logrus"
)

var (
	// Rate is the default global rate limit
	// This rate only limits the rate hitting on endpoint
	// It does not limit the underline resource access
	Rate = NewSema(200)
)

// AuthFunc is a function type to allow pluggable authentication middleware
type AuthFunc func(next http.Handler) http.Handler

// AuthVerifyJWT Authenticate middleware function
func AuthVerifyJWT(next http.Handler) http.Handler {
	switch util.GetConfig().HTTPAuthImpl {
	case "noauth":
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("injectedSubs", util.SuperRoles[0])
			next.ServeHTTP(w, r)
		})
	default:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
			subjects, err := util.JWTAuth.GetTokenSubject(tokenStr)

			if err == nil {
				log.Infof("Authenticated with subjects %s", subjects)
				r.Header.Set("injectedSubs", subjects)
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}

		})
	}
}

// AuthHeaderRequired is a very weak auth to verify token existence only.
func AuthHeaderRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))

		if len(tokenStr) > 1 {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

	})
}

// NoAuth bypasses the auth middleware
func NoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// LimitRate rate limites against http handler
// use semaphore as a simple rate limiter
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

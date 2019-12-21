package route

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pulsar-beam/src/middleware"
	"github.com/pulsar-beam/src/util"
)

// NewRouter - create new router for HTTP routing
func NewRouter(mode *string) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range GetEffectiveRoutes(mode) {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}
	router.Use(middleware.LimitRate)
	router.Use(middleware.Authenticate)
	// router.Use(auth.AuthenticateAuth0) // JWT based authentication

	log.Printf("router added\n")
	return router
}

// GetEffectiveRoutes gets effective routes
func GetEffectiveRoutes(mode *string) Routes {
	switch *mode {
	case util.Hybrid:
		return append(receiverRoutes, restRoutes...)
	case util.Receiver:
		return receiverRoutes
	default:
		return restRoutes
	}
}
